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
	"github.com/iotbzh/xds-server/lib/xsapiv1"
	uuid "github.com/satori/go.uuid"
)

// IFOLDER interface implementation for disabled Syncthing folders
// It's a "fallback" interface used to keep syncthing folders config even
// when syncthing is not running.

// STFolderDisable .
type STFolderDisable struct {
	*Context
	config xsapiv1.FolderConfig
}

// NewFolderSTDisable Create a new instance of STFolderDisable
func NewFolderSTDisable(ctx *Context) *STFolderDisable {
	f := STFolderDisable{
		Context: ctx,
	}
	return &f
}

// NewUID Get a UUID
func (f *STFolderDisable) NewUID(suffix string) string {
	uuid := uuid.NewV1().String()
	if len(suffix) > 0 {
		uuid += "_" + suffix
	}
	return uuid
}

// Add a new folder
func (f *STFolderDisable) Add(cfg xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error) {
	f.config = cfg
	f.config.Status = xsapiv1.StatusDisable
	f.config.IsInSync = false
	return &f.config, nil
}

// GetConfig Get public part of folder config
func (f *STFolderDisable) GetConfig() xsapiv1.FolderConfig {
	return f.config
}

// GetFullPath returns the full path of a directory (from server POV)
func (f *STFolderDisable) GetFullPath(dir string) string {
	return ""
}

// ConvPathCli2Svr Convert path from Client to Server
func (f *STFolderDisable) ConvPathCli2Svr(s string) string {
	return ""
}

// ConvPathSvr2Cli Convert path from Server to Client
func (f *STFolderDisable) ConvPathSvr2Cli(s string) string {
	return ""
}

// Remove a folder
func (f *STFolderDisable) Remove() error {
	return nil
}

// Update update some fields of a folder
func (f *STFolderDisable) Update(cfg xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error) {
	return nil, nil
}

// RegisterEventChange requests registration for folder change event
func (f *STFolderDisable) RegisterEventChange(cb *FolderEventCB, data *FolderEventCBData) error {
	return nil
}

// UnRegisterEventChange remove registered callback
func (f *STFolderDisable) UnRegisterEventChange() error {
	return nil
}

// Sync Force folder files synchronization
func (f *STFolderDisable) Sync() error {
	return nil
}

// IsInSync Check if folder files are in-sync
func (f *STFolderDisable) IsInSync() (bool, error) {
	return false, nil
}
