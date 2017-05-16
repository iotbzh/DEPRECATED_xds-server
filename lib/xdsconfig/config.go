package xdsconfig

import (
	"fmt"

	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

// Config parameters (json format) of /config command
type Config struct {
	Version       string        `json:"version"`
	APIVersion    string        `json:"apiVersion"`
	VersionGitTag string        `json:"gitTag"`
	Builder       BuilderConfig `json:"builder"`
	Folders       FoldersConfig `json:"folders"`

	// Private (un-exported fields in REST GET /config route)
	FileConf     FileConfig     `json:"-"`
	WebAppDir    string         `json:"-"`
	HTTPPort     string         `json:"-"`
	ShareRootDir string         `json:"-"`
	Log          *logrus.Logger `json:"-"`
}

// Config default values
const (
	DefaultAPIVersion = "1"
	DefaultPort       = "8000"
	DefaultShareDir   = "/mnt/share"
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
		Folders:       FoldersConfig{},

		WebAppDir:    "webapp/dist",
		HTTPPort:     DefaultPort,
		ShareRootDir: DefaultShareDir,
		Log:          log,
	}

	// config file settings overwrite default config
	err = updateConfigFromFile(&c, cliCtx.GlobalString("config"))
	if err != nil {
		return nil, err
	}

	// Update location of shared dir if needed
	if !dirExists(c.ShareRootDir) {
		if err := os.MkdirAll(c.ShareRootDir, 0770); err != nil {
			return nil, fmt.Errorf("No valid shared directory found: %v", err)
		}
	}
	c.Log.Infoln("Share root directory: ", c.ShareRootDir)

	if c.FileConf.LogsDir != "" && !dirExists(c.FileConf.LogsDir) {
		if err := os.MkdirAll(c.FileConf.LogsDir, 0770); err != nil {
			return nil, fmt.Errorf("Cannot create logs dir: %v", err)
		}
	}
	c.Log.Infoln("Logs directory: ", c.FileConf.LogsDir)

	return &c, nil
}

func dirExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
