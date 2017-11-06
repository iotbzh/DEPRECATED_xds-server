package apiv1

import (
	"net/http"
	"time"

	"github.com/iotbzh/xds-server/lib/folder"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
)

// EventArgs is the parameters (json format) of /events/register command
type EventRegisterArgs struct {
	Name      string `json:"name"`
	ProjectID string `json:"filterProjectID"`
}

type EventUnRegisterArgs struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// EventMsg Message send
type EventMsg struct {
	Time   string              `json:"time"`
	Type   string              `json:"type"`
	Folder folder.FolderConfig `json:"folder"`
}

// EventEvent Event send in WS when an internal event (eg. Syncthing event is received)
const EventEventAll = "event:all"
const EventEventType = "event:" // following by event type

// eventsList Registering for events that will be send over a WS
func (s *APIService) eventsList(c *gin.Context) {

}

// eventsRegister Registering for events that will be send over a WS
func (s *APIService) eventsRegister(c *gin.Context) {
	var args EventRegisterArgs

	if c.BindJSON(&args) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	sess := s.sessions.Get(c)
	if sess == nil {
		common.APIError(c, "Unknown sessions")
		return
	}

	evType := "FolderStateChanged"
	if args.Name != evType {
		common.APIError(c, "Unsupported event name")
		return
	}

	/* XXX - to be removed if no plan to support "generic" event
	var cbFunc st.EventsCB
	cbFunc = func(ev st.Event, data *st.EventsCBData) {

		evid, _ := strconv.Atoi((*data)["id"].(string))
		ssid := (*data)["sid"].(string)
		so := s.sessions.IOSocketGet(ssid)
		if so == nil {
			s.log.Infof("Event %s not emitted - sid: %s", ev.Type, ssid)

			// Consider that client disconnected, so unregister this event
			s.mfolders.SThg.Events.UnRegister(ev.Type, evid)
			return
		}

		msg := EventMsg{
			Time: ev.Time,
			Type: ev.Type,
			Data: ev.Data,
		}

		if err := (*so).Emit(EventEventAll, msg); err != nil {
			s.log.Errorf("WS Emit Event : %v", err)
		}

		if err := (*so).Emit(EventEventType+ev.Type, msg); err != nil {
			s.log.Errorf("WS Emit Event : %v", err)
		}
	}

	data := make(st.EventsCBData)
	data["sid"] = sess.ID

	id, err := s.mfolders.SThg.Events.Register(args.Name, cbFunc, args.ProjectID, &data)
	*/

	var cbFunc folder.EventCB
	cbFunc = func(cfg *folder.FolderConfig, data *folder.EventCBData) {
		ssid := (*data)["sid"].(string)
		so := s.sessions.IOSocketGet(ssid)
		if so == nil {
			//s.log.Infof("Event %s not emitted - sid: %s", ev.Type, ssid)

			// Consider that client disconnected, so unregister this event
			// SEB FIXMEs.mfolders.RegisterEventChange(ev.Type)
			return
		}

		msg := EventMsg{
			Time:   time.Now().String(),
			Type:   evType,
			Folder: *cfg,
		}

		s.log.Debugf("WS Emit %s - Status=%10s, IsInSync=%6v, ID=%s",
			EventEventType+evType, cfg.Status, cfg.IsInSync, cfg.ID)

		if err := (*so).Emit(EventEventType+evType, msg); err != nil {
			s.log.Errorf("WS Emit Folder StateChanged event : %v", err)
		}
	}
	data := make(folder.EventCBData)
	data["sid"] = sess.ID

	prjID, err := s.mfolders.ResolveID(args.ProjectID)
	if err != nil {
		common.APIError(c, err.Error())
		return
	}
	if err = s.mfolders.RegisterEventChange(prjID, &cbFunc, &data); err != nil {
		common.APIError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// eventsRegister Registering for events that will be send over a WS
func (s *APIService) eventsUnRegister(c *gin.Context) {
	var args EventUnRegisterArgs

	if c.BindJSON(&args) != nil || args.Name == "" || args.ID < 0 {
		common.APIError(c, "Invalid arguments")
		return
	}
	/* TODO
	if err := s.mfolders.SThg.Events.UnRegister(args.Name, args.ID); err != nil {
		common.APIError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
	*/
	common.APIError(c, "Not implemented yet")
}
