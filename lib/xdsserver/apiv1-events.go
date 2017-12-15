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
	"net/http"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
)

// eventsList Registering for events that will be send over a WS
func (s *APIService) eventsList(c *gin.Context) {
	c.JSON(http.StatusOK, s.events.GetList())
}

// eventsRegister Registering for events that will be send over a WS
func (s *APIService) eventsRegister(c *gin.Context) {
	var args xsapiv1.EventRegisterArgs

	if c.BindJSON(&args) != nil || args.Name == "" {
		common.APIError(c, "Invalid arguments")
		return
	}

	sess := s.sessions.Get(c)
	if sess == nil {
		common.APIError(c, "Unknown sessions")
		return
	}

	// Register to all or to a specific events
	if err := s.events.Register(args.Name, sess.ID); err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// eventsRegister Registering for events that will be send over a WS
func (s *APIService) eventsUnRegister(c *gin.Context) {
	var args xsapiv1.EventUnRegisterArgs

	if c.BindJSON(&args) != nil || args.Name == "" {
		common.APIError(c, "Invalid arguments")
		return
	}

	sess := s.sessions.Get(c)
	if sess == nil {
		common.APIError(c, "Unknown sessions")
		return
	}

	// Register to all or to a specific events
	if err := s.events.UnRegister(args.Name, sess.ID); err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
