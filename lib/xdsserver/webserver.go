/*
 * Copyright (C) 2017 "IoT.bzh"
 * Author Sebastien Douheret <sebastien@iot.bzh>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package xdsserver

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
)

// WebServer .
type WebServer struct {
	*Context
	router    *gin.Engine
	api       *APIService
	sIOServer *socketio.Server
	webApp    *gin.RouterGroup
	stop      chan struct{} // signals intentional stop
}

const indexFilename = "index.html"

// NewWebServer creates an instance of WebServer
func NewWebServer(ctx *Context) *WebServer {

	// Setup logging for gin router
	if ctx.Log.Level == logrus.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Redirect gin logs into another logger (LogVerboseOut may be stderr or a file)
	gin.DefaultWriter = ctx.Config.LogVerboseOut
	gin.DefaultErrorWriter = ctx.Config.LogVerboseOut
	log.SetOutput(ctx.Config.LogVerboseOut)

	// FIXME - fix pb about isTerminal=false when out is in VSC Debug Console

	// Creates gin router
	r := gin.New()

	svr := &WebServer{
		Context:   ctx,
		router:    r,
		api:       nil,
		sIOServer: nil,
		webApp:    nil,
		stop:      make(chan struct{}),
	}

	return svr
}

// Serve starts a new instance of the Web Server
func (s *WebServer) Serve() error {
	var err error

	// Setup middlewares
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())
	s.router.Use(s.middlewareXDSDetails())
	s.router.Use(s.middlewareCORS())

	// Create REST API
	s.api = NewAPIV1(s.Context)

	// Websocket routes
	s.sIOServer, err = socketio.NewServer(nil)
	if err != nil {
		s.Log.Fatalln(err)
	}

	s.router.GET("/socket.io/", s.socketHandler)
	s.router.POST("/socket.io/", s.socketHandler)
	/* TODO: do we want to support ws://...  ?
	s.router.Handle("WS", "/socket.io/", s.socketHandler)
	s.router.Handle("WSS", "/socket.io/", s.socketHandler)
	*/

	// Web Application (serve on / )
	idxFile := path.Join(s.Config.FileConf.WebAppDir, indexFilename)
	if _, err := os.Stat(idxFile); err != nil {
		s.Log.Fatalln("Web app directory not found, check/use webAppDir setting in config file: ", idxFile)
	}
	s.Log.Infof("Serve WEB app dir: %s", s.Config.FileConf.WebAppDir)
	s.router.Use(static.Serve("/", static.LocalFile(s.Config.FileConf.WebAppDir, true)))
	s.webApp = s.router.Group("/", s.serveIndexFile)
	{
		s.webApp.GET("/")
	}

	// Serve in the background
	serveError := make(chan error, 1)
	go func() {
		msg := fmt.Sprintf("Web Server running on localhost:%s ...\n", s.Config.FileConf.HTTPPort)
		s.Log.Infof(msg)
		fmt.Printf(msg)
		serveError <- http.ListenAndServe(":"+s.Config.FileConf.HTTPPort, s.router)
	}()

	// Wait for stop, restart or error signals
	select {
	case <-s.stop:
		// Shutting down permanently
		s.sessions.Stop()
		s.sdks.Stop()
		s.Log.Infoln("shutting down (stop)")
	case err = <-serveError:
		// Error due to listen/serve failure
		s.Log.Errorln(err)
	}

	return nil
}

// Stop web server
func (s *WebServer) Stop() {
	close(s.stop)
}

// serveIndexFile provides initial file (eg. index.html) of webapp
func (s *WebServer) serveIndexFile(c *gin.Context) {
	c.HTML(200, indexFilename, gin.H{})
}

// Add details in Header
func (s *WebServer) middlewareXDSDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("XDS-Version", s.Config.Version)
		c.Header("XDS-API-Version", s.Config.APIVersion)
		c.Next()
	}
}

// CORS middleware
func (s *WebServer) middlewareCORS() gin.HandlerFunc {
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
func (s *WebServer) socketHandler(c *gin.Context) {

	// Retrieve user session
	sess := s.sessions.Get(c)
	if sess == nil {
		c.JSON(500, gin.H{"error": "Cannot retrieve session"})
		return
	}

	s.sIOServer.On("connection", func(so socketio.Socket) {
		s.Log.Debugf("WS Connected (SID=%v)", so.Id())
		s.sessions.UpdateIOSocket(sess.ID, &so)

		so.On("disconnection", func() {
			s.Log.Debugf("WS disconnected (SID=%v)", so.Id())
			s.sessions.UpdateIOSocket(sess.ID, nil)
		})
	})

	s.sIOServer.On("error", func(so socketio.Socket, err error) {
		s.Log.Errorf("WS SID=%v Error : %v", so.Id(), err.Error())
	})

	s.sIOServer.ServeHTTP(c.Writer, c.Request)
}
