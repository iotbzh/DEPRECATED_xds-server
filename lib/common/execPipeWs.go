package common

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/googollee/go-socket.io"
)

// EmitOutputCB is the function callback used to emit data
type EmitOutputCB func(sid string, cmdID int, stdout, stderr string)

// EmitExitCB is the function callback used to emit exit proc code
type EmitExitCB func(sid string, cmdID int, code int, err error)

// Inspired by :
// https://github.com/gorilla/websocket/blob/master/examples/command/main.go

// ExecPipeWs executes a command and redirect stdout/stderr into a WebSocket
func ExecPipeWs(cmd string, so *socketio.Socket, sid string, cmdID int,
	cmdExecTimeout int, log *logrus.Logger, eoCB EmitOutputCB, eeCB EmitExitCB) error {

	outr, outw, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("Pipe stdout error: " + err.Error())
	}

	// XXX - do we need to pipe stdin one day ?
	inr, inw, err := os.Pipe()
	if err != nil {
		outr.Close()
		outw.Close()
		return fmt.Errorf("Pipe stdin error: " + err.Error())
	}

	bashArgs := []string{"/bin/bash", "-c", cmd}
	proc, err := os.StartProcess("/bin/bash", bashArgs, &os.ProcAttr{
		Files: []*os.File{inr, outw, outw},
	})
	if err != nil {
		outr.Close()
		outw.Close()
		inr.Close()
		inw.Close()
		return fmt.Errorf("Process start error: " + err.Error())
	}

	go func() {
		defer outr.Close()
		defer outw.Close()
		defer inr.Close()
		defer inw.Close()

		stdoutDone := make(chan struct{})
		go cmdPumpStdout(so, outr, stdoutDone, sid, cmdID, log, eoCB)

		// Blocking function that poll input or wait for end of process
		cmdPumpStdin(so, inw, proc, sid, cmdID, cmdExecTimeout, log, eeCB)

		// Some commands will exit when stdin is closed.
		inw.Close()

		defer outr.Close()

		if status, err := proc.Wait(); err == nil {
			// Other commands need a bonk on the head.
			if !status.Exited() {
				if err := proc.Signal(os.Interrupt); err != nil {
					log.Errorln("Proc interrupt:", err)
				}

				select {
				case <-stdoutDone:
				case <-time.After(time.Second):
					// A bigger bonk on the head.
					if err := proc.Signal(os.Kill); err != nil {
						log.Errorln("Proc term:", err)
					}
					<-stdoutDone
				}
			}
		}
	}()

	return nil
}

func cmdPumpStdin(so *socketio.Socket, w io.Writer, proc *os.Process,
	sid string, cmdID int, tmo int, log *logrus.Logger, exitFuncCB EmitExitCB) {
	/* XXX - code to add to support stdin through WS
	for {
		_, message, err := so. ?? ReadMessage()
		if err != nil {
			break
		}
		message = append(message, '\n')
		if _, err := w.Write(message); err != nil {
			break
		}
	}
	*/

	// Monitor process exit
	type DoneChan struct {
		status int
		err    error
	}
	done := make(chan DoneChan, 1)
	go func() {
		status := 0
		sts, err := proc.Wait()
		if !sts.Success() {
			s := sts.Sys().(syscall.WaitStatus)
			status = s.ExitStatus()
		}
		done <- DoneChan{status, err}
	}()

	// Wait cmd complete
	select {
	case dC := <-done:
		exitFuncCB(sid, cmdID, dC.status, dC.err)
	case <-time.After(time.Duration(tmo) * time.Second):
		exitFuncCB(sid, cmdID, -99,
			fmt.Errorf("Exit Timeout for command ID %v", cmdID))
	}
}

func cmdPumpStdout(so *socketio.Socket, r io.Reader, done chan struct{},
	sid string, cmdID int, log *logrus.Logger, emitFuncCB EmitOutputCB) {
	defer func() {
	}()

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		emitFuncCB(sid, cmdID, string(sc.Bytes()), "")
	}
	if sc.Err() != nil {
		log.Errorln("scan:", sc.Err())
	}
	close(done)
}
