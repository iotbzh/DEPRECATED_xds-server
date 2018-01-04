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
)

// SDKs List of installed SDK
type SDKs struct {
	*Context
	Sdks         map[string]*CrossSDK
	SdksFamilies map[string]*xsapiv1.SDKFamilyConfig

	mutex sync.Mutex
	stop  chan struct{} // signals intentional stop
}

// NewSDKs creates a new instance of SDKs
func NewSDKs(ctx *Context) (*SDKs, error) {
	s := SDKs{
		Context:      ctx,
		Sdks:         make(map[string]*CrossSDK),
		SdksFamilies: make(map[string]*xsapiv1.SDKFamilyConfig),
		stop:         make(chan struct{}),
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
	for _, d := range dirs {
		if !common.IsDir(d) {
			continue
		}

		sdksList, err := ListCrossSDK(d, s.Log)
		if err != nil {
			// allow to use XDS even if error on list
			s.Log.Errorf("Cannot retrieve SDK list: %v", err)
		}
		s.LogSillyf("'%s' SDKs list: %v", d, sdksList)

		for _, sdk := range sdksList {
			cSdk, err := s._createNewCrossSDK(sdk, d, false, false)
			if err != nil {
				s.Log.Debugf("Error while processing SDK sdk=%v\n err=%s", sdk, err.Error())
				continue
			}

			if cSdk.sdk.Status == xsapiv1.SdkStatusInstalled {
				nbInstalled++
			}

			s.SdksFamilies[cSdk.sdk.FamilyConf.FamilyName] = &cSdk.sdk.FamilyConf
		}
	}

	ctx.Log.Debugf("Cross SDKs: %d defined, %d installed", len(s.Sdks), nbInstalled)

	// Start monitor thread to detect new SDKs
	sdksDirs := []string{}
	for _, sf := range s.SdksFamilies {
		sdksDirs = append(sdksDirs, sf.RootDir)
	}

	if len(s.SdksFamilies) == 0 {
		s.Log.Warningf("No cross SDKs definition found")
		/* TODO: used it or cleanup
		} else {
			go s.monitorSDKInstallation(sdksDirs)
		*/
	}

	return &s, nil
}

// _createNewCrossSDK Private function to create a new Cross SDK
func (s *SDKs) _createNewCrossSDK(sdk xsapiv1.SDK, scriptDir string, installing bool, force bool) (*CrossSDK, error) {

	cSdk, err := NewCrossSDK(s.Context, sdk, scriptDir)
	if err != nil {
		return cSdk, err
	}

	// Allow to overwrite not installed SDK or when force is set
	if _, exist := s.Sdks[cSdk.sdk.ID]; exist {
		if !force && cSdk.sdk.Path != "" && common.Exists(cSdk.sdk.Path) {
			return cSdk, fmt.Errorf("SDK ID %s already installed in %s", cSdk.sdk.ID, cSdk.sdk.Path)
		}
		if !force && cSdk.sdk.Status != xsapiv1.SdkStatusNotInstalled {
			return cSdk, fmt.Errorf("Duplicate SDK ID %s (use force to overwrite)", cSdk.sdk.ID)
		}
	}

	// Sanity check
	errMsg := "Invalid SDK definition "
	if installing && cSdk.sdk.Path == "" {
		return cSdk, fmt.Errorf(errMsg + "(path not set)")
	}
	if installing && cSdk.sdk.URL == "" {
		return cSdk, fmt.Errorf(errMsg + "(url not set)")
	}

	// Add to list
	s.Sdks[cSdk.sdk.ID] = cSdk

	return cSdk, err
}

// Stop SDKs management
func (s *SDKs) Stop() {
	close(s.stop)
}

// monitorSDKInstallation
/* TODO: used it or cleanup
import 	"github.com/zillode/notify"

func (s *SDKs) monitorSDKInstallation(watchingDirs []string) {

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
	for _, dir := range watchingDirs {
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
				sdkDef, err := GetSDKInfo(scriptDir, sdk.URL, "", "", s.Log)
				if err != nil {
					s.Log.Warningf("Cannot get sdk info: %v", err)
					continue
				}
				sdk.Path = sdkDef.Path
				sdk.Path = sdkDef.SetupFile

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
	return id, fmt.Errorf("Multiple sdk IDs found: %v", match)
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
func (s *SDKs) Install(id, filepath string, force bool, timeout int, args []string, sess *ClientSession) (*xsapiv1.SDK, error) {

	var sdk *xsapiv1.SDK
	var err error
	scriptDir := ""
	sdkFilename := ""

	if id != "" && filepath != "" {
		return nil, fmt.Errorf("invalid parameter, both id and filepath are set")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if id != "" {
		curSdk, exist := s.Sdks[id]
		if !exist {
			return nil, fmt.Errorf("unknown id")
		}

		sdk = &curSdk.sdk
		scriptDir = sdk.FamilyConf.ScriptsDir

		// Update path when not set
		if sdk.Path == "" {
			sdkDef, err := GetSDKInfo(scriptDir, sdk.URL, "", "", s.Log)
			if err != nil || sdkDef.Path == "" {
				return nil, fmt.Errorf("cannot retrieve sdk path %v", err)
			}
			sdk.Path = sdkDef.Path
		}

	} else if filepath != "" {
		// FIXME support any location and also sharing either by pathmap or Syncthing
		baseDir := "${HOME}/xds-workspace/sdks"
		sdkFilename, _ = common.ResolveEnvVar(path.Join(baseDir, path.Base(filepath)))
		if !common.Exists(sdkFilename) {
			return nil, fmt.Errorf("SDK file not accessible, must be in %s", baseDir)
		}

		for _, sf := range s.SdksFamilies {
			sdkDef, err := GetSDKInfo(sf.ScriptsDir, "", sdkFilename, "", s.Log)
			if err == nil {
				// OK, sdk found
				sdk = &sdkDef
				scriptDir = sf.ScriptsDir
				break
			}

			s.Log.Debugf("GetSDKInfo error: family=%s, sdkFilename=%s, err=%v", sf.FamilyName, path.Base(sdkFilename), err)
		}
		if sdk == nil {
			return nil, fmt.Errorf("Cannot identify SDK family for %s", path.Base(filepath))
		}

	} else {
		return nil, fmt.Errorf("invalid parameter, id or filepath must be set")
	}

	cSdk, err := s._createNewCrossSDK(*sdk, scriptDir, true, force)
	if err != nil {
		return nil, err
	}

	// Launch script to install
	// (note that add event will be generated by monitoring thread)
	if err := cSdk.Install(sdkFilename, force, timeout, args, sess); err != nil {
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
func (s *SDKs) Remove(id string, timeout int, sess *ClientSession) (*xsapiv1.SDK, error) {

	cSdk, exist := s.Sdks[id]
	if !exist {
		return nil, fmt.Errorf("unknown id")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Launch script to remove/uninstall
	// (note that remove event will be generated by monitoring thread)
	if err := cSdk.Remove(timeout, sess); err != nil {
		return &cSdk.sdk, err
	}

	sdk := cSdk.sdk

	// Don't delete it from s.Sdks
	// (always keep sdk reference to allow for example re-install)

	return &sdk, nil
}
