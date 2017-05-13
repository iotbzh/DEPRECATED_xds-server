package xdsconfig

import (
	"fmt"
	"log"
	"path/filepath"
)

// FolderType constances
const (
	FolderTypeDocker           = 0
	FolderTypeWindowsSubsystem = 1
	FolderTypeCloudSync        = 2

	FolderStatusErrorConfig = "ErrorConfig"
	FolderStatusDisable     = "Disable"
	FolderStatusEnable      = "Enable"
)

// FolderType is the type of sharing folder
type FolderType int

// FolderConfig is the config for one folder
type FolderConfig struct {
	ID            string     `json:"id" binding:"required"`
	Label         string     `json:"label"`
	RelativePath  string     `json:"path"`
	Type          FolderType `json:"type"`
	SyncThingID   string     `json:"syncThingID"`
	BuilderSThgID string     `json:"builderSThgID"`
	Status        string     `json:"status"`

	// Private fields
	rootPath string
}

// NewFolderConfig creates a new folder object
func NewFolderConfig(id, label, rootDir, path string) FolderConfig {
	return FolderConfig{
		ID:           id,
		Label:        label,
		RelativePath: path,
		Type:         FolderTypeCloudSync,
		SyncThingID:  "",
		Status:       FolderStatusDisable,
		rootPath:     rootDir,
	}
}

// GetFullPath returns the full path
func (c *FolderConfig) GetFullPath(dir string) string {
	if &dir == nil {
		dir = ""
	}
	if filepath.IsAbs(dir) {
		return filepath.Join(c.rootPath, dir)
	}
	return filepath.Join(c.rootPath, c.RelativePath, dir)
}

// FolderVerify is called to verify that a configuration is valid
func FolderVerify(fCfg FolderConfig) error {
	var err error

	if fCfg.Type != FolderTypeCloudSync {
		err = fmt.Errorf("Unsupported folder type")
	}

	if fCfg.SyncThingID == "" {
		err = fmt.Errorf("device id not set (SyncThingID field)")
	}

	if fCfg.rootPath == "" {
		err = fmt.Errorf("rootPath must not be empty")
	}

	if err != nil {
		fCfg.Status = FolderStatusErrorConfig
		log.Printf("ERROR FolderVerify: %v\n", err)
	}

	return err
}
