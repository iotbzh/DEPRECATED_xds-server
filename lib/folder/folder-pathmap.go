package folder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
	uuid "github.com/satori/go.uuid"
)

// IFOLDER interface implementation for native/path mapping folders

// PathMap .
type PathMap struct {
	globalConfig *xdsconfig.Config
	config       FolderConfig
}

// NewFolderPathMap Create a new instance of PathMap
func NewFolderPathMap(gc *xdsconfig.Config) *PathMap {
	f := PathMap{
		globalConfig: gc,
		config: FolderConfig{
			Status: StatusDisable,
		},
	}
	return &f
}

// NewUID Get a UUID
func (f *PathMap) NewUID(suffix string) string {
	return uuid.NewV1().String() + "_" + suffix
}

// Add a new folder
func (f *PathMap) Add(cfg FolderConfig) (*FolderConfig, error) {
	if cfg.DataPathMap.ServerPath == "" {
		return nil, fmt.Errorf("ServerPath must be set")
	}

	// Use shareRootDir if ServerPath is a relative path
	dir := cfg.DataPathMap.ServerPath
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(f.globalConfig.FileConf.ShareRootDir, dir)
	}

	// Sanity check
	if !common.Exists(dir) {
		// try to create if not existing
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("Cannot create ServerPath directory: %s", dir)
		}
	}
	if !common.Exists(dir) {
		return nil, fmt.Errorf("ServerPath directory is not accessible: %s", dir)
	}

	f.config = cfg
	f.config.RootPath = dir
	f.config.DataPathMap.ServerPath = dir
	f.config.IsInSync = true

	// Verify file created by XDS agent when needed
	if cfg.DataPathMap.CheckFile != "" {
		errMsg := "ServerPath sanity check error (%d): %v"
		ckFile := f.ConvPathCli2Svr(cfg.DataPathMap.CheckFile)
		if !common.Exists(ckFile) {
			return nil, fmt.Errorf(errMsg, 1, "file not present")
		}
		if cfg.DataPathMap.CheckContent != "" {
			fd, err := os.OpenFile(ckFile, os.O_APPEND|os.O_RDWR, 0600)
			if err != nil {
				return nil, fmt.Errorf(errMsg, 2, err)
			}
			defer fd.Close()

			// Check specific message written by agent
			content, err := ioutil.ReadAll(fd)
			if err != nil {
				return nil, fmt.Errorf(errMsg, 3, err)
			}
			if string(content) != cfg.DataPathMap.CheckContent {
				return nil, fmt.Errorf(errMsg, 4, "file content differ")
			}

			// Write a specific message that will be check back on agent side
			msg := "Pathmap checked message written by xds-server ID: " + f.globalConfig.ServerUID + "\n"
			if n, err := fd.WriteString(msg); n != len(msg) || err != nil {
				return nil, fmt.Errorf(errMsg, 5, err)
			}
		}
	}

	f.config.Status = StatusEnable

	return &f.config, nil
}

// GetConfig Get public part of folder config
func (f *PathMap) GetConfig() FolderConfig {
	return f.config
}

// GetFullPath returns the full path of a directory (from server POV)
func (f *PathMap) GetFullPath(dir string) string {
	if &dir == nil {
		return f.config.DataPathMap.ServerPath
	}
	return filepath.Join(f.config.DataPathMap.ServerPath, dir)
}

// ConvPathCli2Svr Convert path from Client to Server
func (f *PathMap) ConvPathCli2Svr(s string) string {
	if f.config.ClientPath != "" && f.config.DataPathMap.ServerPath != "" {
		return strings.Replace(s,
			f.config.ClientPath,
			f.config.DataPathMap.ServerPath,
			-1)
	}
	return s
}

// ConvPathSvr2Cli Convert path from Server to Client
func (f *PathMap) ConvPathSvr2Cli(s string) string {
	if f.config.ClientPath != "" && f.config.DataPathMap.ServerPath != "" {
		return strings.Replace(s,
			f.config.DataPathMap.ServerPath,
			f.config.ClientPath,
			-1)
	}
	return s
}

// Remove a folder
func (f *PathMap) Remove() error {
	// nothing to do
	return nil
}

// RegisterEventChange requests registration for folder change event
func (f *PathMap) RegisterEventChange(cb *EventCB, data *EventCBData) error {
	return nil
}

// UnRegisterEventChange remove registered callback
func (f *PathMap) UnRegisterEventChange() error {
	return nil
}

// Sync Force folder files synchronization
func (f *PathMap) Sync() error {
	return nil
}

// IsInSync Check if folder files are in-sync
func (f *PathMap) IsInSync() (bool, error) {
	return true, nil
}
