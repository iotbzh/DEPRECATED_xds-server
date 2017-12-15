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

import "github.com/iotbzh/xds-server/lib/xsapiv1"

type FolderEventCBData map[string]interface{}
type FolderEventCB func(cfg *xsapiv1.FolderConfig, data *FolderEventCBData)

// IFOLDER Folder interface
type IFOLDER interface {
	NewUID(suffix string) string                                    // Get a new folder UUID
	Add(cfg xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error)    // Add a new folder
	Setup(prj xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error)  // Local setup of the folder
	GetConfig() xsapiv1.FolderConfig                                // Get folder public configuration
	GetFullPath(dir string) string                                  // Get folder full path
	ConvPathCli2Svr(s string) string                                // Convert path from Client to Server
	ConvPathSvr2Cli(s string) string                                // Convert path from Server to Client
	Remove() error                                                  // Remove a folder
	Update(cfg xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error) // Update a new folder
	Sync() error                                                    // Force folder files synchronization
	IsInSync() (bool, error)                                        // Check if folder files are in-sync
}
