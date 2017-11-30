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
	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
)

/* TODO: Deprecated - should be removed
// MakeArgs is the parameters (json format) of /make command
type MakeArgs struct {
	ID            string   `json:"id"`
	SdkID         string   `json:"sdkID"` // sdk ID to use for setting env
	CmdID         string   `json:"cmdID"` // command unique ID
	Args          []string `json:"args"`  // args to pass to make command
	Env           []string `json:"env"`
	RPath         string   `json:"rpath"`         // relative path into project
	ExitImmediate bool     `json:"exitImmediate"` // when true, exit event sent immediately when command exited (IOW, don't wait file synchronization)
	CmdTimeout    int      `json:"timeout"`       // command completion timeout in Second
}

// MakeOutMsg Message send on each output (stdout+stderr) of make command
type MakeOutMsg struct {
	CmdID     string `json:"cmdID"`
	Timestamp string `json:"timestamp"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
}

// MakeExitMsg Message send on make command exit
type MakeExitMsg struct {
	CmdID     string `json:"cmdID"`
	Timestamp string `json:"timestamp"`
	Code      int    `json:"code"`
	Error     error  `json:"error"`
}

// MakeOutEvent Event send in WS when characters are received on stdout/stderr
const MakeOutEvent = "make:output"

// MakeExitEvent Event send in WS when command exited
const MakeExitEvent = "make:exit"

var makeCommandID = 1
*/

func (s *APIService) buildMake(c *gin.Context) {
	common.APIError(c, "/make route is not longer supported, use /exec instead")

	/*
		var args MakeArgs

		if c.BindJSON(&args) != nil {
			common.APIError(c, "Invalid arguments")
			return
		}

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

		// Allow to pass id in url (/make/:id) or as JSON argument
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
		pf := s.mfolders.Get(id)
		if pf == nil {
			common.APIError(c, "Unknown id")
			return
		}
		folder := *pf
		prj := folder.GetConfig()

		execTmo := args.CmdTimeout
		if execTmo == 0 {
			// TODO get default timeout from config.json file
			execTmo = 24 * 60 * 60 // 1 day
		}

		// TODO merge all code below with exec.go

		// Define callback for output
		var oCB common.EmitOutputCB
		oCB = func(sid string, cmdID string, stdout, stderr string, data *map[string]interface{}) {
			// IO socket can be nil when disconnected
			so := s.sessions.IOSocketGet(sid)
			if so == nil {
				s.log.Infof("%s not emitted: WS closed - sid: %s - msg id:%s", MakeOutEvent, sid, cmdID)
				return
			}

			// Retrieve project ID and RootPath
			prjID := (*data)["ID"].(string)
			prjRootPath := (*data)["RootPath"].(string)

			// Cleanup any references to internal rootpath in stdout & stderr
			stdout = strings.Replace(stdout, prjRootPath, "", -1)
			stderr = strings.Replace(stderr, prjRootPath, "", -1)

			s.log.Debugf("%s emitted - WS sid %s - id:%d - prjID:%s", MakeOutEvent, sid, id, prjID)

			// FIXME replace by .BroadcastTo a room
			err := (*so).Emit(MakeOutEvent, MakeOutMsg{
				CmdID:     cmdID,
				Timestamp: time.Now().String(),
				Stdout:    stdout,
				Stderr:    stderr,
			})
			if err != nil {
				s.log.Errorf("WS Emit : %v", err)
			}
		}

		// Define callback for output
		eCB := func(sid string, cmdID string, code int, err error, data *map[string]interface{}) {
			s.log.Debugf("Command [Cmd ID %s] exited: code %d, error: %v", cmdID, code, err)

			// IO socket can be nil when disconnected
			so := s.sessions.IOSocketGet(sid)
			if so == nil {
				s.log.Infof("%s not emitted - WS closed (id:%s", MakeExitEvent, cmdID)
				return
			}

			// Retrieve project ID and RootPath
			prjID := (*data)["ID"].(string)
			exitImm := (*data)["ExitImmediate"].(bool)

			// XXX - workaround to be sure that Syncthing detected all changes
			if err := s.mfolders.ForceSync(prjID); err != nil {
				s.log.Errorf("Error while syncing folder %s: %v", prjID, err)
			}
			if !exitImm {
				// Wait end of file sync
				// FIXME pass as argument
				tmo := 60
				for t := tmo; t > 0; t-- {
					s.log.Debugf("Wait file insync for %s (%d/%d)", prjID, t, tmo)
					if sync, err := s.mfolders.IsFolderInSync(prjID); sync || err != nil {
						if err != nil {
							s.log.Errorf("ERROR IsFolderInSync (%s): %v", prjID, err)
						}
						break
					}
					time.Sleep(time.Second)
				}
			}

			// FIXME replace by .BroadcastTo a room
			e := (*so).Emit(MakeExitEvent, MakeExitMsg{
				CmdID:     id,
				Timestamp: time.Now().String(),
				Code:      code,
				Error:     err,
			})
			if e != nil {
				s.log.Errorf("WS Emit : %v", e)
			}
		}

		// Unique ID for each commands
		if args.CmdID == "" {
			args.CmdID = s.cfg.ServerUID[:18] + "_" + strconv.Itoa(makeCommandID)
			makeCommandID++
		}
		cmd := []string{}

		// Retrieve env command regarding Sdk ID
		if envCmd := s.sdks.GetEnvCmd(args.SdkID, prj.DefaultSdk); len(envCmd) > 0 {
			cmd = append(cmd, envCmd...)
			cmd = append(cmd, "&&")
		}

		cmd = append(cmd, "cd", folder.GetFullPath(args.RPath), "&&", "make")
		if len(args.Args) > 0 {
			cmd = append(cmd, args.Args...)
		}

		s.log.Debugf("Execute [Cmd ID %d]: %v", args.CmdID, cmd)

		data := make(map[string]interface{})
		data["ID"] = prj.ID
		data["RootPath"] = prj.RootPath
		data["ExitImmediate"] = args.ExitImmediate

		err = common.ExecPipeWs(cmd, args.Env, sop, sess.ID, args.CmdID, execTmo, s.log, oCB, eCB, &data)
		if err != nil {
			common.APIError(c, err.Error())
			return
		}

		c.JSON(http.StatusOK,
			gin.H{
				"status": "OK",
				"cmdID":  args.CmdID,
			})
	*/
}
