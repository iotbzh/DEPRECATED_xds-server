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
	"errors"
	"net"

	"github.com/iotbzh/xds-server/lib/xsapiv1"
)

// NewBuilderConfig creates a new BuilderConfig instance
func NewBuilderConfig(stID string) (xsapiv1.BuilderConfig, error) {
	// Do we really need it ? may be not accessible from client side
	ip, err := getLocalIP()
	if err != nil {
		return xsapiv1.BuilderConfig{}, err
	}

	b := xsapiv1.BuilderConfig{
		IP:          ip, // TODO currently not used
		Port:        "", // TODO currently not used
		SyncThingID: stID,
	}
	return b, nil
}

/*** Private ***/

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("Cannot determined local IP")
}
