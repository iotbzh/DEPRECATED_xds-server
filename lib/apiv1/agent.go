package apiv1

import (
	"net/http"

	"path/filepath"

	"github.com/gin-gonic/gin"
)

type XDSAgentTarball struct {
	OS      string `json:"os"`
	FileURL string `json:"fileUrl"`
}
type XDSAgentInfo struct {
	Tarballs []XDSAgentTarball `json:"tarballs"`
}

// getXdsAgentInfo : return various information about Xds Agent
func (s *APIService) getXdsAgentInfo(c *gin.Context) {
	// TODO: retrieve link dynamically by reading assets/xds-agent-tarballs
	tarballDir := "assets/xds-agent-tarballs"
	response := XDSAgentInfo{
		Tarballs: []XDSAgentTarball{
			XDSAgentTarball{
				OS:      "linux",
				FileURL: filepath.Join(tarballDir, "xds-agent_linux-amd64-v0.0.1_3cdf92c.zip"),
			},
			XDSAgentTarball{
				OS:      "windows",
				FileURL: filepath.Join(tarballDir, "xds-agent_windows-386-v0.0.1_3cdf92c.zip"),
			},
		},
	}

	c.JSON(http.StatusOK, response)
}
