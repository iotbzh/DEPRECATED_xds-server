// TODO add Doc
//
package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
	"github.com/iotbzh/xds-server/lib/xdsserver"
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

// XDS Server application main routine
func xdsApp(cliCtx *cli.Context) error {
	var err error

	// Create XDS server context
	ctxSvr := xdsserver.NewXdsServer(cliCtx)

	// Load config
	ctxSvr.Config, err = xdsconfig.Init(cliCtx, ctxSvr.Log)
	if err != nil {
		return cli.NewExitError(err, -2)
	}

	// Run XDS Server (main loop)
	errCode, err := ctxSvr.Run()

	return cli.NewExitError(err, errCode)
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
