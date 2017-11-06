package model

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/folder"
	"github.com/iotbzh/xds-server/lib/syncthing"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
	"github.com/syncthing/syncthing/lib/sync"
)

// Folders Represent a an XDS folders
type Folders struct {
	fileOnDisk string
	Conf       *xdsconfig.Config
	Log        *logrus.Logger
	SThg       *st.SyncThing
	folders    map[string]*folder.IFOLDER
	registerCB []RegisteredCB
}

type RegisteredCB struct {
	cb   *folder.EventCB
	data *folder.EventCBData
}

// Mutex to make add/delete atomic
var fcMutex = sync.NewMutex()
var ffMutex = sync.NewMutex()

// FoldersNew Create a new instance of Model Folders
func FoldersNew(cfg *xdsconfig.Config, st *st.SyncThing) *Folders {
	file, _ := xdsconfig.FoldersConfigFilenameGet()
	return &Folders{
		fileOnDisk: file,
		Conf:       cfg,
		Log:        cfg.Log,
		SThg:       st,
		folders:    make(map[string]*folder.IFOLDER),
		registerCB: []RegisteredCB{},
	}
}

// LoadConfig Load folders configuration from disk
func (f *Folders) LoadConfig() error {
	var flds []folder.FolderConfig
	var stFlds []folder.FolderConfig

	// load from disk
	if f.Conf.Options.NoFolderConfig {
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
				if xf.Type != folder.TypeCloudSync {
					flds[i].Status = folder.StatusErrorConfig
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
			if xf.Type != folder.TypeCloudSync {
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
				flds[i].Status = folder.StatusErrorConfig
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
func (f *Folders) Get(id string) *folder.IFOLDER {
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
func (f *Folders) GetConfigArr() []folder.FolderConfig {
	fcMutex.Lock()
	defer fcMutex.Unlock()

	return f.getConfigArrUnsafe()
}

// getConfigArrUnsafe Same as GetConfigArr without mutex protection
func (f *Folders) getConfigArrUnsafe() []folder.FolderConfig {
	conf := []folder.FolderConfig{}
	for _, v := range f.folders {
		conf = append(conf, (*v).GetConfig())
	}
	return conf
}

// Add adds a new folder
func (f *Folders) Add(newF folder.FolderConfig) (*folder.FolderConfig, error) {
	return f.createUpdate(newF, true, false)
}

// CreateUpdate creates or update a folder
func (f *Folders) createUpdate(newF folder.FolderConfig, create bool, initial bool) (*folder.FolderConfig, error) {

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
	var fld folder.IFOLDER
	switch newF.Type {
	// SYNCTHING
	case folder.TypeCloudSync:
		if f.SThg != nil {
			fld = f.SThg.NewFolderST(f.Conf)
		} else {
			f.Log.Debugf("Disable project %v (syncthing not initialized)", newF.ID)
			fld = folder.NewFolderSTDisable(f.Conf)
		}

	// PATH MAP
	case folder.TypePathMap:
		fld = folder.NewFolderPathMap(f.Conf)
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
		newF.Status = folder.StatusDisable
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
		newF.Status = folder.StatusErrorConfig
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
func (f *Folders) Delete(id string) (folder.FolderConfig, error) {
	var err error

	fcMutex.Lock()
	defer fcMutex.Unlock()

	fld := folder.FolderConfig{}
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

// RegisterEventChange requests registration for folder event change
func (f *Folders) RegisterEventChange(id string, cb *folder.EventCB, data *folder.EventCBData) error {

	flds := make(map[string]*folder.IFOLDER)
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
	XMLName xml.Name              `xml:"folders"`
	Version string                `xml:"version,attr"`
	Folders []folder.FolderConfig `xml:"folders"`
}

// foldersConfigRead reads folders config from disk
func foldersConfigRead(file string, folders *[]folder.FolderConfig) error {
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
		*folders = data.Folders
	}
	return err
}

// foldersConfigWrite writes folders config on disk
func foldersConfigWrite(file string, folders []folder.FolderConfig) error {
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
