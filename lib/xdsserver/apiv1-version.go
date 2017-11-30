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
	"github.com/iotbzh/xds-server/lib/xsapiv1"
)

// getInfo : return various information about server
func (s *APIService) getVersion(c *gin.Context) {
	response := xsapiv1.Version{
		ID:            s.Config.ServerUID,
		Version:       s.Config.Version,
		APIVersion:    s.Config.APIVersion,
		VersionGitTag: s.Config.VersionGitTag,
	}

	c.JSON(http.StatusOK, response)
}
