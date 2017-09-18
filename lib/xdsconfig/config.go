package xdsconfig

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	common "github.com/iotbzh/xds-common/golib"
)

// Config parameters (json format) of /config command
type Config struct {
	Version       string        `json:"version"`
	APIVersion    string        `json:"apiVersion"`
	VersionGitTag string        `json:"gitTag"`
	Builder       BuilderConfig `json:"builder"`

	// Private (un-exported fields in REST GET /config route)
	Options       Options        `json:"-"`
	FileConf      FileConfig     `json:"-"`
	Log           *logrus.Logger `json:"-"`
	LogVerboseOut io.Writer      `json:"-"`
}

// Options set at the command line
type Options struct {
	ConfigFile     string
	LogLevel       string
	LogFile        string
	NoFolderConfig bool
}

// Config default values
const (
	DefaultAPIVersion = "1"
	DefaultPort       = "8000"
	DefaultShareDir   = "${HOME}/.xds-server/projects"
	DefaultSdkRootDir = "/xdt/sdk"
)

// Init loads the configuration on start-up
func Init(cliCtx *cli.Context, log *logrus.Logger) (*Config, error) {
	var err error

	// Define default configuration
	c := Config{
		Version:       cliCtx.App.Metadata["version"].(string),
		APIVersion:    DefaultAPIVersion,
		VersionGitTag: cliCtx.App.Metadata["git-tag"].(string),
		Builder:       BuilderConfig{},

		Options: Options{
			ConfigFile:     cliCtx.GlobalString("config"),
			LogLevel:       cliCtx.GlobalString("log"),
			LogFile:        cliCtx.GlobalString("logfile"),
			NoFolderConfig: cliCtx.GlobalBool("no-folderconfig"),
		},
		FileConf: FileConfig{
			WebAppDir:    "webapp/dist",
			ShareRootDir: DefaultShareDir,
			SdkRootDir:   DefaultSdkRootDir,
			HTTPPort:     DefaultPort,
			LogsDir:      "",
		},
		Log: log,
	}

	// config file settings overwrite default config
	err = readGlobalConfig(&c, c.Options.ConfigFile)
	if err != nil {
		return nil, err
	}

	// Update location of shared dir if needed
	if !common.Exists(c.FileConf.ShareRootDir) {
		if err := os.MkdirAll(c.FileConf.ShareRootDir, 0770); err != nil {
			return nil, fmt.Errorf("No valid shared directory found: %v", err)
		}
	}
	c.Log.Infoln("Share root directory: ", c.FileConf.ShareRootDir)

	// Where Logs are redirected:
	//  default 'stdout' (logfile option default value)
	//  else use file (or filepath) set by --logfile option
	//  that may be overwritten by LogsDir field of config file
	logF := c.Options.LogFile
	logD := c.FileConf.LogsDir
	if logF != "stdout" {
		if logD != "" {
			lf := filepath.Base(logF)
			if lf == "" || lf == "." {
				lf = "xds-server.log"
			}
			logF = filepath.Join(logD, lf)
		} else {
			logD = filepath.Dir(logF)
		}
	}
	if logD == "" || logD == "." {
		logD = "/tmp/xds/logs"
	}
	c.Options.LogFile = logF
	c.FileConf.LogsDir = logD

	if c.FileConf.LogsDir != "" && !common.Exists(c.FileConf.LogsDir) {
		if err := os.MkdirAll(c.FileConf.LogsDir, 0770); err != nil {
			return nil, fmt.Errorf("Cannot create logs dir: %v", err)
		}
	}
	c.Log.Infoln("Logs file:      ", c.Options.LogFile)
	c.Log.Infoln("Logs directory: ", c.FileConf.LogsDir)

	return &c, nil
}
