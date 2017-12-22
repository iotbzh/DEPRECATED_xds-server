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
	"path"
	"path/filepath"
	"strings"
	"sync"

	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
	uuid "github.com/satori/go.uuid"
)

// SDKs List of installed SDK
type SDKs struct {
	*Context
	Sdks map[string]*CrossSDK

	mutex sync.Mutex
	stop  chan struct{} // signals intentional stop
}

// NewSDKs creates a new instance of SDKs
func NewSDKs(ctx *Context) (*SDKs, error) {
	s := SDKs{
		Context: ctx,
		Sdks:    make(map[string]*CrossSDK),
		stop:    make(chan struct{}),
	}

	scriptsDir := ctx.Config.FileConf.SdkScriptsDir
	if !common.Exists(scriptsDir) {
		// allow to use scripts/sdk in debug mode
		scriptsDir = filepath.Join(filepath.Dir(ctx.Config.FileConf.SdkScriptsDir), "scripts", "sdks")
		if !common.Exists(scriptsDir) {
			return &s, fmt.Errorf("scripts directory doesn't exist (%v)", scriptsDir)
		}
	}
	s.Log.Infof("SDK scripts dir: %s", scriptsDir)

	dirs, err := filepath.Glob(path.Join(scriptsDir, "*"))
	if err != nil {
		s.Log.Errorf("Error while retrieving SDK scripts: dir=%s, error=%s", scriptsDir, err.Error())
		return &s, err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Foreach directories in scripts/sdk
	nbInstalled := 0
	monSdksPath := make(map[string]*xsapiv1.SDKFamilyConfig)
	for _, d := range dirs {
		if !common.IsDir(d) {
			continue
		}

		sdksList, err := ListCrossSDK(d, s.Log)
		if err != nil {
			return &s, err
		}
		s.LogSillyf("'%s' SDKs list: %v", d, sdksList)

		for _, sdk := range sdksList {
			cSdk, err := NewCrossSDK(ctx, sdk, d)
			if err != nil {
				s.Log.Debugf("Error while processing SDK sdk=%v\n err=%s", sdk, err.Error())
				continue
			}
			if _, exist := s.Sdks[cSdk.sdk.ID]; exist {
				s.Log.Warningf("Duplicate SDK ID : %v", cSdk.sdk.ID)
				cSdk.sdk.ID += "_DUPLICATE_" + uuid.NewV1().String()
			}
			s.Sdks[cSdk.sdk.ID] = cSdk
			if cSdk.sdk.Status == xsapiv1.SdkStatusInstalled {
				nbInstalled++
			}

			monSdksPath[cSdk.sdk.FamilyConf.RootDir] = &cSdk.sdk.FamilyConf
		}
	}

	ctx.Log.Debugf("Cross SDKs: %d defined, %d installed", len(s.Sdks), nbInstalled)

	// Start monitor thread to detect new SDKs
	if len(monSdksPath) == 0 {
		s.Log.Warningf("No cross SDKs definition found")
	}

	return &s, nil
}

// Stop SDKs management
func (s *SDKs) Stop() {
	close(s.stop)
}

// monitorSDKInstallation
/* TODO: cleanup
func (s *SDKs) monitorSDKInstallation(monSDKs map[string]*xsapiv1.SDKFamilyConfig) {

	// Set up a watchpoint listening for inotify-specific events
	c := make(chan notify.EventInfo, 1)

	addWatcher := func(rootDir string) error {
		s.Log.Debugf("SDK Register watcher: rootDir=%s", rootDir)

		if err := notify.Watch(rootDir+"/...", c, notify.Create, notify.Remove); err != nil {
			return fmt.Errorf("SDK monitor: rootDir=%v err=%v", rootDir, err)
		}
		return nil
	}

	// Add directory watchers
	for dir := range monSDKs {
		if err := addWatcher(dir); err != nil {
			s.Log.Errorln(err.Error())
		}
	}

	// Wait inotify or stop events
	for {
		select {
		case <-s.stop:
			s.Log.Debugln("Stop monitorSDKInstallation")
			notify.Stop(c)
			return
		case ei := <-c:
			s.LogSillyf("monitorSDKInstallation SDKs event %v, path %v\n", ei.Event(), ei.Path())

			// Filter out all event that doesn't match environment file
			if !strings.Contains(ei.Path(), "environment-setup-") {
				continue
			}
			dir := path.Dir(ei.Path())

			sdk, err := s.GetByPath(dir)
			if err != nil {
				s.Log.Warningf("Cannot find SDK path to notify creation")
				s.LogSillyf("event: %v", ei.Event())
				continue
			}

			switch ei.Event() {
			case notify.Create:
				// Emit Folder state change event
				if err := s.events.Emit(xsapiv1.EVTSDKInstall, sdk, ""); err != nil {
					s.Log.Warningf("Cannot notify SDK install: %v", err)
				}

			case notify.Remove, notify.InMovedFrom:
				// Emit Folder state change event
				if err := s.events.Emit(xsapiv1.EVTSDKRemove, sdk, ""); err != nil {
					s.Log.Warningf("Cannot notify SDK remove: %v", err)
				}
			}
		}
	}
}
*/

// ResolveID Complete an SDK ID (helper for user that can use partial ID value)
func (s *SDKs) ResolveID(id string) (string, error) {
	if id == "" {
		return "", nil
	}

	match := []string{}
	for iid := range s.Sdks {
		if strings.HasPrefix(iid, id) {
			match = append(match, iid)
		}
	}

	if len(match) == 1 {
		return match[0], nil
	} else if len(match) == 0 {
		return id, fmt.Errorf("Unknown sdk id")
	}
	return id, fmt.Errorf("Multiple sdk IDs found with provided prefix: " + id)
}

// Get returns an SDK from id
func (s *SDKs) Get(id string) *xsapiv1.SDK {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sc, exist := s.Sdks[id]
	if !exist {
		return nil
	}
	return (*sc).Get()
}

// GetByPath Find a SDK from path
func (s *SDKs) GetByPath(path string) (*xsapiv1.SDK, error) {
	if path == "" {
		return nil, fmt.Errorf("can't found sdk (empty path)")
	}
	for _, ss := range s.Sdks {
		if ss.sdk.Path == path {
			return ss.Get(), nil
		}
	}
	return nil, fmt.Errorf("not found")
}

// GetAll returns all existing SDKs
func (s *SDKs) GetAll() []xsapiv1.SDK {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	res := []xsapiv1.SDK{}
	for _, v := range s.Sdks {
		res = append(res, *(*v).Get())
	}
	return res
}

// GetEnvCmd returns the command used to initialized the environment for an SDK
func (s *SDKs) GetEnvCmd(id string, defaultID string) []string {
	if id == "" && defaultID == "" {
		// no env cmd
		return []string{}
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if iid, err := s.ResolveID(id); err == nil {
		if sdk, exist := s.Sdks[iid]; exist {
			return sdk.GetEnvCmd()
		}
	}

	if sdk, exist := s.Sdks[defaultID]; defaultID != "" && exist {
		return sdk.GetEnvCmd()
	}

	// Return default env that may be empty
	return []string{}
}

// Install Used to install a new SDK
func (s *SDKs) Install(id, filepath string, force bool, timeout int, sess *ClientSession) (*xsapiv1.SDK, error) {
	var cSdk *CrossSDK
	if id != "" && filepath != "" {
		return nil, fmt.Errorf("invalid parameter, both id and filepath are set")
	}
	if id != "" {
		var exist bool
		cSdk, exist = s.Sdks[id]
		if !exist {
			return nil, fmt.Errorf("unknown id")
		}
	} else if filepath != "" {
		// TODO check that file is accessible

	} else {
		return nil, fmt.Errorf("invalid parameter, id or filepath must be set")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Launch script to install
	// (note that add event will be generated by monitoring thread)
	if err := cSdk.Install(filepath, force, timeout, sess); err != nil {
		return &cSdk.sdk, err
	}

	return &cSdk.sdk, nil
}

// AbortInstall Used to abort SDK installation
func (s *SDKs) AbortInstall(id string, timeout int) (*xsapiv1.SDK, error) {

	if id == "" {
		return nil, fmt.Errorf("invalid parameter")
	}
	cSdk, exist := s.Sdks[id]
	if !exist {
		return nil, fmt.Errorf("unknown id")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := cSdk.AbortInstallRemove(timeout)

	return &cSdk.sdk, err
}

// Remove Used to uninstall a SDK
func (s *SDKs) Remove(id string) (*xsapiv1.SDK, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cSdk, exist := s.Sdks[id]
	if !exist {
		return nil, fmt.Errorf("unknown id")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Launch script to remove/uninstall
	// (note that remove event will be generated by monitoring thread)
	if err := cSdk.Remove(); err != nil {
		return &cSdk.sdk, err
	}

	sdk := cSdk.sdk

	// Don't delete it from s.Sdks
	// (always keep sdk reference to allow for example re-install)

	return &sdk, nil
}
