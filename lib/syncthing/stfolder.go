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

package st

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/iotbzh/xds-server/lib/xsapiv1"
	stconfig "github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/protocol"
)

// FolderLoadFromStConfig Load/Retrieve folder config from syncthing database
func (s *SyncThing) FolderLoadFromStConfig(f *[]xsapiv1.FolderConfig) error {

	defaultSdk := "" // cannot know which was the default sdk

	stCfg, err := s.ConfigGet()
	if err != nil {
		return err
	}
	if len(stCfg.Devices) < 1 {
		return fmt.Errorf("Cannot load syncthing config: no device defined")
	}
	devID := stCfg.Devices[0].DeviceID.String()
	if devID == s.MyID {
		if len(stCfg.Devices) < 2 {
			return fmt.Errorf("Cannot load syncthing config: no valid device found")
		}
		devID = stCfg.Devices[1].DeviceID.String()
	}

	for _, stFld := range stCfg.Folders {
		cliPath := strings.TrimPrefix(stFld.Path, s.conf.FileConf.ShareRootDir)
		if cliPath == "" {
			cliPath = stFld.Path
		}
		*f = append(*f, xsapiv1.FolderConfig{
			ID:            stFld.ID,
			Label:         stFld.Label,
			ClientPath:    strings.TrimRight(cliPath, "/"),
			Type:          xsapiv1.TypeCloudSync,
			Status:        xsapiv1.StatusDisable,
			DefaultSdk:    defaultSdk,
			RootPath:      s.conf.FileConf.ShareRootDir,
			DataCloudSync: xsapiv1.CloudSyncConfig{SyncThingID: devID},
		})
	}

	return nil
}

// FolderChange is called when configuration has changed
func (s *SyncThing) FolderChange(f xsapiv1.FolderConfig) (string, error) {

	// Get current config
	stCfg, err := s.ConfigGet()
	if err != nil {
		s.log.Errorln(err)
		return "", err
	}

	stClientID := f.DataCloudSync.SyncThingID
	// Add new Device if needed
	var devID protocol.DeviceID
	if err := devID.UnmarshalText([]byte(stClientID)); err != nil {
		s.log.Errorf("not a valid device id (err %v)", err)
		return "", err
	}

	newDevice := stconfig.DeviceConfiguration{
		DeviceID:  devID,
		Name:      stClientID,
		Addresses: []string{"dynamic"},
	}

	var found = false
	for _, device := range stCfg.Devices {
		if device.DeviceID == devID {
			found = true
			break
		}
	}
	if !found {
		stCfg.Devices = append(stCfg.Devices, newDevice)
	}

	// Add or update Folder settings
	var label, id string
	if label = f.Label; label == "" {
		label = strings.Split(id, "/")[0]
	}
	if id = f.ID; id == "" {
		id = stClientID[0:15] + "_" + label
	}

	folder := stconfig.FolderConfiguration{
		ID:    id,
		Label: label,
		Path:  filepath.Join(s.conf.FileConf.ShareRootDir, f.ClientPath),
	}

	if s.conf.FileConf.SThgConf.RescanIntervalS > 0 {
		folder.RescanIntervalS = s.conf.FileConf.SThgConf.RescanIntervalS
	}

	folder.Devices = append(folder.Devices, stconfig.FolderDeviceConfiguration{
		DeviceID: newDevice.DeviceID,
	})

	found = false
	var fld stconfig.FolderConfiguration
	for _, fld = range stCfg.Folders {
		if folder.ID == fld.ID {
			fld = folder
			found = true
			break
		}
	}
	if !found {
		stCfg.Folders = append(stCfg.Folders, folder)
		fld = stCfg.Folders[0]
	}

	err = s.ConfigSet(stCfg)
	if err != nil {
		s.log.Errorln(err)
	}

	return id, nil
}

// FolderDelete is called to delete a folder config
func (s *SyncThing) FolderDelete(id string) error {
	// Get current config
	stCfg, err := s.ConfigGet()
	if err != nil {
		s.log.Errorln(err)
		return err
	}

	for i, fld := range stCfg.Folders {
		if id == fld.ID {
			stCfg.Folders = append(stCfg.Folders[:i], stCfg.Folders[i+1:]...)
			err = s.ConfigSet(stCfg)
			if err != nil {
				s.log.Errorln(err)
				return err
			}
		}
	}

	return nil
}

// FolderConfigGet Returns the configuration of a specific folder
func (s *SyncThing) FolderConfigGet(folderID string) (stconfig.FolderConfiguration, error) {
	fc := stconfig.FolderConfiguration{}
	if folderID == "" {
		return fc, fmt.Errorf("folderID not set")
	}
	cfg, err := s.ConfigGet()
	if err != nil {
		return fc, err
	}
	for _, f := range cfg.Folders {
		if f.ID == folderID {
			fc = f
			return fc, nil
		}
	}
	return fc, fmt.Errorf("id not found")
}

// FolderStatus Returns all information about the current
func (s *SyncThing) FolderStatus(folderID string) (*FolderStatus, error) {
	var data []byte
	var res FolderStatus
	if folderID == "" {
		return nil, fmt.Errorf("folderID not set")
	}
	if err := s.client.HTTPGet("db/status?folder="+folderID, &data); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// IsFolderInSync Returns true when folder is in sync
func (s *SyncThing) IsFolderInSync(folderID string) (bool, error) {
	sts, err := s.FolderStatus(folderID)
	if err != nil {
		return false, err
	}
	return sts.NeedBytes == 0 && sts.State == "idle", nil
}

// FolderScan Request immediate folder scan.
// Scan all folders if folderID param is empty
func (s *SyncThing) FolderScan(folderID string, subpath string) error {
	url := "db/scan"
	if folderID != "" {
		url += "?folder=" + folderID

		if subpath != "" {
			url += "&sub=" + subpath
		}
	}
	return s.client.HTTPPost(url, "")
}
