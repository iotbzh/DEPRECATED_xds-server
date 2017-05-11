package apiv1

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-server/lib/common"
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

	if err := s.cfg.UpdateAll(cfgArg); err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, s.cfg)
}
