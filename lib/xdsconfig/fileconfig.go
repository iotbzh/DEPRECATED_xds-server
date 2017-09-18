package xdsconfig

import (
	"encoding/json"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	common "github.com/iotbzh/xds-common/golib"
)

const (
	// ConfigDir Directory in user HOME directory where xds config will be saved
	ConfigDir = ".xds-server"
	// GlobalConfigFilename Global config filename
	GlobalConfigFilename = "config.json"
	// FoldersConfigFilename Folders config filename
	FoldersConfigFilename = "server-config_folders.xml"
)

// SyncThingConf definition
type SyncThingConf struct {
	BinDir          string `json:"binDir"`
	Home            string `json:"home"`
	GuiAddress      string `json:"gui-address"`
	GuiAPIKey       string `json:"gui-apikey"`
	RescanIntervalS int    `json:"rescanIntervalS"`
}

// FileConfig is the JSON structure of xds-server config file (config.json)
type FileConfig struct {
	WebAppDir    string         `json:"webAppDir"`
	ShareRootDir string         `json:"shareRootDir"`
	SdkRootDir   string         `json:"sdkRootDir"`
	HTTPPort     string         `json:"httpPort"`
	SThgConf     *SyncThingConf `json:"syncthing"`
	LogsDir      string         `json:"logsDir"`
}

// readGlobalConfig reads configuration from a config file.
// Order to determine which config file is used:
//  1/ from command line option: "--config myConfig.json"
//  2/ $HOME/.xds-server/config.json file
//  3/ /etc/xds-server/config.json file
//  4/ <xds-server executable dir>/config.json file
func readGlobalConfig(c *Config, confFile string) error {

	searchIn := make([]string, 0, 3)
	if confFile != "" {
		searchIn = append(searchIn, confFile)
	}
	if usr, err := user.Current(); err == nil {
		searchIn = append(searchIn, path.Join(usr.HomeDir, ConfigDir,
			GlobalConfigFilename))
	}

	searchIn = append(searchIn, "/etc/xds-server/config.json")

	exePath := os.Args[0]
	ee, _ := os.Executable()
	exeAbsPath, err := filepath.Abs(ee)
	if err == nil {
		exePath, err = filepath.EvalSymlinks(exeAbsPath)
		if err == nil {
			exePath = filepath.Dir(ee)
		} else {
			exePath = filepath.Dir(exeAbsPath)
		}
	}
	searchIn = append(searchIn, path.Join(exePath, "config.json"))

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

	// Support environment variables (IOW ${MY_ENV_VAR} syntax) in config.json
	vars := []*string{
		&fCfg.WebAppDir,
		&fCfg.ShareRootDir,
		&fCfg.SdkRootDir,
		&fCfg.LogsDir}
	if fCfg.SThgConf != nil {
		vars = append(vars, &fCfg.SThgConf.Home, &fCfg.SThgConf.BinDir)
	}
	for _, field := range vars {
		var err error
		if *field, err = common.ResolveEnvVar(*field); err != nil {
			return err
		}
	}

	// Use config file settings else use default config
	if fCfg.WebAppDir == "" {
		fCfg.WebAppDir = c.FileConf.WebAppDir
	}
	if fCfg.ShareRootDir == "" {
		fCfg.ShareRootDir = c.FileConf.ShareRootDir
	}
	if fCfg.SdkRootDir == "" {
		fCfg.SdkRootDir = c.FileConf.SdkRootDir
	}
	if fCfg.HTTPPort == "" {
		fCfg.HTTPPort = c.FileConf.HTTPPort
	}
	if fCfg.LogsDir == "" {
		fCfg.LogsDir = c.FileConf.LogsDir
	}

	// Resolve webapp dir (support relative or full path)
	fCfg.WebAppDir = strings.Trim(fCfg.WebAppDir, " ")
	if !strings.HasPrefix(fCfg.WebAppDir, "/") && exePath != "" {
		cwd, _ := os.Getwd()

		// Check first from current directory
		for _, rootD := range []string{exePath, cwd} {
			ff := path.Join(rootD, fCfg.WebAppDir, "index.html")
			if common.Exists(ff) {
				fCfg.WebAppDir = path.Join(rootD, fCfg.WebAppDir)
				break
			}
		}
	}

	c.FileConf = fCfg
	return nil
}

// FoldersConfigFilenameGet
func FoldersConfigFilenameGet() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return path.Join(usr.HomeDir, ConfigDir, FoldersConfigFilename), nil
}
