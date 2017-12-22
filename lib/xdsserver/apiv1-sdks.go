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

// installSdk Install a new Sdk
func (s *APIService) installSdk(c *gin.Context) {
	var args xsapiv1.SDKInstallArgs

	if err := c.BindJSON(&args); err != nil {
		common.APIError(c, "Invalid arguments")
		return
	}
	id, err := s.sdks.ResolveID(args.ID)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	// Support install from ID->URL or from local file
	if id != "" {
		s.Log.Debugf("Installing SDK id %s (force %v)", id, args.Force)
	} else if args.Filename != "" {
		s.Log.Debugf("Installing SDK filename %s (force %v)", args.Filename, args.Force)
	}

	// Retrieve session info
	sess := s.sessions.Get(c)
	if sess == nil {
		common.APIError(c, "Unknown sessions")
		return
	}

	sdk, err := s.sdks.Install(id, args.Filename, args.Force, args.Timeout, sess)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, sdk)
}

// abortInstallSdk Abort a SDK installation
func (s *APIService) abortInstallSdk(c *gin.Context) {
	var args xsapiv1.SDKInstallArgs

	if err := c.BindJSON(&args); err != nil {
		common.APIError(c, "Invalid arguments")
		return
	}
	id, err := s.sdks.ResolveID(args.ID)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	sdk, err := s.sdks.AbortInstall(id, args.Timeout)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, sdk)
}

// removeSdk Uninstall a Sdk
func (s *APIService) removeSdk(c *gin.Context) {
	id, err := s.sdks.ResolveID(c.Param("id"))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	// Retrieve session info
	sess := s.sessions.Get(c)
	if sess == nil {
		common.APIError(c, "Unknown sessions")
		return
	}

	s.Log.Debugln("Remove SDK id ", id)

	delEntry, err := s.sdks.Remove(id, -1, sess)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, delEntry)
}
