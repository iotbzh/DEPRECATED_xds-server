package apiv1

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
)

var confMut sync.Mutex

// GetConfig returns server configuration
func (s *APIService) getConfig(c *gin.Context) {
	confMut.Lock()
	defer confMut.Unlock()

	c.JSON(http.StatusOK, s.cfg)
}

// SetConfig sets server configuration
func (s *APIService) setConfig(c *gin.Context) {
	// FIXME - must be tested
	c.JSON(http.StatusNotImplemented, "Not implemented")

	var cfgArg xdsconfig.Config

	if c.BindJSON(&cfgArg) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	confMut.Lock()
	defer confMut.Unlock()

	s.log.Debugln("SET config: ", cfgArg)

	common.APIError(c, "Not Supported")
}
