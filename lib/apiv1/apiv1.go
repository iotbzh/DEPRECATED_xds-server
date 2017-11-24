package apiv1

import (
	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"

	"github.com/iotbzh/xds-server/lib/crosssdk"
	"github.com/iotbzh/xds-server/lib/model"
	"github.com/iotbzh/xds-server/lib/session"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
)

// APIService .
type APIService struct {
	router    *gin.Engine
	apiRouter *gin.RouterGroup
	sessions  *session.Sessions
	cfg       *xdsconfig.Config
	mfolders  *model.Folders
	sdks      *crosssdk.SDKs
	log       *logrus.Logger
}

// New creates a new instance of API service
func New(r *gin.Engine, sess *session.Sessions, cfg *xdsconfig.Config, mfolders *model.Folders, sdks *crosssdk.SDKs) *APIService {
	s := &APIService{
		router:    r,
		sessions:  sess,
		apiRouter: r.Group("/api/v1"),
		cfg:       cfg,
		mfolders:  mfolders,
		sdks:      sdks,
		log:       cfg.Log,
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
