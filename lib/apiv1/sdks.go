package apiv1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-server/lib/common"
)

// getSdks returns all SDKs configuration
func (s *APIService) getSdks(c *gin.Context) {
	c.JSON(http.StatusOK, s.sdks.GetAll())
}

// getSdk returns a specific Sdk configuration
func (s *APIService) getSdk(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.APIError(c, "Invalid id")
		return
	}

	sdk := s.sdks.Get(id)
	if sdk.Profile == "" {
		common.APIError(c, "Invalid id")
		return
	}

	c.JSON(http.StatusOK, sdk)
}
