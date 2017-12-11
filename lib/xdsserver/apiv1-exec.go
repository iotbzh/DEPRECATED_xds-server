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
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-common/golib/eows"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
	"github.com/kr/pty"
)

var execCommandID = 1

// ExecCmd executes remotely a command
func (s *APIService) execCmd(c *gin.Context) {
	var gdbPty, gdbTty *os.File
	var err error
	var args xsapiv1.ExecArgs
	if c.BindJSON(&args) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	// TODO: add permission ?

	// Retrieve session info
	sess := s.sessions.Get(c)
	if sess == nil {
		common.APIError(c, "Unknown sessions")
		return
	}
	sop := sess.IOSocket
	if sop == nil {
		common.APIError(c, "Websocket not established")
		return
	}

	// Allow to pass id in url (/exec/:id) or as JSON argument
	idArg := c.Param("id")
	if idArg == "" {
		idArg = args.ID
	}
	if idArg == "" {
		common.APIError(c, "Invalid id")
		return
	}
	id, err := s.mfolders.ResolveID(idArg)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}
	f := s.mfolders.Get(id)
	if f == nil {
		common.APIError(c, "Unknown id")
		return
	}
	fld := *f
	prj := fld.GetConfig()

	// Build command line
	cmd := []string{}
	// Setup env var regarding Sdk ID (used for example to setup cross toolchain)
	if envCmd := s.sdks.GetEnvCmd(args.SdkID, prj.DefaultSdk); len(envCmd) > 0 {
		cmd = append(cmd, envCmd...)
		cmd = append(cmd, "&&")
	} else {
		// It's an error if no envcmd found while a sdkid has been provided
		if args.SdkID != "" {
			common.APIError(c, "Unknown sdkid")
			return
		}
	}

	cmd = append(cmd, "cd", "\""+fld.GetFullPath(args.RPath)+"\"")
	// FIXME - add 'exec' prevents to use syntax:
	//       xds-exec -l debug -c xds-config.env -- "cd build && cmake .."
	//  but exec is mandatory to allow to pass correctly signals
	//  As workaround, exec is set for now on client side (eg. in xds-gdb)
	//cmd = append(cmd, "&&", "exec", args.Cmd)
	cmd = append(cmd, "&&", args.Cmd)

	// Process command arguments
	cmdArgs := make([]string, len(args.Args)+1)

	// Copy and Translate path from client to server
	for _, aa := range args.Args {
		if strings.Contains(aa, prj.ClientPath) {
			cmdArgs = append(cmdArgs, fld.ConvPathCli2Svr(aa))
		} else {
			cmdArgs = append(cmdArgs, aa)
		}
	}

	// Allocate pts if tty if used
	if args.TTY {
		gdbPty, gdbTty, err = pty.Open()
		if err != nil {
			common.APIError(c, err.Error())
			return
		}

		s.Log.Debugf("Client command tty: %v %v\n", gdbTty.Name(), gdbTty.Name())
		cmdArgs = append(cmdArgs, "--tty="+gdbTty.Name())
	}

	// Unique ID for each commands
	if args.CmdID == "" {
		args.CmdID = s.Config.ServerUID[:18] + "_" + strconv.Itoa(execCommandID)
		execCommandID++
	}

	// Create new execution over WS context
	execWS := eows.New(strings.Join(cmd, " "), cmdArgs, sop, sess.ID, args.CmdID)
	execWS.Log = s.Log

	// Append client project dir to environment
	execWS.Env = append(args.Env, "CLIENT_PROJECT_DIR="+prj.ClientPath)

	// Set command execution timeout
	if args.CmdTimeout == 0 {
		// 0 : default timeout
		// TODO get default timeout from server-config.json file
		execWS.CmdExecTimeout = 24 * 60 * 60 // 1 day
	} else {
		execWS.CmdExecTimeout = args.CmdTimeout
	}

	// Define callback for input (stdin)
	execWS.InputEvent = xsapiv1.ExecInEvent
	execWS.InputCB = func(e *eows.ExecOverWS, stdin string) (string, error) {
		s.Log.Debugf("STDIN <<%v>>", strings.Replace(stdin, "\n", "\\n", -1))

		// Handle Ctrl-D
		if len(stdin) == 1 && stdin == "\x04" {
			// Close stdin
			errMsg := fmt.Errorf("close stdin: %v", stdin)
			return "", errMsg
		}

		// Set correct path
		data := e.UserData
		prjID := (*data)["ID"].(string)
		f := s.mfolders.Get(prjID)
		if f == nil {
			s.Log.Errorf("InputCB: Cannot get folder ID %s", prjID)
		} else {
			// Translate paths from client to server
			stdin = (*f).ConvPathCli2Svr(stdin)
		}

		return stdin, nil
	}

	// Define callback for output (stdout+stderr)
	execWS.OutputCB = func(e *eows.ExecOverWS, stdout, stderr string) {
		// IO socket can be nil when disconnected
		so := s.sessions.IOSocketGet(e.Sid)
		if so == nil {
			s.Log.Infof("%s not emitted: WS closed (sid:%s, msgid:%s)", xsapiv1.ExecOutEvent, e.Sid, e.CmdID)
			return
		}

		// Retrieve project ID and RootPath
		data := e.UserData
		prjID := (*data)["ID"].(string)
		gdbServerTTY := (*data)["gdbServerTTY"].(string)

		f := s.mfolders.Get(prjID)
		if f == nil {
			s.Log.Errorf("OutputCB: Cannot get folder ID %s", prjID)
		} else {
			// Translate paths from server to client
			stdout = (*f).ConvPathSvr2Cli(stdout)
			stderr = (*f).ConvPathSvr2Cli(stderr)
		}

		s.Log.Debugf("%s emitted - WS sid[4:] %s - id:%s - prjID:%s", xsapiv1.ExecOutEvent, e.Sid[4:], e.CmdID, prjID)
		if stdout != "" {
			s.Log.Debugf("STDOUT <<%v>>", strings.Replace(stdout, "\n", "\\n", -1))
		}
		if stderr != "" {
			s.Log.Debugf("STDERR <<%v>>", strings.Replace(stderr, "\n", "\\n", -1))
		}

		// FIXME replace by .BroadcastTo a room
		err := (*so).Emit(xsapiv1.ExecOutEvent, xsapiv1.ExecOutMsg{
			CmdID:     e.CmdID,
			Timestamp: time.Now().String(),
			Stdout:    stdout,
			Stderr:    stderr,
		})
		if err != nil {
			s.Log.Errorf("WS Emit : %v", err)
		}

		// XXX - Workaround due to gdbserver bug that doesn't redirect
		// inferior output (https://bugs.eclipse.org/bugs/show_bug.cgi?id=437532#c13)
		if gdbServerTTY == "workaround" && len(stdout) > 1 && stdout[0] == '&' {

			// Extract and cleanup string like &"bla bla\n"
			re := regexp.MustCompile("&\"(.*)\"")
			rer := re.FindAllStringSubmatch(stdout, -1)
			out := ""
			if rer != nil && len(rer) > 0 {
				for _, o := range rer {
					if len(o) >= 1 {
						out = strings.Replace(o[1], "\\n", "\n", -1)
						out = strings.Replace(out, "\\r", "\r", -1)
						out = strings.Replace(out, "\\t", "\t", -1)

						s.Log.Debugf("STDOUT INFERIOR: <<%v>>", out)
						err := (*so).Emit(xsapiv1.ExecInferiorOutEvent, xsapiv1.ExecOutMsg{
							CmdID:     e.CmdID,
							Timestamp: time.Now().String(),
							Stdout:    out,
							Stderr:    "",
						})
						if err != nil {
							s.Log.Errorf("WS Emit : %v", err)
						}
					}
				}
			} else {
				s.Log.Errorf("INFERIOR out parsing error: stdout=<%v>", stdout)
			}
		}
	}

	// Define callback for output
	execWS.ExitCB = func(e *eows.ExecOverWS, code int, err error) {
		s.Log.Debugf("Command [Cmd ID %s] exited: code %d, error: %v", e.CmdID, code, err)

		// Close client tty
		defer func() {
			if gdbPty != nil {
				gdbPty.Close()
			}
			if gdbTty != nil {
				gdbTty.Close()
			}
		}()

		// IO socket can be nil when disconnected
		so := s.sessions.IOSocketGet(e.Sid)
		if so == nil {
			s.Log.Infof("%s not emitted - WS closed (id:%s)", xsapiv1.ExecExitEvent, e.CmdID)
			return
		}

		// Retrieve project ID and RootPath
		data := e.UserData
		prjID := (*data)["ID"].(string)
		exitImm := (*data)["ExitImmediate"].(bool)

		// XXX - workaround to be sure that Syncthing detected all changes
		if err := s.mfolders.ForceSync(prjID); err != nil {
			s.Log.Errorf("Error while syncing folder %s: %v", prjID, err)
		}
		if !exitImm {
			// Wait end of file sync
			// FIXME pass as argument
			tmo := 60
			for t := tmo; t > 0; t-- {
				s.Log.Debugf("Wait file in-sync for %s (%d/%d)", prjID, t, tmo)
				if sync, err := s.mfolders.IsFolderInSync(prjID); sync || err != nil {
					if err != nil {
						s.Log.Errorf("ERROR IsFolderInSync (%s): %v", prjID, err)
					}
					break
				}
				time.Sleep(time.Second)
			}
			s.Log.Debugf("OK file are synchronized.")
		}

		// FIXME replace by .BroadcastTo a room
		errSoEmit := (*so).Emit(xsapiv1.ExecExitEvent, xsapiv1.ExecExitMsg{
			CmdID:     e.CmdID,
			Timestamp: time.Now().String(),
			Code:      code,
			Error:     err,
		})
		if errSoEmit != nil {
			s.Log.Errorf("WS Emit : %v", errSoEmit)
		}
	}

	// User data (used within callbacks)
	data := make(map[string]interface{})
	data["ID"] = prj.ID
	data["ExitImmediate"] = args.ExitImmediate
	if args.TTY && args.TTYGdbserverFix {
		data["gdbServerTTY"] = "workaround"
	} else {
		data["gdbServerTTY"] = ""
	}
	execWS.UserData = &data

	// Start command execution
	s.Log.Infof("Execute [Cmd ID %s]: %v %v", execWS.CmdID, execWS.Cmd, execWS.Args)

	err = execWS.Start()
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, xsapiv1.ExecResult{Status: "OK", CmdID: execWS.CmdID})
}

// ExecCmd executes remotely a command
func (s *APIService) execSignalCmd(c *gin.Context) {
	var args xsapiv1.ExecSignalArgs

	if c.BindJSON(&args) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	s.Log.Debugf("Signal %s for command ID %s", args.Signal, args.CmdID)

	e := eows.GetEows(args.CmdID)
	if e == nil {
		common.APIError(c, "unknown cmdID")
		return
	}

	err := e.Signal(args.Signal)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, xsapiv1.ExecSigResult{Status: "OK", CmdID: args.CmdID})
}
