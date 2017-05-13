package xdsconfig

import (
	"fmt"
	"strings"

	"os"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/iotbzh/xds-server/lib/syncthing"
)

// Config parameters (json format) of /config command
type Config struct {
	Version       string        `json:"version"`
	APIVersion    string        `json:"apiVersion"`
	VersionGitTag string        `json:"gitTag"`
	Builder       BuilderConfig `json:"builder"`
	Folders       FoldersConfig `json:"folders"`

	// Private / un-exported fields
	progName     string
	fileConf     FileConfig
	WebAppDir    string         `json:"-"`
	HTTPPort     string         `json:"-"`
	ShareRootDir string         `json:"-"`
	Log          *logrus.Logger `json:"-"`
	SThg         *st.SyncThing  `json:"-"`
}

// Config default values
const (
	DefaultAPIVersion = "1"
	DefaultPort       = "8000"
	DefaultShareDir   = "/mnt/share"
	DefaultLogLevel   = "error"
)

// Init loads the configuration on start-up
func Init(ctx *cli.Context) (Config, error) {
	var err error

	// Set logger level and formatter
	log := ctx.App.Metadata["logger"].(*logrus.Logger)

	logLevel := ctx.GlobalString("log")
	if logLevel == "" {
		logLevel = DefaultLogLevel
	}
	if log.Level, err = logrus.ParseLevel(logLevel); err != nil {
		fmt.Printf("Invalid log level : \"%v\"\n", logLevel)
		os.Exit(1)
	}
	log.Formatter = &logrus.TextFormatter{}

	// Define default configuration
	c := Config{
		Version:       ctx.App.Metadata["version"].(string),
		APIVersion:    DefaultAPIVersion,
		VersionGitTag: ctx.App.Metadata["git-tag"].(string),
		Builder:       BuilderConfig{},
		Folders:       FoldersConfig{},

		progName:     ctx.App.Name,
		WebAppDir:    "webapp/dist",
		HTTPPort:     DefaultPort,
		ShareRootDir: DefaultShareDir,
		Log:          log,
		SThg:         nil,
	}

	// config file settings overwrite default config
	err = updateConfigFromFile(&c, ctx.GlobalString("config"))
	if err != nil {
		return Config{}, err
	}

	// Update location of shared dir if needed
	if !dirExists(c.ShareRootDir) {
		if err := os.MkdirAll(c.ShareRootDir, 0770); err != nil {
			c.Log.Fatalf("No valid shared directory found (err=%v)", err)
		}
	}
	c.Log.Infoln("Share root directory: ", c.ShareRootDir)

	// FIXME - add a builder interface and support other builder type (eg. native)
	builderType := "syncthing"

	switch builderType {
	case "syncthing":
		// Syncthing settings only configurable from config.json file
		stGuiAddr := c.fileConf.SThgConf.GuiAddress
		stGuiApikey := c.fileConf.SThgConf.GuiAPIKey
		if stGuiAddr == "" {
			stGuiAddr = "http://localhost:8384"
		}
		if stGuiAddr[0:7] != "http://" {
			stGuiAddr = "http://" + stGuiAddr
		}

		// Retry if connection fail
		retry := 5
		for retry > 0 {
			c.SThg = st.NewSyncThing(stGuiAddr, stGuiApikey, c.Log)
			if c.SThg != nil {
				break
			}
			c.Log.Warningf("Establishing connection to Syncthing (retry %d/5)", retry)
			time.Sleep(time.Second)
			retry--
		}
		if c.SThg == nil {
			c.Log.Fatalf("ERROR: cannot connect to Syncthing (url: %s)", stGuiAddr)
		}

		// Retrieve Syncthing config
		id, err := c.SThg.IDGet()
		if err != nil {
			return Config{}, err
		}

		if c.Builder, err = NewBuilderConfig(id); err != nil {
			c.Log.Fatalln(err)
		}

		// Retrieve initial Syncthing config
		stCfg, err := c.SThg.ConfigGet()
		if err != nil {
			return Config{}, err
		}
		for _, stFld := range stCfg.Folders {
			relativePath := strings.TrimPrefix(stFld.RawPath, c.ShareRootDir)
			if relativePath == "" {
				relativePath = stFld.RawPath
			}
			newFld := NewFolderConfig(stFld.ID, stFld.Label, c.ShareRootDir, strings.Trim(relativePath, "/"))
			c.Folders = c.Folders.Update(FoldersConfig{newFld})
		}

	default:
		log.Fatalln("Unsupported builder type")
	}

	return c, nil
}

// GetFolderFromID retrieves the Folder config from id
func (c *Config) GetFolderFromID(id string) *FolderConfig {
	if idx := c.Folders.GetIdx(id); idx != -1 {
		return &c.Folders[idx]
	}
	return nil
}

// UpdateAll updates all the current configuration
func (c *Config) UpdateAll(newCfg Config) error {
	return fmt.Errorf("Not Supported")
	/*
		if err := VerifyConfig(newCfg); err != nil {
			return err
		}

		// TODO: c.Builder = c.Builder.Update(newCfg.Builder)
		c.Folders = c.Folders.Update(newCfg.Folders)

		// SEB A SUP model.NotifyListeners(c, NotifyFoldersChange, FolderConfig{})
		// FIXME To be tested & improved error handling
		for _, f := range c.Folders {
			if err := c.SThg.FolderChange(st.FolderChangeArg{
				ID:           f.ID,
				Label:        f.Label,
				RelativePath: f.RelativePath,
				SyncThingID:  f.SyncThingID,
				ShareRootDir: c.ShareRootDir,
			}); err != nil {
				return err
			}
		}

		return nil
	*/
}

// UpdateFolder updates a specific folder into the current configuration
func (c *Config) UpdateFolder(newFolder FolderConfig) (FolderConfig, error) {
	// rootPath should not be empty
	if newFolder.rootPath == "" {
		newFolder.rootPath = c.ShareRootDir
	}

	// Sanity check of folder settings
	if err := FolderVerify(newFolder); err != nil {
		return FolderConfig{}, err
	}

	c.Folders = c.Folders.Update(FoldersConfig{newFolder})

	// SEB A SUP model.NotifyListeners(c, NotifyFolderAdd, newFolder)
	err := c.SThg.FolderChange(st.FolderChangeArg{
		ID:           newFolder.ID,
		Label:        newFolder.Label,
		RelativePath: newFolder.RelativePath,
		SyncThingID:  newFolder.SyncThingID,
		ShareRootDir: c.ShareRootDir,
	})

	newFolder.BuilderSThgID = c.Builder.SyncThingID // FIXME - should be removed after local ST config rework
	newFolder.Status = FolderStatusEnable

	return newFolder, err
}

// DeleteFolder deletes a specific folder
func (c *Config) DeleteFolder(id string) (FolderConfig, error) {
	var fld FolderConfig
	var err error

	//SEB A SUP model.NotifyListeners(c, NotifyFolderDelete, fld)
	if err = c.SThg.FolderDelete(id); err != nil {
		return fld, err
	}

	c.Folders, fld, err = c.Folders.Delete(id)

	return fld, err
}

func dirExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
