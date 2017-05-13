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
)

type SyncThingConf struct {
	Home       string `json:"home"`
	GuiAddress string `json:"gui-address"`
	GuiAPIKey  string `json:"gui-apikey"`
}

type FileConfig struct {
	WebAppDir    string        `json:"webAppDir"`
	ShareRootDir string        `json:"shareRootDir"`
	HTTPPort     string        `json:"httpPort"`
	SThgConf     SyncThingConf `json:"syncthing"`
}

// getConfigFromFile reads configuration from a config file.
// Order to determine which config file is used:
//  1/ from command line option: "--config myConfig.json"
//  2/ $HOME/.xds/config.json file
//  3/ <xds-server executable dir>/config.json file

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
	c.fileConf = fCfg

	// Support environment variables (IOW ${MY_ENV_VAR} syntax) in config.json
	// TODO: better to use reflect package to iterate on fields and be more generic
	var rep string
	if rep, err = resolveEnvVar(fCfg.WebAppDir); err != nil {
		return err
	}
	fCfg.WebAppDir = path.Clean(rep)

	if rep, err = resolveEnvVar(fCfg.ShareRootDir); err != nil {
		return err
	}
	fCfg.ShareRootDir = path.Clean(rep)

	if rep, err = resolveEnvVar(fCfg.SThgConf.Home); err != nil {
		return err
	}
	fCfg.SThgConf.Home = path.Clean(rep)

	// Config file settings overwrite default config

	if fCfg.WebAppDir != "" {
		c.WebAppDir = strings.Trim(fCfg.WebAppDir, " ")
	}
	// Is it a full path ?
	if !strings.HasPrefix(c.WebAppDir, "/") && exePath != "" {
		// Check first from current directory
		for _, rootD := range []string{cwd, exePath} {
			ff := path.Join(rootD, c.WebAppDir, "index.html")
			if exists(ff) {
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

// exists returns whether the given file or directory exists or not
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}
