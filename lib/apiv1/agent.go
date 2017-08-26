package apiv1

import (
	"net/http"
	"path"
	"strings"

	"path/filepath"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
)

// XDSAgentTarball .
type XDSAgentTarball struct {
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	Version    string `json:"version"`
	RawVersion string `json:"raw-version"`
	FileURL    string `json:"fileUrl"`
}

// XDSAgentInfo .
type XDSAgentInfo struct {
	Tarballs []XDSAgentTarball `json:"tarballs"`
}

// getXdsAgentInfo : return various information about Xds Agent
func (s *APIService) getXdsAgentInfo(c *gin.Context) {

	res := XDSAgentInfo{}
	tarballURL := "assets/xds-agent-tarballs"
	tarballDir := filepath.Join(s.cfg.FileConf.WebAppDir, "assets", "xds-agent-tarballs")
	if common.Exists(tarballDir) {
		files, err := filepath.Glob(path.Join(tarballDir, "xds-agent_*.zip"))
		if err != nil {
			s.log.Debugf("Error while retrieving xds-agent tarballs: dir=%s, error=%v", tarballDir, err)
		}
		for _, ff := range files {
			file := filepath.Base(ff)
			// Assume that tarball name format is: xds-agent_OS-ARCH-RAWVERSION.zip
			fs := strings.TrimSuffix(strings.TrimPrefix(file, "xds-agent_"), ".zip")
			f := strings.Split(fs, "-")

			if len(f) >= 3 {
				vers := strings.Split(f[2], "_")
				ver := f[2]
				if len(vers) > 1 {
					ver = vers[0]
				}

				newT := XDSAgentTarball{
					OS:         f[0],
					Arch:       f[1],
					Version:    ver,
					RawVersion: f[2],
					FileURL:    filepath.Join(tarballURL, file),
				}

				s.log.Infof("Added XDS-Agent tarball: %s", file)
				res.Tarballs = append(res.Tarballs, newT)

			} else {
				s.log.Debugf("Error while retrieving xds-agent, decoding failure: file:%v", ff)
			}
		}
	}

	c.JSON(http.StatusOK, res)
}
