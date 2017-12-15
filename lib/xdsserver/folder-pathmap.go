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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
	uuid "github.com/satori/go.uuid"
)

// IFOLDER interface implementation for native/path mapping folders

// PathMap .
type PathMap struct {
	*Context
	fConfig xsapiv1.FolderConfig
}

// NewFolderPathMap Create a new instance of PathMap
func NewFolderPathMap(ctx *Context) *PathMap {
	f := PathMap{
		Context: ctx,
		fConfig: xsapiv1.FolderConfig{
			Status: xsapiv1.StatusDisable,
		},
	}
	return &f
}

// NewUID Get a UUID
func (f *PathMap) NewUID(suffix string) string {
	uuid := uuid.NewV1().String()
	if len(suffix) > 0 {
		uuid += "_" + suffix
	}
	return uuid
}

// Add a new folder
func (f *PathMap) Add(cfg xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error) {
	return f.Setup(cfg)
}

// Setup Setup local project config
func (f *PathMap) Setup(cfg xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error) {

	if cfg.DataPathMap.ServerPath == "" {
		return nil, fmt.Errorf("ServerPath must be set")
	}

	// Use shareRootDir if ServerPath is a relative path
	dir := cfg.DataPathMap.ServerPath
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(f.Config.FileConf.ShareRootDir, dir)
	}

	// Sanity check
	if !common.Exists(dir) {
		// try to create if not existing
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("Cannot create ServerPath directory: %s", dir)
		}
	}
	if !common.Exists(dir) {
		return nil, fmt.Errorf("ServerPath directory is not accessible: %s", dir)
	}

	f.fConfig = cfg
	f.fConfig.RootPath = dir
	f.fConfig.DataPathMap.ServerPath = dir
	f.fConfig.IsInSync = true

	// Verify file created by XDS agent when needed
	if cfg.DataPathMap.CheckFile != "" {
		errMsg := "ServerPath sanity check error (%d): %v"
		ckFile := f.ConvPathCli2Svr(cfg.DataPathMap.CheckFile)
		if !common.Exists(ckFile) {
			return nil, fmt.Errorf(errMsg, 1, "file not present")
		}
		if cfg.DataPathMap.CheckContent != "" {
			fd, err := os.OpenFile(ckFile, os.O_APPEND|os.O_RDWR, 0600)
			if err != nil {
				return nil, fmt.Errorf(errMsg, 2, err)
			}
			defer fd.Close()

			// Check specific message written by agent
			content, err := ioutil.ReadAll(fd)
			if err != nil {
				return nil, fmt.Errorf(errMsg, 3, err)
			}
			if string(content) != cfg.DataPathMap.CheckContent {
				return nil, fmt.Errorf(errMsg, 4, "file content differ")
			}

			// Write a specific message that will be check back on agent side
			msg := "Pathmap checked message written by xds-server ID: " + f.Config.ServerUID + "\n"
			if n, err := fd.WriteString(msg); n != len(msg) || err != nil {
				return nil, fmt.Errorf(errMsg, 5, err)
			}
		}
	}

	f.fConfig.Status = xsapiv1.StatusEnable

	return &f.fConfig, nil
}

// GetConfig Get public part of folder config
func (f *PathMap) GetConfig() xsapiv1.FolderConfig {
	return f.fConfig
}

// GetFullPath returns the full path of a directory (from server POV)
func (f *PathMap) GetFullPath(dir string) string {
	if &dir == nil {
		return f.fConfig.DataPathMap.ServerPath
	}
	return filepath.Join(f.fConfig.DataPathMap.ServerPath, dir)
}

// ConvPathCli2Svr Convert path from Client to Server
func (f *PathMap) ConvPathCli2Svr(s string) string {
	if f.fConfig.ClientPath != "" && f.fConfig.DataPathMap.ServerPath != "" {
		return strings.Replace(s,
			f.fConfig.ClientPath,
			f.fConfig.DataPathMap.ServerPath,
			-1)
	}
	return s
}

// ConvPathSvr2Cli Convert path from Server to Client
func (f *PathMap) ConvPathSvr2Cli(s string) string {
	if f.fConfig.ClientPath != "" && f.fConfig.DataPathMap.ServerPath != "" {
		return strings.Replace(s,
			f.fConfig.DataPathMap.ServerPath,
			f.fConfig.ClientPath,
			-1)
	}
	return s
}

// Remove a folder
func (f *PathMap) Remove() error {
	// nothing to do
	return nil
}

// Update update some fields of a folder
func (f *PathMap) Update(cfg xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error) {
	if f.fConfig.ID != cfg.ID {
		return nil, fmt.Errorf("Invalid id")
	}
	f.fConfig = cfg
	return &f.fConfig, nil
}

// Sync Force folder files synchronization
func (f *PathMap) Sync() error {
	return nil
}

// IsInSync Check if folder files are in-sync
func (f *PathMap) IsInSync() (bool, error) {
	return true, nil
}
