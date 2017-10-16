package webserver

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"path"

	"github.com/Sirupsen/logrus"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/googollee/go-socket.io"
	"github.com/iotbzh/xds-server/lib/apiv1"
	"github.com/iotbzh/xds-server/lib/crosssdk"
	"github.com/iotbzh/xds-server/lib/model"
	"github.com/iotbzh/xds-server/lib/session"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
)

// Server .
type Server struct {
	router    *gin.Engine
	api       *apiv1.APIService
	sIOServer *socketio.Server
	webApp    *gin.RouterGroup
	cfg       *xdsconfig.Config
	sessions  *session.Sessions
	mfolders  *model.Folders
	sdks      *crosssdk.SDKs
	log       *logrus.Logger
	sillyLog  bool
	stop      chan struct{} // signals intentional stop
}

const indexFilename = "index.html"
const cookieMaxAge = "3600"

// New creates an instance of Server
func New(cfg *xdsconfig.Config, mfolders *model.Folders, sdks *crosssdk.SDKs, logr *logrus.Logger, sillyLog bool) *Server {

	// Setup logging for gin router
	if logr.Level == logrus.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Redirect gin logs into another logger (LogVerboseOut may be stderr or a file)
	gin.DefaultWriter = cfg.LogVerboseOut
	gin.DefaultErrorWriter = cfg.LogVerboseOut
	log.SetOutput(cfg.LogVerboseOut)

	// FIXME - fix pb about isTerminal=false when out is in VSC Debug Console

	// Creates gin router
	r := gin.New()

	svr := &Server{
		router:    r,
		api:       nil,
		sIOServer: nil,
		webApp:    nil,
		cfg:       cfg,
		sessions:  nil,
		mfolders:  mfolders,
		sdks:      sdks,
		log:       logr,
		sillyLog:  sillyLog,
		stop:      make(chan struct{}),
	}

	return svr
}

// Serve starts a new instance of the Web Server
func (s *Server) Serve() error {
	var err error

	// Setup middlewares
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())
	s.router.Use(s.middlewareXDSDetails())
	s.router.Use(s.middlewareCORS())

	// Sessions manager
	s.sessions = session.NewClientSessions(s.router, s.log, cookieMaxAge, s.sillyLog)

	// Create REST API
	s.api = apiv1.New(s.router, s.sessions, s.cfg, s.mfolders, s.sdks)

	// Websocket routes
	s.sIOServer, err = socketio.NewServer(nil)
	if err != nil {
		s.log.Fatalln(err)
	}

	s.router.GET("/socket.io/", s.socketHandler)
	s.router.POST("/socket.io/", s.socketHandler)
	/* TODO: do we want to support ws://...  ?
	s.router.Handle("WS", "/socket.io/", s.socketHandler)
	s.router.Handle("WSS", "/socket.io/", s.socketHandler)
	*/

	// Web Application (serve on / )
	idxFile := path.Join(s.cfg.FileConf.WebAppDir, indexFilename)
	if _, err := os.Stat(idxFile); err != nil {
		s.log.Fatalln("Web app directory not found, check/use webAppDir setting in config file: ", idxFile)
	}
	s.log.Infof("Serve WEB app dir: %s", s.cfg.FileConf.WebAppDir)
	s.router.Use(static.Serve("/", static.LocalFile(s.cfg.FileConf.WebAppDir, true)))
	s.webApp = s.router.Group("/", s.serveIndexFile)
	{
		s.webApp.GET("/")
	}

	// Serve in the background
	serveError := make(chan error, 1)
	go func() {
		msg := fmt.Sprintf("Web Server running on localhost:%s ...\n", s.cfg.FileConf.HTTPPort)
		s.log.Infof(msg)
		fmt.Printf(msg)
		serveError <- http.ListenAndServe(":"+s.cfg.FileConf.HTTPPort, s.router)
	}()

	// Wait for stop, restart or error signals
	select {
	case <-s.stop:
		// Shutting down permanently
		s.sessions.Stop()
		s.log.Infoln("shutting down (stop)")
	case err = <-serveError:
		// Error due to listen/serve failure
		s.log.Errorln(err)
	}

	return nil
}

// Stop web server
func (s *Server) Stop() {
	close(s.stop)
}

// serveIndexFile provides initial file (eg. index.html) of webapp
func (s *Server) serveIndexFile(c *gin.Context) {
	c.HTML(200, indexFilename, gin.H{})
}

// Add details in Header
func (s *Server) middlewareXDSDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("XDS-Version", s.cfg.Version)
		c.Header("XDS-API-Version", s.cfg.APIVersion)
		c.Next()
	}
}

// CORS middleware
func (s *Server) middlewareCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE")
			c.Header("Access-Control-Max-Age", cookieMaxAge)
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// socketHandler is the handler for the "main" websocket connection
func (s *Server) socketHandler(c *gin.Context) {

	// Retrieve user session
	sess := s.sessions.Get(c)
	if sess == nil {
		c.JSON(500, gin.H{"error": "Cannot retrieve session"})
		return
	}

	s.sIOServer.On("connection", func(so socketio.Socket) {
		s.log.Debugf("WS Connected (SID=%v)", so.Id())
		s.sessions.UpdateIOSocket(sess.ID, &so)

		so.On("disconnection", func() {
			s.log.Debugf("WS disconnected (SID=%v)", so.Id())
			s.sessions.UpdateIOSocket(sess.ID, nil)
		})
	})

	s.sIOServer.On("error", func(so socketio.Socket, err error) {
		s.log.Errorf("WS SID=%v Error : %v", so.Id(), err.Error())
	})

	s.sIOServer.ServeHTTP(c.Writer, c.Request)
}
