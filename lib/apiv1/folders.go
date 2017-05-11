package apiv1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-server/lib/common"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
)

// getFolders returns all folders configuration
func (s *APIService) getFolders(c *gin.Context) {
	confMut.Lock()
	defer confMut.Unlock()

	c.JSON(http.StatusOK, s.cfg.Folders)
}

// getFolder returns a specific folder configuration
func (s *APIService) getFolder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 0 || id > len(s.cfg.Folders) {
		common.APIError(c, "Invalid id")
		return
	}

	confMut.Lock()
	defer confMut.Unlock()

	c.JSON(http.StatusOK, s.cfg.Folders[id])
}

// addFolder adds a new folder to server config
func (s *APIService) addFolder(c *gin.Context) {
	var cfgArg xdsconfig.FolderConfig
	if c.BindJSON(&cfgArg) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	confMut.Lock()
	defer confMut.Unlock()

	s.log.Debugln("Add folder config: ", cfgArg)

	newFld, err := s.cfg.UpdateFolder(cfgArg)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, newFld)
}

// delFolder deletes folder from server config
func (s *APIService) delFolder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.APIError(c, "Invalid id")
		return
	}

	confMut.Lock()
	defer confMut.Unlock()

	s.log.Debugln("Delete folder id ", id)

	var delEntry xdsconfig.FolderConfig
	var err error
	if delEntry, err = s.cfg.DeleteFolder(id); err != nil {
		common.APIError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, delEntry)

}
