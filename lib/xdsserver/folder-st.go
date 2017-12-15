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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	st "github.com/iotbzh/xds-server/lib/syncthing"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
	uuid "github.com/satori/go.uuid"
	"github.com/syncthing/syncthing/lib/config"
)

// IFOLDER interface implementation for syncthing

// STFolder .
type STFolder struct {
	*Context
	st        *st.SyncThing
	fConfig   xsapiv1.FolderConfig
	stfConfig config.FolderConfiguration
	eventIDs  []string
}

var stEventMonitored = []string{st.EventStateChanged, st.EventFolderPaused}

// NewFolderST Create a new instance of STFolder
func NewFolderST(ctx *Context, sthg *st.SyncThing) *STFolder {
	return &STFolder{
		Context: ctx,
		st:      sthg,
	}
}

// NewUID Get a UUID
func (f *STFolder) NewUID(suffix string) string {
	i := len(f.st.MyID)
	if i > 15 {
		i = 15
	}
	uuid := uuid.NewV1().String()[:14] + f.st.MyID[:i]
	if len(suffix) > 0 {
		uuid += "_" + suffix
	}
	return uuid
}

// Add a new folder
func (f *STFolder) Add(cfg xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error) {

	// Sanity check
	if cfg.DataCloudSync.SyncThingID == "" {
		return nil, fmt.Errorf("device id not set (SyncThingID field)")
	}

	// rootPath should not be empty
	if cfg.RootPath == "" {
		cfg.RootPath = f.Config.FileConf.ShareRootDir
	}

	f.fConfig = cfg

	// Update Syncthing folder
	_, err := f.st.FolderChange(f.fConfig)
	if err != nil {
		return nil, err
	}

	// Use Setup function to setup remains fields
	return f.Setup(f.fConfig)
}

// Setup Setup local project config
func (f *STFolder) Setup(fld xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error) {

	var err error

	// Update folder Config
	f.fConfig = fld

	// Retrieve Syncthing folder config
	f.stfConfig, err = f.st.FolderConfigGet(f.fConfig.ID)
	if err != nil {
		f.fConfig.Status = xsapiv1.StatusErrorConfig
		return nil, err
	}

	// Register to events to update folder status
	for _, evName := range stEventMonitored {
		evID, err := f.st.Events.Register(evName, f.cbEventState, f.fConfig.ID, nil)
		if err != nil {
			return nil, err
		}
		f.eventIDs = append(f.eventIDs, evID)
	}

	f.fConfig.IsInSync = false // will be updated later by events
	f.fConfig.Status = xsapiv1.StatusEnable

	return &f.fConfig, nil
}

// GetConfig Get public part of folder config
func (f *STFolder) GetConfig() xsapiv1.FolderConfig {
	return f.fConfig
}

// GetFullPath returns the full path of a directory (from server POV)
func (f *STFolder) GetFullPath(dir string) string {
	if &dir == nil {
		dir = ""
	}
	if filepath.IsAbs(dir) {
		return filepath.Join(f.fConfig.RootPath, dir)
	}
	return filepath.Join(f.fConfig.RootPath, f.fConfig.ClientPath, dir)
}

// ConvPathCli2Svr Convert path from Client to Server
func (f *STFolder) ConvPathCli2Svr(s string) string {
	if f.fConfig.ClientPath != "" && f.fConfig.RootPath != "" {
		return strings.Replace(s,
			f.fConfig.ClientPath,
			f.fConfig.RootPath+"/"+f.fConfig.ClientPath,
			-1)
	}
	return s
}

// ConvPathSvr2Cli Convert path from Server to Client
func (f *STFolder) ConvPathSvr2Cli(s string) string {
	if f.fConfig.ClientPath != "" && f.fConfig.RootPath != "" {
		return strings.Replace(s,
			f.fConfig.RootPath+"/"+f.fConfig.ClientPath,
			f.fConfig.ClientPath,
			-1)
	}
	return s
}

// Remove a folder
func (f *STFolder) Remove() error {
	var err1 error
	// Un-register events
	for _, evID := range f.eventIDs {
		if err := f.st.Events.UnRegister(evID); err != nil && err1 == nil {
			// only report 1st error
			err1 = err
		}
	}

	// Delete in Syncthing
	err2 := f.st.FolderDelete(f.stfConfig.ID)

	// Delete folder on server side
	err3 := os.RemoveAll(f.GetFullPath(""))

	if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	}
	return err3
}

// Update update some fields of a folder
func (f *STFolder) Update(cfg xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error) {
	if f.fConfig.ID != cfg.ID {
		return nil, fmt.Errorf("Invalid id")
	}
	f.fConfig = cfg
	return &f.fConfig, nil
}

// Sync Force folder files synchronization
func (f *STFolder) Sync() error {
	return f.st.FolderScan(f.stfConfig.ID, "")
}

// IsInSync Check if folder files are in-sync
func (f *STFolder) IsInSync() (bool, error) {
	sts, err := f.st.IsFolderInSync(f.stfConfig.ID)
	if err != nil {
		return false, err
	}
	f.fConfig.IsInSync = sts
	return sts, nil
}

// callback use to update IsInSync status
func (f *STFolder) cbEventState(ev st.Event, data *st.EventsCBData) {
	prevSync := f.fConfig.IsInSync
	prevStatus := f.fConfig.Status

	switch ev.Type {

	case st.EventStateChanged:
		to := ev.Data["to"]
		switch to {
		case "scanning", "syncing":
			f.fConfig.Status = xsapiv1.StatusSyncing
		case "idle":
			f.fConfig.Status = xsapiv1.StatusEnable
		}
		f.fConfig.IsInSync = (to == "idle")

	case st.EventFolderPaused:
		if f.fConfig.Status == xsapiv1.StatusEnable {
			f.fConfig.Status = xsapiv1.StatusPause
		}
		f.fConfig.IsInSync = false
	}

	if prevSync != f.fConfig.IsInSync || prevStatus != f.fConfig.Status {
		// Emit Folder state change event
		if err := f.events.Emit(xsapiv1.EVTFolderStateChange, &f.fConfig, ""); err != nil {
			f.Log.Warningf("Cannot notify folder change: %v", err)
		}
	}
}
