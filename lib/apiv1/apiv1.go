package apiv1

import (
	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"

	"github.com/iotbzh/xds-server/lib/session"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
)

// APIService .
type APIService struct {
	router    *gin.Engine
	apiRouter *gin.RouterGroup
	sessions  *session.Sessions
	cfg       xdsconfig.Config
	log       *logrus.Logger
}

// New creates a new instance of API service
func New(sess *session.Sessions, cfg xdsconfig.Config, r *gin.Engine) *APIService {
	s := &APIService{
		router:    r,
		sessions:  sess,
		apiRouter: r.Group("/api/v1"),
		cfg:       cfg,
		log:       cfg.Log,
	}

	s.apiRouter.GET("/version", s.getVersion)

	s.apiRouter.GET("/config", s.getConfig)
	s.apiRouter.POST("/config", s.setConfig)

	s.apiRouter.GET("/folders", s.getFolders)
	s.apiRouter.GET("/folder/:id", s.getFolder)
	s.apiRouter.POST("/folder", s.addFolder)
	s.apiRouter.DELETE("/folder/:id", s.delFolder)

	s.apiRouter.POST("/make", s.buildMake)
	s.apiRouter.POST("/make/:id", s.buildMake)

	/* TODO: to be tested and then enabled
	s.apiRouter.POST("/exec", s.execCmd)
	s.apiRouter.POST("/exec/:id", s.execCmd)
	*/

	return s
}
