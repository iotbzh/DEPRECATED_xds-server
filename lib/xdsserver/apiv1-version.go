package xdsserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
)

// getInfo : return various information about server
func (s *APIService) getVersion(c *gin.Context) {
	response := xsapiv1.Version{
		ID:            s.Config.ServerUID,
		Version:       s.Config.Version,
		APIVersion:    s.Config.APIVersion,
		VersionGitTag: s.Config.VersionGitTag,
	}

	c.JSON(http.StatusOK, response)
}
