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

package xdsconfig

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	common "github.com/iotbzh/xds-common/golib"
	uuid "github.com/satori/go.uuid"
	"github.com/syncthing/syncthing/lib/sync"
)

// xmlServerData contains persistent data stored/loaded by server
type xmlServerData struct {
	XMLName xml.Name   `xml:"XDS-Server"`
	Version string     `xml:"version,attr"`
	Data    ServerData `xml:"server-data"`
}

type ServerData struct {
	ID string `xml:"id"`
}

var sdMutex = sync.NewMutex()

// ServerIDGet
func ServerIDGet() (string, error) {
	var f string
	var err error

	d := ServerData{}
	if f, err = ServerDataFilenameGet(); err != nil {
		return "", err
	}
	if err = serverDataRead(f, &d); err != nil || d.ID == "" {
		// Create a new uuid when not found
		d.ID = uuid.NewV1().String()
		if err := serverDataWrite(f, d); err != nil {
			return "", err
		}
	}
	return d.ID, nil
}

// serverDataRead reads data saved on disk
func serverDataRead(file string, data *ServerData) error {
	if !common.Exists(file) {
		return fmt.Errorf("No folder config file found (%s)", file)
	}

	sdMutex.Lock()
	defer sdMutex.Unlock()

	fd, err := os.Open(file)
	defer fd.Close()
	if err != nil {
		return err
	}

	xsd := xmlServerData{}
	err = xml.NewDecoder(fd).Decode(&xsd)
	if err == nil {
		*data = xsd.Data
	}
	return err
}

// serverDataWrite writes persistant data to disk
func serverDataWrite(file string, data ServerData) error {
	sdMutex.Lock()
	defer sdMutex.Unlock()

	dir := filepath.Dir(file)
	if !common.Exists(dir) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("Cannot create server data directory: %s", dir)
		}
	}

	fd, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	defer fd.Close()
	if err != nil {
		return err
	}

	xsd := &xmlServerData{
		Version: "1",
		Data:    data,
	}

	enc := xml.NewEncoder(fd)
	enc.Indent("", "  ")
	return enc.Encode(xsd)
}
