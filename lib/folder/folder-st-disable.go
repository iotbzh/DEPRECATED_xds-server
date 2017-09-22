package folder

import (
	"github.com/iotbzh/xds-server/lib/xdsconfig"
	uuid "github.com/satori/go.uuid"
)

// IFOLDER interface implementation for disabled Syncthing folders
// It's a "fallback" interface used to keep syncthing folders config even
// when syncthing is not running.

// STFolderDisable .
type STFolderDisable struct {
	globalConfig *xdsconfig.Config
	config       FolderConfig
}

// NewFolderSTDisable Create a new instance of STFolderDisable
func NewFolderSTDisable(gc *xdsconfig.Config) *STFolderDisable {
	f := STFolderDisable{
		globalConfig: gc,
	}
	return &f
}

// NewUID Get a UUID
func (f *STFolderDisable) NewUID(suffix string) string {
	return uuid.NewV1().String() + "_" + suffix
}

// Add a new folder
func (f *STFolderDisable) Add(cfg FolderConfig) (*FolderConfig, error) {
	f.config = cfg
	f.config.Status = StatusDisable
	f.config.IsInSync = false
	return &f.config, nil
}

// GetConfig Get public part of folder config
func (f *STFolderDisable) GetConfig() FolderConfig {
	return f.config
}

// GetFullPath returns the full path of a directory (from server POV)
func (f *STFolderDisable) GetFullPath(dir string) string {
	return ""
}

// ConvPathCli2Svr Convert path from Client to Server
func (f *STFolderDisable) ConvPathCli2Svr(s string) string {
	return ""
}

// ConvPathSvr2Cli Convert path from Server to Client
func (f *STFolderDisable) ConvPathSvr2Cli(s string) string {
	return ""
}

// Remove a folder
func (f *STFolderDisable) Remove() error {
	return nil
}

// RegisterEventChange requests registration for folder change event
func (f *STFolderDisable) RegisterEventChange(cb *EventCB, data *EventCBData) error {
	return nil
}

// UnRegisterEventChange remove registered callback
func (f *STFolderDisable) UnRegisterEventChange() error {
	return nil
}

// Sync Force folder files synchronization
func (f *STFolderDisable) Sync() error {
	return nil
}

// IsInSync Check if folder files are in-sync
func (f *STFolderDisable) IsInSync() (bool, error) {
	return false, nil
}
