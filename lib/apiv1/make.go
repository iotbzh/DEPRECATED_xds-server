package apiv1

import (
	"net/http"

	"time"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-server/lib/common"
)

// MakeArgs is the parameters (json format) of /make command
type MakeArgs struct {
	ID         string `json:"id"`
	RPath      string `json:"rpath"`   // relative path into project
	Args       string `json:"args"`    // args to pass to make command
	SdkID      string `json:"sdkid"`   // sdk ID to use for setting env
	CmdTimeout int    `json:"timeout"` // command completion timeout in Second
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

func (s *APIService) buildMake(c *gin.Context) {
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
	id := c.Param("id")
	if id == "" {
		id = args.ID
	}
	if id == "" {
		common.APIError(c, "Invalid id")
		return
	}

	prj := s.mfolder.GetFolderFromID(id)
	if prj == nil {
		common.APIError(c, "Unknown id")
		return
	}

	execTmo := args.CmdTimeout
	if execTmo == 0 {
		// TODO get default timeout from config.json file
		execTmo = 24 * 60 * 60 // 1 day
	}

	cmd := "cd " + prj.GetFullPath(args.RPath) + " && make"
	if args.Args != "" {
		cmd += " " + args.Args
	}

	// Define callback for output
	var oCB common.EmitOutputCB
	oCB = func(sid string, id int, stdout, stderr string) {
		// IO socket can be nil when disconnected
		so := s.sessions.IOSocketGet(sid)
		if so == nil {
			s.log.Infof("%s not emitted: WS closed - sid: %s - msg id:%d", MakeOutEvent, sid, id)
			return
		}
		s.log.Debugf("%s emitted - WS sid %s - id:%d", MakeOutEvent, sid, id)

		// FIXME replace by .BroadcastTo a room
		err := (*so).Emit(MakeOutEvent, MakeOutMsg{
			CmdID:     strconv.Itoa(id),
			Timestamp: time.Now().String(),
			Stdout:    stdout,
			Stderr:    stderr,
		})
		if err != nil {
			s.log.Errorf("WS Emit : %v", err)
		}
	}

	// Define callback for output
	eCB := func(sid string, id int, code int, err error) {
		s.log.Debugf("Command [Cmd ID %d] exited: code %d, error: %v", id, code, err)

		// IO socket can be nil when disconnected
		so := s.sessions.IOSocketGet(sid)
		if so == nil {
			s.log.Infof("%s not emitted - WS closed (id:%d", MakeExitEvent, id)
			return
		}

		// FIXME replace by .BroadcastTo a room
		e := (*so).Emit(MakeExitEvent, MakeExitMsg{
			CmdID:     strconv.Itoa(id),
			Timestamp: time.Now().String(),
			Code:      code,
			Error:     err,
		})
		if e != nil {
			s.log.Errorf("WS Emit : %v", e)
		}
	}

	cmdID := makeCommandID
	makeCommandID++

	// Retrieve env command regarding Sdk ID
	if envCmd := s.sdks.GetEnvCmd(args.SdkID, prj.DefaultSdk); envCmd != "" {
		cmd = envCmd + " && " + cmd
	}

	s.log.Debugf("Execute [Cmd ID %d]: %v", cmdID, cmd)
	err := common.ExecPipeWs(cmd, sop, sess.ID, cmdID, execTmo, s.log, oCB, eCB)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK,
		gin.H{
			"status": "OK",
			"cmdID":  cmdID,
		})
}
