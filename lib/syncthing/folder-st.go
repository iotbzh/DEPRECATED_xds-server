package st

import (
	"fmt"
	"path/filepath"

	"github.com/iotbzh/xds-server/lib/folder"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
	"github.com/syncthing/syncthing/lib/config"
)

// IFOLDER interface implementation for syncthing

// STFolder .
type STFolder struct {
	globalConfig *xdsconfig.Config
	st           *SyncThing
	fConfig      folder.FolderConfig
	stfConfig    config.FolderConfiguration
}

// NewFolderST Create a new instance of STFolder
func (s *SyncThing) NewFolderST(gc *xdsconfig.Config) *STFolder {
	return &STFolder{
		globalConfig: gc,
		st:           s,
	}
}

// Add a new folder
func (f *STFolder) Add(cfg folder.FolderConfig) (*folder.FolderConfig, error) {

	// Sanity check
	if cfg.DataCloudSync.SyncThingID == "" {
		return nil, fmt.Errorf("device id not set (SyncThingID field)")
	}

	// rootPath should not be empty
	if cfg.RootPath == "" {
		cfg.RootPath = f.globalConfig.FileConf.ShareRootDir
	}

	f.fConfig = cfg

	f.fConfig.DataCloudSync.BuilderSThgID = f.st.MyID // FIXME - should be removed after local ST config rework

	// Update Syncthing folder
	// (expect if status is ErrorConfig)
	// TODO: add cache to avoid multiple requests on startup
	if f.fConfig.Status != folder.StatusErrorConfig {
		id, err := f.st.FolderChange(f.fConfig)
		if err != nil {
			return nil, err
		}

		f.stfConfig, err = f.st.FolderConfigGet(id)
		if err != nil {
			f.fConfig.Status = folder.StatusErrorConfig
			return nil, err
		}

		f.fConfig.Status = folder.StatusEnable
	}

	return &f.fConfig, nil
}

// GetConfig Get public part of folder config
func (f *STFolder) GetConfig() folder.FolderConfig {
	return f.fConfig
}

// GetFullPath returns the full path
func (f *STFolder) GetFullPath(dir string) string {
	if &dir == nil {
		dir = ""
	}
	if filepath.IsAbs(dir) {
		return filepath.Join(f.fConfig.RootPath, dir)
	}
	return filepath.Join(f.fConfig.RootPath, f.fConfig.ClientPath, dir)
}

// Remove a folder
func (f *STFolder) Remove() error {
	return f.st.FolderDelete(f.stfConfig.ID)
}

// Sync Force folder files synchronization
func (f *STFolder) Sync() error {
	return f.st.FolderScan(f.stfConfig.ID, "")
}

// IsInSync Check if folder files are in-sync
func (f *STFolder) IsInSync() (bool, error) {
	return f.st.IsFolderInSync(f.stfConfig.ID)
}
