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

package st

import (
	"encoding/json"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"strings"

	"fmt"

	"io"

	"io/ioutil"

	"regexp"

	"github.com/Sirupsen/logrus"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
	"github.com/syncthing/syncthing/lib/config"
)

// SyncThing .
type SyncThing struct {
	BaseURL   string
	APIKey    string
	Home      string
	STCmd     *exec.Cmd
	STICmd    *exec.Cmd
	MyID      string
	Connected bool
	Events    *Events

	// Private fields
	binDir      string
	logsDir     string
	exitSTChan  chan ExitChan
	exitSTIChan chan ExitChan
	client      *common.HTTPClient
	log         *logrus.Logger
	conf        *xdsconfig.Config
}

// ExitChan Channel used for process exit
type ExitChan struct {
	status int
	err    error
}

// ConfigInSync Check whether if Syncthing configuration is in sync
type configInSync struct {
	ConfigInSync bool `json:"configInSync"`
}

// FolderStatus Information about the current status of a folder.
type FolderStatus struct {
	GlobalFiles       int   `json:"globalFiles"`
	GlobalDirectories int   `json:"globalDirectories"`
	GlobalSymlinks    int   `json:"globalSymlinks"`
	GlobalDeleted     int   `json:"globalDeleted"`
	GlobalBytes       int64 `json:"globalBytes"`

	LocalFiles       int   `json:"localFiles"`
	LocalDirectories int   `json:"localDirectories"`
	LocalSymlinks    int   `json:"localSymlinks"`
	LocalDeleted     int   `json:"localDeleted"`
	LocalBytes       int64 `json:"localBytes"`

	NeedFiles       int   `json:"needFiles"`
	NeedDirectories int   `json:"needDirectories"`
	NeedSymlinks    int   `json:"needSymlinks"`
	NeedDeletes     int   `json:"needDeletes"`
	NeedBytes       int64 `json:"needBytes"`

	InSyncFiles int   `json:"inSyncFiles"`
	InSyncBytes int64 `json:"inSyncBytes"`

	State        string    `json:"state"`
	StateChanged time.Time `json:"stateChanged"`

	Sequence int64 `json:"sequence"`

	IgnorePatterns bool `json:"ignorePatterns"`
}

// NewSyncThing creates a new instance of Syncthing
func NewSyncThing(conf *xdsconfig.Config, log *logrus.Logger) *SyncThing {
	var url, apiKey, home, binDir string

	stCfg := conf.FileConf.SThgConf
	if stCfg != nil {
		url = stCfg.GuiAddress
		apiKey = stCfg.GuiAPIKey
		home = stCfg.Home
		binDir = stCfg.BinDir
	}

	if url == "" {
		url = "http://localhost:8385"
	}
	if url[0:7] != "http://" {
		url = "http://" + url
	}

	if home == "" {
		panic("home parameter must be set")
	}

	s := SyncThing{
		BaseURL: url,
		APIKey:  apiKey,
		Home:    home,
		binDir:  binDir,
		logsDir: conf.FileConf.LogsDir,
		log:     log,
		conf:    conf,
	}

	// Create Events monitoring
	s.Events = s.NewEventListener()

	return &s
}

// Start Starts syncthing process
func (s *SyncThing) startProc(exeName string, args []string, env []string, eChan *chan ExitChan) (*exec.Cmd, error) {
	var err error
	var exePath string

	// Kill existing process (useful for debug ;-) )
	if os.Getenv("DEBUG_MODE") != "" {
		fmt.Printf("\n!!! DEBUG_MODE set: KILL existing %s process(es) !!!\n", exeName)
		exec.Command("bash", "-c", "ps -ax |grep "+exeName+" |grep "+s.BaseURL+" |cut  -d' ' -f 1|xargs -I{} kill -9 {}").Output()
	}

	// When not set (or set to '.') set bin to path of xds-agent executable
	bdir := s.binDir
	if bdir == "" || bdir == "." {
		exe, _ := os.Executable()
		if exeAbsPath, err := filepath.Abs(exe); err == nil {
			if exePath, err := filepath.EvalSymlinks(exeAbsPath); err == nil {
				bdir = filepath.Dir(exePath)
			}
		}
	}

	exePath, err = exec.LookPath(path.Join(bdir, exeName))
	if err != nil {
		// Let's try in /opt/AGL/bin
		exePath, err = exec.LookPath(path.Join("opt", "AGL", "bin", exeName))
		if err != nil {
			return nil, fmt.Errorf("Cannot find %s executable in %s", exeName, bdir)
		}
	}
	cmd := exec.Command(exePath, args...)
	cmd.Env = os.Environ()
	for _, ev := range env {
		cmd.Env = append(cmd.Env, ev)
	}

	// open log file
	var outfile *os.File
	logFilename := filepath.Join(s.logsDir, exeName+".log")
	if s.logsDir != "" {
		outfile, err := os.Create(logFilename)
		if err != nil {
			return nil, fmt.Errorf("Cannot create log file %s", logFilename)
		}

		cmdOut, err := cmd.StdoutPipe()
		if err != nil {
			return nil, fmt.Errorf("Pipe stdout error for : %s", err)
		}

		go io.Copy(outfile, cmdOut)
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	*eChan = make(chan ExitChan, 1)
	go func(c *exec.Cmd, oF *os.File) {
		status := 0
		sts, err := c.Process.Wait()
		if !sts.Success() {
			s := sts.Sys().(syscall.WaitStatus)
			status = s.ExitStatus()
		}
		if oF != nil {
			oF.Close()
		}
		s.log.Debugf("%s exited with status %d, err %v", exeName, status, err)

		*eChan <- ExitChan{status, err}
	}(cmd, outfile)

	return cmd, nil
}

// Start Starts syncthing process
func (s *SyncThing) Start() (*exec.Cmd, error) {
	var err error

	s.log.Infof(" ST home=%s", s.Home)
	s.log.Infof(" ST  url=%s", s.BaseURL)

	args := []string{
		"--home=" + s.Home,
		"-no-browser",
		"--gui-address=" + s.BaseURL,
	}

	if s.APIKey != "" {
		args = append(args, "-gui-apikey=\""+s.APIKey+"\"")
		s.log.Infof(" ST apikey=%s", s.APIKey)
	}
	if s.log.Level == logrus.DebugLevel {
		args = append(args, "-verbose")
	}

	env := []string{
		"STNODEFAULTFOLDER=1",
		"STNOUPGRADE=1",
	}

	s.STCmd, err = s.startProc("syncthing", args, env, &s.exitSTChan)

	// Use autogenerated apikey if not set by config.json
	if err == nil && s.APIKey == "" {
		if fd, err := os.Open(filepath.Join(s.Home, "config.xml")); err == nil {
			defer fd.Close()
			if b, err := ioutil.ReadAll(fd); err == nil {
				re := regexp.MustCompile("<apikey>(.*)</apikey>")
				key := re.FindStringSubmatch(string(b))
				if len(key) >= 1 {
					s.APIKey = key[1]
				}
			}
		}
	}

	return s.STCmd, err
}

// StartInotify Starts syncthing-inotify process
func (s *SyncThing) StartInotify() (*exec.Cmd, error) {
	var err error
	exeName := "syncthing-inotify"

	s.log.Infof(" STI  url=%s", s.BaseURL)

	args := []string{
		"-target=" + s.BaseURL,
	}
	if s.APIKey != "" {
		args = append(args, "-api="+s.APIKey)
		s.log.Infof("%s uses apikey=%s", exeName, s.APIKey)
	}
	if s.log.Level == logrus.DebugLevel {
		args = append(args, "-verbosity=4")
	}

	env := []string{}

	s.STICmd, err = s.startProc(exeName, args, env, &s.exitSTIChan)

	return s.STICmd, err
}

func (s *SyncThing) stopProc(pname string, proc *os.Process, exit chan ExitChan) {
	if err := proc.Signal(os.Interrupt); err != nil {
		s.log.Infof("Proc interrupt %s error: %s", pname, err.Error())

		select {
		case <-exit:
		case <-time.After(time.Second):
			// A bigger bonk on the head.
			if err := proc.Signal(os.Kill); err != nil {
				s.log.Infof("Proc term %s error: %s", pname, err.Error())
			}
			<-exit
		}
	}
	s.log.Infof("%s stopped (PID %d)", pname, proc.Pid)
}

// Stop Stops syncthing process
func (s *SyncThing) Stop() {
	if s.STCmd == nil {
		return
	}
	s.stopProc("syncthing", s.STCmd.Process, s.exitSTChan)
	s.STCmd = nil
}

// StopInotify Stops syncthing process
func (s *SyncThing) StopInotify() {
	if s.STICmd == nil {
		return
	}
	s.stopProc("syncthing-inotify", s.STICmd.Process, s.exitSTIChan)
	s.STICmd = nil
}

// Connect Establish HTTP connection with Syncthing
func (s *SyncThing) Connect() error {
	var err error
	s.Connected = false
	s.client, err = common.HTTPNewClient(s.BaseURL,
		common.HTTPClientConfig{
			URLPrefix:           "/rest",
			HeaderClientKeyName: "X-Syncthing-ID",
			LogOut:              s.conf.LogVerboseOut,
			LogPrefix:           "SYNCTHING: ",
			LogLevel:            common.HTTPLogLevelWarning,
		})
	s.client.SetLogLevel(s.log.Level.String())

	if err != nil {
		msg := ": " + err.Error()
		if strings.Contains(err.Error(), "connection refused") {
			msg = fmt.Sprintf("(url: %s)", s.BaseURL)
		}
		return fmt.Errorf("ERROR: cannot connect to Syncthing %s", msg)
	}
	if s.client == nil {
		return fmt.Errorf("ERROR: cannot connect to Syncthing (null client)")
	}

	s.MyID, err = s.IDGet()
	if err != nil {
		return fmt.Errorf("ERROR: cannot retrieve ID")
	}

	s.Connected = true

	// Start events monitoring
	err = s.Events.Start()

	return err
}

// IDGet returns the Syncthing ID of Syncthing instance running locally
func (s *SyncThing) IDGet() (string, error) {
	var data []byte
	if err := s.client.HTTPGet("system/status", &data); err != nil {
		return "", err
	}
	status := make(map[string]interface{})
	json.Unmarshal(data, &status)
	return status["myID"].(string), nil
}

// ConfigGet returns the current Syncthing configuration
func (s *SyncThing) ConfigGet() (config.Configuration, error) {
	var data []byte
	config := config.Configuration{}
	if err := s.client.HTTPGet("system/config", &data); err != nil {
		return config, err
	}
	err := json.Unmarshal(data, &config)
	return config, err
}

// ConfigSet set Syncthing configuration
func (s *SyncThing) ConfigSet(cfg config.Configuration) error {
	body, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return s.client.HTTPPost("system/config", string(body))
}

// IsConfigInSync Returns true if configuration is in sync
func (s *SyncThing) IsConfigInSync() (bool, error) {
	var data []byte
	var d configInSync
	if err := s.client.HTTPGet("system/config/insync", &data); err != nil {
		return false, err
	}
	if err := json.Unmarshal(data, &d); err != nil {
		return false, err
	}
	return d.ConfigInSync, nil
}
