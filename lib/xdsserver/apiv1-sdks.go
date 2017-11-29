package xdsserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
)

// getSdks returns all SDKs configuration
func (s *APIService) getSdks(c *gin.Context) {
	c.JSON(http.StatusOK, s.sdks.GetAll())
}

// getSdk returns a specific Sdk configuration
func (s *APIService) getSdk(c *gin.Context) {
	id, err := s.sdks.ResolveID(c.Param("id"))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}
	sdk := s.sdks.Get(id)
	if sdk.Profile == "" {
		common.APIError(c, "Invalid id")
		return
	}

	c.JSON(http.StatusOK, sdk)
}
