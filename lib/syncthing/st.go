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

	"github.com/Sirupsen/logrus"
	"github.com/iotbzh/xds-server/lib/common"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
	"github.com/syncthing/syncthing/lib/config"
)

// SyncThing .
type SyncThing struct {
	BaseURL string
	APIKey  string
	Home    string
	STCmd   *exec.Cmd

	// Private fields
	binDir     string
	logsDir    string
	exitSTChan chan ExitChan
	client     *common.HTTPClient
	log        *logrus.Logger
}

// ExitChan Channel used for process exit
type ExitChan struct {
	status int
	err    error
}

// NewSyncThing creates a new instance of Syncthing
func NewSyncThing(conf *xdsconfig.Config, log *logrus.Logger) *SyncThing {
	var url, apiKey, home, binDir string
	var err error

	stCfg := conf.FileConf.SThgConf
	if stCfg != nil {
		url = stCfg.GuiAddress
		apiKey = stCfg.GuiAPIKey
		home = stCfg.Home
		binDir = stCfg.BinDir
	}

	if url == "" {
		url = "http://localhost:8384"
	}
	if url[0:7] != "http://" {
		url = "http://" + url
	}

	if home == "" {
		home = "/mnt/share"
	}

	if binDir == "" {
		if binDir, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
			binDir = "/usr/local/bin"
		}
	}

	s := SyncThing{
		BaseURL: url,
		APIKey:  apiKey,
		Home:    home,
		binDir:  binDir,
		logsDir: conf.FileConf.LogsDir,
		log:     log,
	}

	return &s
}

// Start Starts syncthing process
func (s *SyncThing) startProc(exeName string, args []string, env []string, eChan *chan ExitChan) (*exec.Cmd, error) {

	// Kill existing process (useful for debug ;-) )
	if os.Getenv("DEBUG_MODE") != "" {
		exec.Command("bash", "-c", "pkill -9 "+exeName).Output()
	}

	path, err := exec.LookPath(path.Join(s.binDir, exeName))
	if err != nil {
		return nil, fmt.Errorf("Cannot find %s executable in %s", exeName, s.binDir)
	}
	cmd := exec.Command(path, args...)
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
	}

	s.STCmd, err = s.startProc("syncthing", args, env, &s.exitSTChan)

	return s.STCmd, err
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

// Connect Establish HTTP connection with Syncthing
func (s *SyncThing) Connect() error {
	var err error
	s.client, err = common.HTTPNewClient(s.BaseURL,
		common.HTTPClientConfig{
			URLPrefix:           "/rest",
			HeaderClientKeyName: "X-Syncthing-ID",
		})
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
	return nil
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
