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

// EventRegisterArgs Parameters (json format) of /events/register command
type EventRegisterArgs struct {
	Name      string `json:"name"`
	ProjectID string `json:"filterProjectID"`
}

// EventUnRegisterArgs Parameters of /events/unregister command
type EventUnRegisterArgs struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// EventMsg Message send
type EventMsg struct {
	Time   string       `json:"time"`
	Type   string       `json:"type"`
	Folder FolderConfig `json:"folder"`
}

// EventEvent Event send in WS when an internal event (eg. Syncthing event is received)
const (
	// EventTypePrefix Used as event prefix
	EventTypePrefix = "event:" // following by event type

	// Supported Events type
	EVTAll               = EventTypePrefix + "all"
	EVTFolderChange      = EventTypePrefix + "folder-change"       // type EventMsg with Data type xsapiv1.???
	EVTFolderStateChange = EventTypePrefix + "folder-state-change" // type EventMsg with Data type xsapiv1.???
)
