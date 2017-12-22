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

// SDK status definition
const (
	SdkStatusDisable      = "Disable"
	SdkStatusNotInstalled = "Not Installed"
	SdkStatusInstalling   = "Installing"
	SdkStatusUninstalling = "Un-installing"
	SdkStatusInstalled    = "Installed"
)

// SDK Define a cross tool chain used to build application
type SDK struct {
	ID          string `json:"id" binding:"required"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Profile     string `json:"profile"`
	Version     string `json:"version"`
	Arch        string `json:"arch"`
	Path        string `json:"path"`
	URL         string `json:"url"`
	Status      string `json:"status"`
	Date        string `json:"date"`
	Size        string `json:"size"`
	Md5sum      string `json:"md5sum"`
	SetupFile   string `json:"setupFile"`
	LastError   string `json:"lastError"`

	// Not exported fields
	FamilyConf SDKFamilyConfig `json:"-"`
}

// SDKFamilyConfig Configuration structure to define a SDKs family
type SDKFamilyConfig struct {
	FamilyName   string `json:"familyName"`
	Description  string `json:"description"`
	RootDir      string `json:"rootDir"`
	EnvSetupFile string `json:"envSetupFilename"`
	ScriptsDir   string `json:"scriptsDir"`
}

// SDKInstallArgs JSON parameters of POST /sdks or /sdks/abortinstall commands
type SDKInstallArgs struct {
	ID       string `json:"id" binding:"required"` // install by ID (must be part of GET /sdks result)
	Filename string `json:"filename"`              // install by using a file
	Force    bool   `json:"force"`                 // force SDK install when already existing
	Timeout  int    `json:"timeout"`               // 1800 == default 30 minutes
}

// SDKManagementMsg Message send during SDK installation or when installation is complete
type SDKManagementMsg struct {
	CmdID     string `json:"cmdID"`
	Timestamp string `json:"timestamp"`
	Sdk       SDK    `json:"sdk"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	Progress  int    `json:"progress"` // 0 = not started to 100% = complete
	Exited    bool   `json:"exited"`
	Code      int    `json:"code"`
	Error     string `json:"error"`
}
