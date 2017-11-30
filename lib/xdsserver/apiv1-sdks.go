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
