package xdsconfig

import (
	"fmt"
)

// FoldersConfig contains all the folder configurations
type FoldersConfig []FolderConfig

// GetIdx returns the index of the folder matching id in FoldersConfig array
func (c FoldersConfig) GetIdx(id string) int {
	for i := range c {
		if id == c[i].ID {
			return i
		}
	}
	return -1
}

// Update is used to fully update or add a new FolderConfig
func (c FoldersConfig) Update(newCfg FoldersConfig) FoldersConfig {
	for i := range newCfg {
		found := false
		for j := range c {
			if newCfg[i].ID == c[j].ID {
				c[j] = newCfg[i]
				found = true
				break
			}
		}
		if !found {
			c = append(c, newCfg[i])
		}
	}
	return c
}

// Delete is used to delete a folder matching id in FoldersConfig array
func (c FoldersConfig) Delete(id string) (FoldersConfig, FolderConfig, error) {
	if idx := c.GetIdx(id); idx != -1 {
		f := c[idx]
		c = append(c[:idx], c[idx+1:]...)
		return c, f, nil
	}

	return c, FolderConfig{}, fmt.Errorf("invalid id")
}
