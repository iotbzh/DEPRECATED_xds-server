package st

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/iotbzh/xds-server/lib/folder"
	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/protocol"
)

// FolderLoadFromStConfig Load/Retrieve folder config from syncthing database
func (s *SyncThing) FolderLoadFromStConfig(f *[]folder.FolderConfig) error {

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
		*f = append(*f, folder.FolderConfig{
			ID:            stFld.ID,
			Label:         stFld.Label,
			ClientPath:    strings.TrimRight(cliPath, "/"),
			Type:          folder.TypeCloudSync,
			Status:        folder.StatusDisable,
			DefaultSdk:    defaultSdk,
			RootPath:      s.conf.FileConf.ShareRootDir,
			DataCloudSync: folder.CloudSyncConfig{SyncThingID: devID},
		})
	}

	return nil
}

// FolderChange is called when configuration has changed
func (s *SyncThing) FolderChange(f folder.FolderConfig) (string, error) {

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

	newDevice := config.DeviceConfiguration{
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

	folder := config.FolderConfiguration{
		ID:    id,
		Label: label,
		Path:  filepath.Join(s.conf.FileConf.ShareRootDir, f.ClientPath),
	}

	if s.conf.FileConf.SThgConf.RescanIntervalS > 0 {
		folder.RescanIntervalS = s.conf.FileConf.SThgConf.RescanIntervalS
	}

	folder.Devices = append(folder.Devices, config.FolderDeviceConfiguration{
		DeviceID: newDevice.DeviceID,
	})

	found = false
	var fld config.FolderConfiguration
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
func (s *SyncThing) FolderConfigGet(folderID string) (config.FolderConfiguration, error) {
	fc := config.FolderConfiguration{}
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
