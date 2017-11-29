package xdsserver

import (
	"github.com/gin-gonic/gin"
)

// APIService .
type APIService struct {
	*Context
	apiRouter *gin.RouterGroup
}

// NewAPIV1 creates a new instance of API service
func NewAPIV1(ctx *Context) *APIService {
	s := &APIService{
		Context:   ctx,
		apiRouter: ctx.WWWServer.router.Group("/api/v1"),
	}

	s.apiRouter.GET("/version", s.getVersion)

	s.apiRouter.GET("/config", s.getConfig)
	s.apiRouter.POST("/config", s.setConfig)

	s.apiRouter.GET("/folders", s.getFolders)
	s.apiRouter.GET("/folders/:id", s.getFolder)
	s.apiRouter.PUT("/folders/:id", s.updateFolder)
	s.apiRouter.POST("/folders", s.addFolder)
	s.apiRouter.POST("/folders/sync/:id", s.syncFolder)
	s.apiRouter.DELETE("/folders/:id", s.delFolder)

	s.apiRouter.GET("/sdks", s.getSdks)
	s.apiRouter.GET("/sdks/:id", s.getSdk)

	s.apiRouter.POST("/make", s.buildMake)
	s.apiRouter.POST("/make/:id", s.buildMake)

	s.apiRouter.POST("/exec", s.execCmd)
	s.apiRouter.POST("/exec/:id", s.execCmd)
	s.apiRouter.POST("/signal", s.execSignalCmd)

	s.apiRouter.GET("/events", s.eventsList)
	s.apiRouter.POST("/events/register", s.eventsRegister)
	s.apiRouter.POST("/events/unregister", s.eventsUnRegister)

	return s
}
