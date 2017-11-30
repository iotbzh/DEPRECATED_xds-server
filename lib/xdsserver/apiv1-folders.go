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
	"os"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
)

// getFolders returns all folders configuration
func (s *APIService) getFolders(c *gin.Context) {
	c.JSON(http.StatusOK, s.mfolders.GetConfigArr())
}

// getFolder returns a specific folder configuration
func (s *APIService) getFolder(c *gin.Context) {
	id, err := s.mfolders.ResolveID(c.Param("id"))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}
	f := s.mfolders.Get(id)
	if f == nil {
		common.APIError(c, "Invalid id")
		return
	}

	c.JSON(http.StatusOK, (*f).GetConfig())
}

// addFolder adds a new folder to server config
func (s *APIService) addFolder(c *gin.Context) {
	var cfgArg xsapiv1.FolderConfig
	if c.BindJSON(&cfgArg) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	s.Log.Debugln("Add folder config: ", cfgArg)

	newFld, err := s.mfolders.Add(cfgArg)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	// Create xds-project.conf file
	// FIXME: move to folders.createUpdate func (but gin context needed)
	fld := s.mfolders.Get(newFld.ID)
	prjConfFile := (*fld).GetFullPath("xds-project.conf")
	if !common.Exists(prjConfFile) {
		fd, err := os.OpenFile(prjConfFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
		if err != nil {
			common.APIError(c, err.Error())
			return
		}
		fd.WriteString("# XDS project settings\n")
		fd.WriteString("export XDS_SERVER_URL=" + c.Request.Host + "\n")
		fd.WriteString("export XDS_PROJECT_ID=" + newFld.ID + "\n")
		if newFld.DefaultSdk == "" {
			sdks := s.sdks.GetAll()
			newFld.DefaultSdk = sdks[0].ID
		}
		fd.WriteString("export XDS_SDK_ID=" + newFld.DefaultSdk + "\n")
		fd.Close()
	}

	c.JSON(http.StatusOK, newFld)
}

// syncFolder force synchronization of folder files
func (s *APIService) syncFolder(c *gin.Context) {
	id, err := s.mfolders.ResolveID(c.Param("id"))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}
	s.Log.Debugln("Sync folder id: ", id)

	err = s.mfolders.ForceSync(id)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, "")
}

// delFolder deletes folder from server config
func (s *APIService) delFolder(c *gin.Context) {
	id, err := s.mfolders.ResolveID(c.Param("id"))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	s.Log.Debugln("Delete folder id ", id)

	delEntry, err := s.mfolders.Delete(id)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, delEntry)
}

// updateFolder update some field of a folder
func (s *APIService) updateFolder(c *gin.Context) {
	id, err := s.mfolders.ResolveID(c.Param("id"))
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	s.Log.Debugln("Update folder id ", id)

	var cfgArg xsapiv1.FolderConfig
	if c.BindJSON(&cfgArg) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	upFld, err := s.mfolders.Update(id, cfgArg)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, upFld)
}
