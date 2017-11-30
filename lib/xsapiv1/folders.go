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

package xsapiv1

// FolderType definition
type FolderType string

const (
	TypePathMap   = "PathMap"
	TypeCloudSync = "CloudSync"
	TypeCifsSmb   = "CIFS"
)

// Folder Status definition
const (
	StatusErrorConfig = "ErrorConfig"
	StatusDisable     = "Disable"
	StatusEnable      = "Enable"
	StatusPause       = "Pause"
	StatusSyncing     = "Syncing"
)

// FolderConfig is the config for one folder
type FolderConfig struct {
	ID         string     `json:"id"`
	Label      string     `json:"label"`
	ClientPath string     `json:"path"`
	Type       FolderType `json:"type"`
	Status     string     `json:"status"`
	IsInSync   bool       `json:"isInSync"`
	DefaultSdk string     `json:"defaultSdk"`
	ClientData string     `json:"clientData"` // free form field that can used by client

	// Not exported fields from REST API point of view
	RootPath string `json:"-"`

	// FIXME: better to define an equivalent to union data and then implement
	// UnmarshalJSON/MarshalJSON to decode/encode according to Type value
	// Data interface{} `json:"data"`

	// Specific data depending on which Type is used
	DataPathMap   PathMapConfig   `json:"dataPathMap,omitempty"`
	DataCloudSync CloudSyncConfig `json:"dataCloudSync,omitempty"`
}

// FolderConfigUpdatableFields List fields that can be updated using Update function
var FolderConfigUpdatableFields = []string{
	"Label", "DefaultSdk", "ClientData",
}

// PathMapConfig Path mapping specific data
type PathMapConfig struct {
	ServerPath string `json:"serverPath"`

	// Don't keep temporary file name (IOW we don't want to save it and reuse it)
	CheckFile    string `json:"checkFile" xml:"-"`
	CheckContent string `json:"checkContent" xml:"-"`
}

// CloudSyncConfig CloudSync (AKA Syncthing) specific data
type CloudSyncConfig struct {
	SyncThingID string `json:"syncThingID"`

	// Not exported fields (only used internally)
	STSvrStatus   string `json:"-"`
	STSvrIsInSync bool   `json:"-"`
	STLocStatus   string `json:"-"`
	STLocIsInSync bool   `json:"-"`
}
