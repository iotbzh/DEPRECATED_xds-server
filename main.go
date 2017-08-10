// TODO add Doc
//
package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/iotbzh/xds-server/lib/crosssdk"
	"github.com/iotbzh/xds-server/lib/model"
	"github.com/iotbzh/xds-server/lib/syncthing"
	"github.com/iotbzh/xds-server/lib/webserver"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
)

const (
	appName        = "xds-server"
	appDescription = "X(cross) Development System Server is a web server that allows to remotely cross build applications."
	appCopyright   = "Apache-2.0"
	appUsage       = "X(cross) Development System Server"
)

var appAuthors = []cli.Author{
	cli.Author{Name: "Sebastien Douheret", Email: "sebastien@iot.bzh"},
}

// AppVersion is the version of this application
var AppVersion = "?.?.?"

// AppSubVersion is the git tag id added to version string
// Should be set by compilation -ldflags "-X main.AppSubVersion=xxx"
var AppSubVersion = "unknown-dev"

// Context holds the XDS server context
type Context struct {
	ProgName    string
	Cli         *cli.Context
	Config      *xdsconfig.Config
	Log         *logrus.Logger
	SThg        *st.SyncThing
	SThgCmd     *exec.Cmd
	SThgInotCmd *exec.Cmd
	MFolders    *model.Folders
	SDKs        *crosssdk.SDKs
	WWWServer   *webserver.Server
	Exit        chan os.Signal
}

// NewContext Create a new instance of XDS server
func NewContext(cliCtx *cli.Context) *Context {
	var err error

	// Set logger level and formatter
	log := cliCtx.App.Metadata["logger"].(*logrus.Logger)

	logLevel := cliCtx.GlobalString("log")
	if logLevel == "" {
		logLevel = "error" // FIXME get from Config DefaultLogLevel
	}
	if log.Level, err = logrus.ParseLevel(logLevel); err != nil {
		fmt.Printf("Invalid log level : \"%v\"\n", logLevel)
		os.Exit(1)
	}
	log.Formatter = &logrus.TextFormatter{}

	// Define default configuration
	ctx := Context{
		ProgName: cliCtx.App.Name,
		Cli:      cliCtx,
		Log:      log,
		Exit:     make(chan os.Signal, 1),
	}

	// register handler on SIGTERM / exit
	signal.Notify(ctx.Exit, os.Interrupt, syscall.SIGTERM)
	go handlerSigTerm(&ctx)

	return &ctx
}

// Handle exit and properly stop/close all stuff
func handlerSigTerm(ctx *Context) {
	<-ctx.Exit
	if ctx.SThg != nil {
		ctx.Log.Infof("Stoping Syncthing... (PID %d)", ctx.SThgCmd.Process.Pid)
		ctx.SThg.Stop()
		ctx.Log.Infof("Stoping Syncthing-inotify... (PID %d)", ctx.SThgInotCmd.Process.Pid)
		ctx.SThg.StopInotify()
	}
	if ctx.WWWServer != nil {
		ctx.Log.Infof("Stoping Web server...")
		ctx.WWWServer.Stop()
	}
	os.Exit(0)
}

// XDS Server application main routine
func xdsApp(cliCtx *cli.Context) error {
	var err error

	// Create XDS server context
	ctx := NewContext(cliCtx)

	// Load config
	cfg, err := xdsconfig.Init(ctx.Cli, ctx.Log)
	if err != nil {
		return cli.NewExitError(err, -2)
	}
	ctx.Config = cfg

	// Logs redirected into a file when logsDir is set
	logfilename := cliCtx.GlobalString("logfile")
	if ctx.Config.FileConf.LogsDir != "" && logfilename != "stdout" {
		if logfilename == "" {
			logfilename = "xds-server.log"
		}
		// is it an absolute path ?
		logFile := logfilename
		if logfilename[0] == '.' || logfilename[0] != '/' {
			logFile = filepath.Join(ctx.Config.FileConf.LogsDir, logfilename)
		}
		fmt.Printf("Logging file: %s\n", logFile)
		fdL, err := os.OpenFile(logFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
		if err != nil {
			msgErr := fmt.Sprintf("Cannot create log file %s", logFile)
			return cli.NewExitError(msgErr, int(syscall.EPERM))
		}
		ctx.Log.Out = fdL
	}

	// Create syncthing instance when section "syncthing" is present in config.json
	if ctx.Config.FileConf.SThgConf != nil {
		ctx.SThg = st.NewSyncThing(ctx.Config, ctx.Log)
	}

	// Start local instance of Syncthing and Syncthing-notify
	if ctx.SThg != nil {
		ctx.Log.Infof("Starting Syncthing...")
		ctx.SThgCmd, err = ctx.SThg.Start()
		if err != nil {
			return cli.NewExitError(err, -4)
		}
		fmt.Printf("Syncthing started (PID %d)\n", ctx.SThgCmd.Process.Pid)

		ctx.Log.Infof("Starting Syncthing-inotify...")
		ctx.SThgInotCmd, err = ctx.SThg.StartInotify()
		if err != nil {
			return cli.NewExitError(err, -4)
		}
		fmt.Printf("Syncthing-inotify started (PID %d)\n", ctx.SThgInotCmd.Process.Pid)

		// Establish connection with local Syncthing (retry if connection fail)
		fmt.Printf("Establishing connection with Syncthing...\n")
		time.Sleep(2 * time.Second)
		maxRetry := 30
		retry := maxRetry
		err = nil
		for retry > 0 {
			if err = ctx.SThg.Connect(); err == nil {
				break
			}
			ctx.Log.Warningf("Establishing connection to Syncthing (retry %d/%d)", retry, maxRetry)
			time.Sleep(time.Second)
			retry--
		}
		if err != nil || retry == 0 {
			return cli.NewExitError(err, -4)
		}

		// FIXME: do we still need Builder notion ? if no cleanup
		if ctx.Config.Builder, err = xdsconfig.NewBuilderConfig(ctx.SThg.MyID); err != nil {
			return cli.NewExitError(err, -4)
		}
	}

	// Init model folder
	ctx.MFolders = model.FoldersNew(ctx.Config, ctx.SThg)

	// Load initial folders config from disk
	if err := ctx.MFolders.LoadConfig(); err != nil {
		return cli.NewExitError(err, -5)
	}

	// Init cross SDKs
	ctx.SDKs, err = crosssdk.Init(ctx.Config, ctx.Log)
	if err != nil {
		return cli.NewExitError(err, -6)
	}

	// Create and start Web Server
	ctx.WWWServer = webserver.New(ctx.Config, ctx.MFolders, ctx.SDKs, ctx.Log)
	if err = ctx.WWWServer.Serve(); err != nil {
		ctx.Log.Println(err)
		return cli.NewExitError(err, -7)
	}

	return cli.NewExitError("Program exited ", -99)
}

// main
func main() {

	// Create a new instance of the logger
	log := logrus.New()

	// Create a new App instance
	app := cli.NewApp()
	app.Name = appName
	app.Description = appDescription
	app.Usage = appUsage
	app.Version = AppVersion + " (" + AppSubVersion + ")"
	app.Authors = appAuthors
	app.Copyright = appCopyright
	app.Metadata = make(map[string]interface{})
	app.Metadata["version"] = AppVersion
	app.Metadata["git-tag"] = AppSubVersion
	app.Metadata["logger"] = log

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config, c",
			Usage:  "JSON config file to use\n\t",
			EnvVar: "APP_CONFIG",
		},
		cli.StringFlag{
			Name:   "log, l",
			Value:  "error",
			Usage:  "logging level (supported levels: panic, fatal, error, warn, info, debug)\n\t",
			EnvVar: "LOG_LEVEL",
		},
		cli.StringFlag{
			Name:   "logfile",
			Value:  "stdout",
			Usage:  "filename where logs will be redirected (default stdout)\n\t",
			EnvVar: "LOG_FILENAME",
		},
		cli.BoolFlag{
			Name:   "no-folderconfig, nfc",
			Usage:  fmt.Sprintf("Do not read folder config file (%s)\n\t", xdsconfig.FoldersConfigFilename),
			EnvVar: "NO_FOLDERCONFIG",
		},
	}

	// only one action: Web Server
	app.Action = xdsApp

	app.Run(os.Args)
}
