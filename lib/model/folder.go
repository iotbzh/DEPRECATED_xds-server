package model

import (
	"fmt"

	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/syncthing"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
)

// Folder Represent a an XDS folder
type Folder struct {
	Conf *xdsconfig.Config
	SThg *st.SyncThing
}

// NewFolder Create a new instance of Model Folder
func NewFolder(cfg *xdsconfig.Config, st *st.SyncThing) *Folder {
	return &Folder{
		Conf: cfg,
		SThg: st,
	}
}

// GetFolderFromID retrieves the Folder config from id
func (c *Folder) GetFolderFromID(id string) *xdsconfig.FolderConfig {
	if idx := c.Conf.Folders.GetIdx(id); idx != -1 {
		return &c.Conf.Folders[idx]
	}
	return nil
}

// UpdateAll updates all the current configuration
func (c *Folder) UpdateAll(newCfg xdsconfig.Config) error {
	return fmt.Errorf("Not Supported")
	/*
		if err := VerifyConfig(newCfg); err != nil {
			return err
		}

		// TODO: c.Builder = c.Builder.Update(newCfg.Builder)
		c.Folders = c.Folders.Update(newCfg.Folders)

		// FIXME To be tested & improved error handling
		for _, f := range c.Folders {
			if err := c.SThg.FolderChange(st.FolderChangeArg{
				ID:           f.ID,
				Label:        f.Label,
				RelativePath: f.RelativePath,
				SyncThingID:  f.SyncThingID,
				ShareRootDir: c.FileConf.ShareRootDir,
			}); err != nil {
				return err
			}
		}

		return nil
	*/
}

// UpdateFolder updates a specific folder into the current configuration
func (c *Folder) UpdateFolder(newFolder xdsconfig.FolderConfig) (xdsconfig.FolderConfig, error) {
	// rootPath should not be empty
	if newFolder.RootPath == "" {
		newFolder.RootPath = c.Conf.FileConf.ShareRootDir
	}

	// Sanity check of folder settings
	if err := newFolder.Verify(); err != nil {
		return xdsconfig.FolderConfig{}, err
	}

	// Normalize path (needed for Windows path including bashlashes)
	newFolder.RelativePath = common.PathNormalize(newFolder.RelativePath)

	// Update config folder
	c.Conf.Folders = c.Conf.Folders.Update(xdsconfig.FoldersConfig{newFolder})

	// Update Syncthing folder
	err := c.SThg.FolderChange(newFolder)

	newFolder.BuilderSThgID = c.Conf.Builder.SyncThingID // FIXME - should be removed after local ST config rework
	newFolder.Status = xdsconfig.FolderStatusEnable

	return newFolder, err
}

// DeleteFolder deletes a specific folder
func (c *Folder) DeleteFolder(id string) (xdsconfig.FolderConfig, error) {
	var fld xdsconfig.FolderConfig
	var err error

	if err = c.SThg.FolderDelete(id); err != nil {
		return fld, err
	}

	c.Conf.Folders, fld, err = c.Conf.Folders.Delete(id)

	return fld, err
}

// ForceSync Force the synchronization of a folder
func (c *Folder) ForceSync(id string) error {
	return c.SThg.FolderScan(id, "")
}

// IsFolderInSync Returns true when folder is in sync
func (c *Folder) IsFolderInSync(id string) (bool, error) {
	return c.SThg.IsFolderInSync(id)
}
