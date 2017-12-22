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
	s.apiRouter.POST("/sdks", s.installSdk)
	s.apiRouter.POST("/sdks/abortinstall", s.abortInstallSdk)
	s.apiRouter.DELETE("/sdks/:id", s.removeSdk)

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
