// TODO add Doc
//
package main

import (
	"log"
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

// Web server main routine
func webServer(ctx *cli.Context) error {

	// Init config
	cfg, err := xdsconfig.Init(ctx)
	if err != nil {
		return cli.NewExitError(err, 2)
	}

	// Create and start Web Server
	svr := xdsserver.NewServer(cfg)
	if err = svr.Serve(); err != nil {
		log.Println(err)
		return cli.NewExitError(err, 3)
	}

	return cli.NewExitError("Program exited ", 4)
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
	}

	// only one action: Web Server
	app.Action = webServer

	app.Run(os.Args)
}
