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

// APIConfig parameters (json format) of /config command
type APIConfig struct {
	ServerUID        string          `json:"id"`
	Version          string          `json:"version"`
	APIVersion       string          `json:"apiVersion"`
	VersionGitTag    string          `json:"gitTag"`
	SupportedSharing map[string]bool `json:"supportedSharing"`
	Builder          BuilderConfig   `json:"builder"`
}

// BuilderConfig represents the builder container configuration
type BuilderConfig struct {
	IP          string `json:"ip"`
	Port        string `json:"port"`
	SyncThingID string `json:"syncThingID"`
}
