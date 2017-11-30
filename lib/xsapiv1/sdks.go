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

// SDK Define a cross tool chain used to build application
type SDK struct {
	ID      string `json:"id" binding:"required"`
	Name    string `json:"name"`
	Profile string `json:"profile"`
	Version string `json:"version"`
	Arch    string `json:"arch"`
	Path    string `json:"path"`

	// Not exported fields
	EnvFile string `json:"-"`
}
