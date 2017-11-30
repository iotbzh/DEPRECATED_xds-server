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
	"sync"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
)

var confMut sync.Mutex

// GetConfig returns server configuration
func (s *APIService) getConfig(c *gin.Context) {
	confMut.Lock()
	defer confMut.Unlock()

	c.JSON(http.StatusOK, s.Config)
}

// SetConfig sets server configuration
func (s *APIService) setConfig(c *gin.Context) {
	// FIXME - must be tested
	c.JSON(http.StatusNotImplemented, "Not implemented")

	var cfgArg xsapiv1.APIConfig

	if c.BindJSON(&cfgArg) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	confMut.Lock()
	defer confMut.Unlock()

	s.Log.Debugln("SET config: ", cfgArg)

	common.APIError(c, "Not Supported")
}
