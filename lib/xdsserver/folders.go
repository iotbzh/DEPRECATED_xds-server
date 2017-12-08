/*
 * Copyright (C) 2017 "IoT.bzh"
 * Author Sebastien Douheret <sebastien@iot.bzh>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package xdsserver

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/franciscocpg/reflectme"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
	"github.com/syncthing/syncthing/lib/sync"
)

// Folders Represent a an XDS folders
type Folders struct {
	*Context
	fileOnDisk string
	folders    map[string]*IFOLDER
	registerCB []RegisteredCB
}

// RegisteredCB Hold registered callbacks
type RegisteredCB struct {
	cb   *FolderEventCB
	data *FolderEventCBData
}

// Mutex to make add/delete atomic
var fcMutex = sync.NewMutex()
var ffMutex = sync.NewMutex()

// FoldersNew Create a new instance of Model Folders
func FoldersNew(ctx *Context) *Folders {
	file, _ := xdsconfig.FoldersConfigFilenameGet()
	return &Folders{
		Context:    ctx,
		fileOnDisk: file,
		folders:    make(map[string]*IFOLDER),
		registerCB: []RegisteredCB{},
	}
}

// LoadConfig Load folders configuration from disk
func (f *Folders) LoadConfig() error {
	var flds []xsapiv1.FolderConfig
	var stFlds []xsapiv1.FolderConfig

	// load from disk
	if f.Config.Options.NoFolderConfig {
		f.Log.Infof("Don't read folder config file (-no-folderconfig option is set)")
	} else if f.fileOnDisk != "" {
		f.Log.Infof("Use folder config file: %s", f.fileOnDisk)
		err := foldersConfigRead(f.fileOnDisk, &flds)
		if err != nil {
			if strings.HasPrefix(err.Error(), "No folder config") {
				f.Log.Warnf(err.Error())
			} else {
				return err
			}
		}
	} else {
		f.Log.Warnf("Folders config filename not set")
	}

	// Retrieve initial Syncthing config (just append don't overwrite existing ones)
	if f.SThg != nil {
		f.Log.Infof("Retrieve syncthing folder config")
		if err := f.SThg.FolderLoadFromStConfig(&stFlds); err != nil {
			// Don't exit on such error, just log it
			f.Log.Errorf(err.Error())
		}
	} else {
		f.Log.Infof("Syncthing support is disabled.")
	}

	// Merge syncthing folders into XDS folders
	for _, stf := range stFlds {
		found := false
		for i, xf := range flds {
			if xf.ID == stf.ID {
				found = true
				// sanity check
				if xf.Type != xsapiv1.TypeCloudSync {
					flds[i].Status = xsapiv1.StatusErrorConfig
				}
				break
			}
		}
		// add it
		if !found {
			flds = append(flds, stf)
		}
	}

	// Detect ghost project
	// (IOW existing in xds file config and not in syncthing database)
	if f.SThg != nil {
		for i, xf := range flds {
			// only for syncthing project
			if xf.Type != xsapiv1.TypeCloudSync {
				continue
			}
			found := false
			for _, stf := range stFlds {
				if stf.ID == xf.ID {
					found = true
					break
				}
			}
			if !found {
				flds[i].Status = xsapiv1.StatusErrorConfig
			}
		}
	}

	// Update folders
	f.Log.Infof("Loading initial folders config: %d folders found", len(flds))
	for _, fc := range flds {
		if _, err := f.createUpdate(fc, false, true); err != nil {
			return err
		}
	}

	// Save config on disk
	err := f.SaveConfig()

	return err
}

// SaveConfig Save folders configuration to disk
func (f *Folders) SaveConfig() error {
	if f.fileOnDisk == "" {
		return fmt.Errorf("Folders config filename not set")
	}

	// FIXME: buffered save or avoid to write on disk each time
	return foldersConfigWrite(f.fileOnDisk, f.getConfigArrUnsafe())
}

// ResolveID Complete a Folder ID (helper for user that can use partial ID value)
func (f *Folders) ResolveID(id string) (string, error) {
	if id == "" {
		return "", nil
	}

	match := []string{}
	for iid := range f.folders {
		if strings.HasPrefix(iid, id) {
			match = append(match, iid)
		}
	}

	if len(match) == 1 {
		return match[0], nil
	} else if len(match) == 0 {
		return id, fmt.Errorf("Unknown id")
	}
	return id, fmt.Errorf("Multiple IDs found with provided prefix: " + id)
}

// Get returns the folder config or nil if not existing
func (f *Folders) Get(id string) *IFOLDER {
	if id == "" {
		return nil
	}
	fc, exist := f.folders[id]
	if !exist {
		return nil
	}
	return fc
}

// GetConfigArr returns the config of all folders as an array
func (f *Folders) GetConfigArr() []xsapiv1.FolderConfig {
	fcMutex.Lock()
	defer fcMutex.Unlock()

	return f.getConfigArrUnsafe()
}

// getConfigArrUnsafe Same as GetConfigArr without mutex protection
func (f *Folders) getConfigArrUnsafe() []xsapiv1.FolderConfig {
	conf := []xsapiv1.FolderConfig{}
	for _, v := range f.folders {
		conf = append(conf, (*v).GetConfig())
	}
	return conf
}

// Add adds a new folder
func (f *Folders) Add(newF xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error) {
	return f.createUpdate(newF, true, false)
}

// CreateUpdate creates or update a folder
func (f *Folders) createUpdate(newF xsapiv1.FolderConfig, create bool, initial bool) (*xsapiv1.FolderConfig, error) {

	fcMutex.Lock()
	defer fcMutex.Unlock()

	// Sanity check
	if _, exist := f.folders[newF.ID]; create && exist {
		return nil, fmt.Errorf("ID already exists")
	}
	if newF.ClientPath == "" {
		return nil, fmt.Errorf("ClientPath must be set")
	}

	// Create a new folder object
	var fld IFOLDER
	switch newF.Type {
	// SYNCTHING
	case xsapiv1.TypeCloudSync:
		if f.SThg != nil {
			fld = NewFolderST(f.Context, f.SThg)
		} else {
			f.Log.Debugf("Disable project %v (syncthing not initialized)", newF.ID)
			fld = NewFolderSTDisable(f.Context)
		}

	// PATH MAP
	case xsapiv1.TypePathMap:
		fld = NewFolderPathMap(f.Context)
	default:
		return nil, fmt.Errorf("Unsupported folder type")
	}

	// Allocate a new UUID
	if create {
		newF.ID = fld.NewUID("")
	}
	if !create && newF.ID == "" {
		return nil, fmt.Errorf("Cannot update folder with null ID")
	}

	// Set default value if needed
	if newF.Status == "" {
		newF.Status = xsapiv1.StatusDisable
	}
	if newF.Label == "" {
		newF.Label = filepath.Base(newF.ClientPath)
		if len(newF.ID) > 8 {
			newF.Label += "_" + newF.ID[0:8]
		}
	}

	// Normalize path (needed for Windows path including bashlashes)
	newF.ClientPath = common.PathNormalize(newF.ClientPath)

	// Add new folder
	newFolder, err := fld.Add(newF)
	if err != nil {
		newF.Status = xsapiv1.StatusErrorConfig
		log.Printf("ERROR Adding folder: %v\n", err)
		return newFolder, err
	}

	// Add to folders list
	f.folders[newF.ID] = &fld

	// Save config on disk
	if !initial {
		if err := f.SaveConfig(); err != nil {
			return newFolder, err
		}
	}

	// Register event change callback
	for _, rcb := range f.registerCB {
		if err := fld.RegisterEventChange(rcb.cb, rcb.data); err != nil {
			return newFolder, err
		}
	}

	// Force sync after creation
	// (need to defer to be sure that WS events will arrive after HTTP creation reply)
	go func() {
		time.Sleep(time.Millisecond * 500)
		fld.Sync()
	}()

	return newFolder, nil
}

// Delete deletes a specific folder
func (f *Folders) Delete(id string) (xsapiv1.FolderConfig, error) {
	var err error

	fcMutex.Lock()
	defer fcMutex.Unlock()

	fld := xsapiv1.FolderConfig{}
	fc, exist := f.folders[id]
	if !exist {
		return fld, fmt.Errorf("unknown id")
	}

	fld = (*fc).GetConfig()

	if err = (*fc).Remove(); err != nil {
		return fld, err
	}

	delete(f.folders, id)

	// Save config on disk
	err = f.SaveConfig()

	return fld, err
}

// Update Update a specific folder
func (f *Folders) Update(id string, cfg xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error) {
	fcMutex.Lock()
	defer fcMutex.Unlock()

	fc, exist := f.folders[id]
	if !exist {
		return nil, fmt.Errorf("unknown id")
	}

	// Copy current in a new object to change nothing in case of an error rises
	newCfg := xsapiv1.FolderConfig{}
	reflectme.Copy((*fc).GetConfig(), &newCfg)

	// Only update some fields
	dirty := false
	for _, fieldName := range xsapiv1.FolderConfigUpdatableFields {
		valNew, err := reflectme.GetField(cfg, fieldName)
		if err == nil {
			valCur, err := reflectme.GetField(newCfg, fieldName)
			if err == nil && valNew != valCur {
				err = reflectme.SetField(&newCfg, fieldName, valNew)
				if err != nil {
					return nil, err
				}
				dirty = true
			}
		}
	}

	if !dirty {
		return &newCfg, nil
	}

	fld, err := (*fc).Update(newCfg)
	if err != nil {
		return fld, err
	}

	// Save config on disk
	err = f.SaveConfig()

	// Send event to notified changes
	// TODO emit folder change event

	return fld, err
}

// RegisterEventChange requests registration for folder event change
func (f *Folders) RegisterEventChange(id string, cb *FolderEventCB, data *FolderEventCBData) error {

	flds := make(map[string]*IFOLDER)
	if id != "" {
		// Register to a specific folder
		flds[id] = f.Get(id)
	} else {
		// Register to all folders
		flds = f.folders
		f.registerCB = append(f.registerCB, RegisteredCB{cb: cb, data: data})
	}

	for _, fld := range flds {
		err := (*fld).RegisterEventChange(cb, data)
		if err != nil {
			return err
		}
	}

	return nil
}

// ForceSync Force the synchronization of a folder
func (f *Folders) ForceSync(id string) error {
	fc := f.Get(id)
	if fc == nil {
		return fmt.Errorf("Unknown id")
	}
	return (*fc).Sync()
}

// IsFolderInSync Returns true when folder is in sync
func (f *Folders) IsFolderInSync(id string) (bool, error) {
	fc := f.Get(id)
	if fc == nil {
		return false, fmt.Errorf("Unknown id")
	}
	return (*fc).IsInSync()
}

//*** Private functions ***

// Use XML format and not json to be able to save/load all fields including
// ones that are masked in json (IOW defined with `json:"-"`)
type xmlFolders struct {
	XMLName xml.Name               `xml:"folders"`
	Version string                 `xml:"version,attr"`
	Folders []xsapiv1.FolderConfig `xml:"folders"`
}

// foldersConfigRead reads folders config from disk
func foldersConfigRead(file string, folders *[]xsapiv1.FolderConfig) error {
	if !common.Exists(file) {
		return fmt.Errorf("No folder config file found (%s)", file)
	}

	ffMutex.Lock()
	defer ffMutex.Unlock()

	fd, err := os.Open(file)
	defer fd.Close()
	if err != nil {
		return err
	}

	data := xmlFolders{}
	err = xml.NewDecoder(fd).Decode(&data)
	if err == nil {
		// Decode old type encoding (number) for backward compatibility
		for i, d := range data.Folders {
			switch d.Type {
			case "1":
				data.Folders[i].Type = xsapiv1.TypePathMap
			case "2":
				data.Folders[i].Type = xsapiv1.TypeCloudSync
			case "3":
				data.Folders[i].Type = xsapiv1.TypeCifsSmb
			}
		}

		*folders = data.Folders
	}
	return err
}

// foldersConfigWrite writes folders config on disk
func foldersConfigWrite(file string, folders []xsapiv1.FolderConfig) error {
	ffMutex.Lock()
	defer ffMutex.Unlock()

	fd, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	defer fd.Close()
	if err != nil {
		return err
	}

	data := &xmlFolders{
		Version: "1",
		Folders: folders,
	}

	enc := xml.NewEncoder(fd)
	enc.Indent("", "  ")
	return enc.Encode(data)
}
