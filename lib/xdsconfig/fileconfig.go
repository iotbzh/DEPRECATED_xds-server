package xdsconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/iotbzh/xds-server/lib/common"
)

type SyncThingConf struct {
	BinDir     string `json:"binDir"`
	Home       string `json:"home"`
	GuiAddress string `json:"gui-address"`
	GuiAPIKey  string `json:"gui-apikey"`
}

type FileConfig struct {
	WebAppDir    string         `json:"webAppDir"`
	ShareRootDir string         `json:"shareRootDir"`
	SdkRootDir   string        `json:"sdkRootDir"`
	HTTPPort     string         `json:"httpPort"`
	SThgConf     *SyncThingConf `json:"syncthing"`
	LogsDir      string         `json:"logsDir"`
}

// getConfigFromFile reads configuration from a config file.
// Order to determine which config file is used:
//  1/ from command line option: "--config myConfig.json"
//  2/ $HOME/.xds/config.json file
//  3/ <current_dir>/config.json file
//  4/ <xds-server executable dir>/config.json file

func updateConfigFromFile(c *Config, confFile string) error {

	searchIn := make([]string, 0, 3)
	if confFile != "" {
		searchIn = append(searchIn, confFile)
	}
	if usr, err := user.Current(); err == nil {
		searchIn = append(searchIn, path.Join(usr.HomeDir, ".xds", "config.json"))
	}
	cwd, err := os.Getwd()
	if err == nil {
		searchIn = append(searchIn, path.Join(cwd, "config.json"))
	}
	exePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		searchIn = append(searchIn, path.Join(exePath, "config.json"))
	}

	var cFile *string
	for _, p := range searchIn {
		if _, err := os.Stat(p); err == nil {
			cFile = &p
			break
		}
	}
	if cFile == nil {
		// No config file found
		return nil
	}

	c.Log.Infof("Use config file: %s", *cFile)

	// TODO move on viper package to support comments in JSON and also
	// bind with flags (command line options)
	// see https://github.com/spf13/viper#working-with-flags

	fd, _ := os.Open(*cFile)
	defer fd.Close()
	fCfg := FileConfig{}
	if err := json.NewDecoder(fd).Decode(&fCfg); err != nil {
		return err
	}
	c.FileConf = fCfg

	// Support environment variables (IOW ${MY_ENV_VAR} syntax) in config.json
	for _, field := range []*string{
		&fCfg.WebAppDir,
		&fCfg.ShareRootDir,
		&fCfg.SdkRootDir,
		&fCfg.LogsDir,
		&fCfg.SThgConf.Home} {

		rep, err := resolveEnvVar(*field)
		if err != nil {
			return err
		}
		*field = path.Clean(rep)
	}
	
	// Config file settings overwrite default config

	if fCfg.WebAppDir != "" {
		c.WebAppDir = strings.Trim(fCfg.WebAppDir, " ")
	}
	// Is it a full path ?
	if !strings.HasPrefix(c.WebAppDir, "/") && exePath != "" {
		// Check first from current directory
		for _, rootD := range []string{cwd, exePath} {
			ff := path.Join(rootD, c.WebAppDir, "index.html")
			if common.Exists(ff) {
				c.WebAppDir = path.Join(rootD, c.WebAppDir)
				break
			}
		}
	}

	if fCfg.ShareRootDir != "" {
		c.ShareRootDir = fCfg.ShareRootDir
	}

	if fCfg.HTTPPort != "" {
		c.HTTPPort = fCfg.HTTPPort
	}

	return nil
}

// resolveEnvVar Resolved environment variable regarding the syntax ${MYVAR}
func resolveEnvVar(s string) (string, error) {
	re := regexp.MustCompile("\\${(.*)}")
	vars := re.FindAllStringSubmatch(s, -1)
	res := s
	for _, v := range vars {
		val := os.Getenv(v[1])
		if val == "" {
			return res, fmt.Errorf("ERROR: %s env variable not defined", v[1])
		}

		rer := regexp.MustCompile("\\${" + v[1] + "}")
		res = rer.ReplaceAllString(res, val)
	}

	return res, nil
}
