package apiv1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type version struct {
	Version       string `json:"version"`
	APIVersion    string `json:"apiVersion"`
	VersionGitTag string `json:"gitTag"`
}

// getInfo : return various information about server
func (s *APIService) getVersion(c *gin.Context) {
	response := version{
		Version:       s.cfg.Version,
		APIVersion:    s.cfg.APIVersion,
		VersionGitTag: s.cfg.VersionGitTag,
	}

	c.JSON(http.StatusOK, response)
}
