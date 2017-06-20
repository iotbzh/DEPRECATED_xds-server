package st

import (
	"path/filepath"
	"strings"

	"github.com/iotbzh/xds-server/lib/xdsconfig"
	"github.com/syncthing/syncthing/lib/config"
	"github.com/syncthing/syncthing/lib/protocol"
)

// FolderChange is called when configuration has changed
func (s *SyncThing) FolderChange(f xdsconfig.FolderConfig) error {

	// Get current config
	stCfg, err := s.ConfigGet()
	if err != nil {
		s.log.Errorln(err)
		return err
	}

	// Add new Device if needed
	var devID protocol.DeviceID
	if err := devID.UnmarshalText([]byte(f.SyncThingID)); err != nil {
		s.log.Errorf("not a valid device id (err %v)\n", err)
		return err
	}

	newDevice := config.DeviceConfiguration{
		DeviceID:  devID,
		Name:      f.SyncThingID,
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
		id = f.SyncThingID[0:15] + "_" + label
	}

	folder := config.FolderConfiguration{
		ID:      id,
		Label:   label,
		RawPath: filepath.Join(s.conf.FileConf.ShareRootDir, f.RelativePath),
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

	return nil
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
