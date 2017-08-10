package apiv1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/folder"
)

// getFolders returns all folders configuration
func (s *APIService) getFolders(c *gin.Context) {
	c.JSON(http.StatusOK, s.mfolders.GetConfigArr())
}

// getFolder returns a specific folder configuration
func (s *APIService) getFolder(c *gin.Context) {
	f := s.mfolders.Get(c.Param("id"))
	if f == nil {
		common.APIError(c, "Invalid id")
		return
	}

	c.JSON(http.StatusOK, (*f).GetConfig())
}

// addFolder adds a new folder to server config
func (s *APIService) addFolder(c *gin.Context) {
	var cfgArg folder.FolderConfig
	if c.BindJSON(&cfgArg) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	s.log.Debugln("Add folder config: ", cfgArg)

	newFld, err := s.mfolders.Add(cfgArg)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, newFld)
}

// delFolder deletes folder from server config
func (s *APIService) delFolder(c *gin.Context) {
	id := c.Param("id")

	s.log.Debugln("Delete folder id ", id)

	delEntry, err := s.mfolders.Delete(id)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, delEntry)

}
