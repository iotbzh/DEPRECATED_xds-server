package xdsserver

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
)

var confMut sync.Mutex

// GetConfig returns server configuration
func (s *APIService) getConfig(c *gin.Context) {
	confMut.Lock()
	defer confMut.Unlock()

	c.JSON(http.StatusOK, s.Config)
}

// SetConfig sets server configuration
func (s *APIService) setConfig(c *gin.Context) {
	// FIXME - must be tested
	c.JSON(http.StatusNotImplemented, "Not implemented")

	var cfgArg xsapiv1.APIConfig

	if c.BindJSON(&cfgArg) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	confMut.Lock()
	defer confMut.Unlock()

	s.Log.Debugln("SET config: ", cfgArg)

	common.APIError(c, "Not Supported")
}
