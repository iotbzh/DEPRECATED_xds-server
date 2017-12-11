/*
 * Copyright (C) 2017 "IoT.bzh"
 * Author Sebastien Douheret <sebastien@iot.bzh>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package xdsserver

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
	"github.com/iotbzh/xds-server/lib/xsapiv1"

	"github.com/iotbzh/xds-server/lib/syncthing"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
)

const cookieMaxAge = "3600"

// Context holds the XDS server context
type Context struct {
	ProgName      string
	Cli           *cli.Context
	Config        *xdsconfig.Config
	Log           *logrus.Logger
	LogLevelSilly bool
	LogSillyf     func(format string, args ...interface{})
	SThg          *st.SyncThing
	SThgCmd       *exec.Cmd
	SThgInotCmd   *exec.Cmd
	mfolders      *Folders
	sdks          *SDKs
	WWWServer     *WebServer
	sessions      *Sessions
	Exit          chan os.Signal
}

// NewXdsServer Create a new instance of XDS server
func NewXdsServer(cliCtx *cli.Context) *Context {
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

	// Support silly logging (printed on log.debug)
	sillyVal, sillyLog := os.LookupEnv("XDS_LOG_SILLY")
	logSilly := sillyLog && sillyVal == "1"
	sillyFunc := func(format string, args ...interface{}) {
		if logSilly {
			log.Debugf("SILLY: "+format, args...)
		}
	}

	// Define default configuration
	ctx := Context{
		ProgName:      cliCtx.App.Name,
		Cli:           cliCtx,
		Log:           log,
		LogLevelSilly: logSilly,
		LogSillyf:     sillyFunc,
		Exit:          make(chan os.Signal, 1),
	}

	// register handler on SIGTERM / exit
	signal.Notify(ctx.Exit, os.Interrupt, syscall.SIGTERM)
	go handlerSigTerm(&ctx)

	return &ctx
}

// Run Main function called to run XDS Server
func (ctx *Context) Run() (int, error) {
	var err error

	// Logs redirected into a file when logfile option or logsDir config is set
	ctx.Config.LogVerboseOut = os.Stderr
	if ctx.Config.FileConf.LogsDir != "" {
		if ctx.Config.Options.LogFile != "stdout" {
			logFile := ctx.Config.Options.LogFile

			fdL, err := os.OpenFile(logFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
			if err != nil {
				msgErr := fmt.Sprintf("Cannot create log file %s", logFile)
				return int(syscall.EPERM), fmt.Errorf(msgErr)
			}
			ctx.Log.Out = fdL

			ctx._logPrint("Logging file: %s\n", logFile)
		}

		logFileHTTPReq := filepath.Join(ctx.Config.FileConf.LogsDir, "xds-server-verbose.log")
		fdLH, err := os.OpenFile(logFileHTTPReq, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
		if err != nil {
			msgErr := fmt.Sprintf("Cannot create log file %s", logFileHTTPReq)
			return int(syscall.EPERM), fmt.Errorf(msgErr)
		}
		ctx.Config.LogVerboseOut = fdLH

		ctx._logPrint("Logging file for HTTP requests:  %s\n", logFileHTTPReq)
	}

	// Create syncthing instance when section "syncthing" is present in server-config.json
	if ctx.Config.FileConf.SThgConf != nil {
		ctx.SThg = st.NewSyncThing(ctx.Config, ctx.Log)
	}

	// Start local instance of Syncthing and Syncthing-notify
	if ctx.SThg != nil {
		ctx.Log.Infof("Starting Syncthing...")
		ctx.SThgCmd, err = ctx.SThg.Start()
		if err != nil {
			return -4, err
		}
		ctx._logPrint("Syncthing started (PID %d)\n", ctx.SThgCmd.Process.Pid)

		ctx.Log.Infof("Starting Syncthing-inotify...")
		ctx.SThgInotCmd, err = ctx.SThg.StartInotify()
		if err != nil {
			return -4, err
		}
		ctx._logPrint("Syncthing-inotify started (PID %d)\n", ctx.SThgInotCmd.Process.Pid)

		// Establish connection with local Syncthing (retry if connection fail)
		ctx._logPrint("Establishing connection with Syncthing...\n")
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
			return -4, err
		}

		// FIXME: do we still need Builder notion ? if no cleanup
		if ctx.Config.Builder, err = xdsconfig.NewBuilderConfig(ctx.SThg.MyID); err != nil {
			return -4, err
		}
		ctx.Config.SupportedSharing[xsapiv1.TypeCloudSync] = true
	}

	// Init model folder
	ctx.mfolders = FoldersNew(ctx)

	// Load initial folders config from disk
	if err := ctx.mfolders.LoadConfig(); err != nil {
		return -5, err
	}

	// Init cross SDKs
	ctx.sdks, err = NewSDKs(ctx)
	if err != nil {
		return -6, err
	}

	// Create Web Server
	ctx.WWWServer = NewWebServer(ctx)

	// Sessions manager
	ctx.sessions = NewClientSessions(ctx, cookieMaxAge)

	// Run Web Server until exit requested (blocking call)
	if err = ctx.WWWServer.Serve(); err != nil {
		ctx.Log.Println(err)
		return -7, err
	}

	return -99, fmt.Errorf("Program exited ")
}

// Helper function to log message on both stdout and logger
func (ctx *Context) _logPrint(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	if ctx.Log.Out != os.Stdout {
		ctx.Log.Infof(format, args...)
	}
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
