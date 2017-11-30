package xdsconfig

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
)

// Config parameters (json format) of /config command
type Config struct {
	// Public APIConfig fields
	xsapiv1.APIConfig

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
	DefaultSTHomeDir  = "${HOME}/.xds-server/syncthing-config"
	DefaultSdkRootDir = "/xdt/sdk"
)

// Init loads the configuration on start-up
func Init(cliCtx *cli.Context, log *logrus.Logger) (*Config, error) {
	var err error

	dfltShareDir := DefaultShareDir
	dfltSTHomeDir := DefaultSTHomeDir
	if resDir, err := common.ResolveEnvVar(DefaultShareDir); err == nil {
		dfltShareDir = resDir
	}
	if resDir, err := common.ResolveEnvVar(DefaultSTHomeDir); err == nil {
		dfltSTHomeDir = resDir
	}

	// Retrieve Server ID (or create one the first time)
	uuid, err := ServerIDGet()
	if err != nil {
		return nil, err
	}

	// Define default configuration
	c := Config{
		APIConfig: xsapiv1.APIConfig{
			ServerUID:        uuid,
			Version:          cliCtx.App.Metadata["version"].(string),
			APIVersion:       DefaultAPIVersion,
			VersionGitTag:    cliCtx.App.Metadata["git-tag"].(string),
			Builder:          xsapiv1.BuilderConfig{},
			SupportedSharing: map[string]bool{ /*FIXME USE folder.TypePathMap*/ "PathMap": true},
		},

		Options: Options{
			ConfigFile:     cliCtx.GlobalString("config"),
			LogLevel:       cliCtx.GlobalString("log"),
			LogFile:        cliCtx.GlobalString("logfile"),
			NoFolderConfig: cliCtx.GlobalBool("no-folderconfig"),
		},
		FileConf: FileConfig{
			WebAppDir:    "webapp/dist",
			ShareRootDir: dfltShareDir,
			SdkRootDir:   DefaultSdkRootDir,
			HTTPPort:     DefaultPort,
			SThgConf:     &SyncThingConf{Home: dfltSTHomeDir},
			LogsDir:      "",
		},
		Log: log,
	}

	c.Log.Infoln("Server UUID:          ", uuid)

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

	c.Log.Infoln("Logs file:            ", c.Options.LogFile)
	c.Log.Infoln("Logs directory:       ", c.FileConf.LogsDir)

	return &c, nil
}
