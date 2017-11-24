package st

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iotbzh/xds-server/lib/folder"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
	uuid "github.com/satori/go.uuid"
	"github.com/syncthing/syncthing/lib/config"
)

// IFOLDER interface implementation for syncthing

// STFolder .
type STFolder struct {
	globalConfig      *xdsconfig.Config
	st                *SyncThing
	fConfig           folder.FolderConfig
	stfConfig         config.FolderConfiguration
	eventIDs          []int
	eventChangeCB     *folder.EventCB
	eventChangeCBData *folder.EventCBData
}

// NewFolderST Create a new instance of STFolder
func (s *SyncThing) NewFolderST(gc *xdsconfig.Config) *STFolder {
	return &STFolder{
		globalConfig: gc,
		st:           s,
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

	// Update Syncthing folder
	// (except if status is ErrorConfig)
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

		// Register to events to update folder status
		for _, evName := range []string{EventStateChanged, EventFolderPaused} {
			evID, err := f.st.Events.Register(evName, f.cbEventState, id, nil)
			if err != nil {
				return nil, err
			}
			f.eventIDs = append(f.eventIDs, evID)
		}

		f.fConfig.IsInSync = false // will be updated later by events
		f.fConfig.Status = folder.StatusEnable
	}

	return &f.fConfig, nil
}

// GetConfig Get public part of folder config
func (f *STFolder) GetConfig() folder.FolderConfig {
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
	err := f.st.FolderDelete(f.stfConfig.ID)

	// Delete folder on server side
	err2 := os.RemoveAll(f.GetFullPath(""))

	if err != nil {
		return err
	}
	return err2
}

// Update update some fields of a folder
func (f *STFolder) Update(cfg folder.FolderConfig) (*folder.FolderConfig, error) {
	if f.fConfig.ID != cfg.ID {
		return nil, fmt.Errorf("Invalid id")
	}
	f.fConfig = cfg
	return &f.fConfig, nil
}

// RegisterEventChange requests registration for folder event change
func (f *STFolder) RegisterEventChange(cb *folder.EventCB, data *folder.EventCBData) error {
	f.eventChangeCB = cb
	f.eventChangeCBData = data
	return nil
}

// UnRegisterEventChange remove registered callback
func (f *STFolder) UnRegisterEventChange() error {
	f.eventChangeCB = nil
	f.eventChangeCBData = nil
	return nil
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
func (f *STFolder) cbEventState(ev Event, data *EventsCBData) {
	prevSync := f.fConfig.IsInSync
	prevStatus := f.fConfig.Status

	switch ev.Type {

	case EventStateChanged:
		to := ev.Data["to"]
		switch to {
		case "scanning", "syncing":
			f.fConfig.Status = folder.StatusSyncing
		case "idle":
			f.fConfig.Status = folder.StatusEnable
		}
		f.fConfig.IsInSync = (to == "idle")

	case EventFolderPaused:
		if f.fConfig.Status == folder.StatusEnable {
			f.fConfig.Status = folder.StatusPause
		}
		f.fConfig.IsInSync = false
	}

	if f.eventChangeCB != nil &&
		(prevSync != f.fConfig.IsInSync || prevStatus != f.fConfig.Status) {
		cpConf := f.fConfig
		(*f.eventChangeCB)(&cpConf, f.eventChangeCBData)
	}
}
