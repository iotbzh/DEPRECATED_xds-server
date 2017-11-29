package xdsserver

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
)

// eventsList Registering for events that will be send over a WS
func (s *APIService) eventsList(c *gin.Context) {

}

// eventsRegister Registering for events that will be send over a WS
func (s *APIService) eventsRegister(c *gin.Context) {
	var args xsapiv1.EventRegisterArgs

	if c.BindJSON(&args) != nil {
		common.APIError(c, "Invalid arguments")
		return
	}

	sess := s.sessions.Get(c)
	if sess == nil {
		common.APIError(c, "Unknown sessions")
		return
	}

	evType := strings.TrimPrefix(xsapiv1.EVTFolderStateChange, xsapiv1.EventTypePrefix)
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

		if err := (*so).Emit(EVTAll, msg); err != nil {
			s.log.Errorf("WS Emit Event : %v", err)
		}

		if err := (*so).Emit(EventTypePrefix+ev.Type, msg); err != nil {
			s.log.Errorf("WS Emit Event : %v", err)
		}
	}

	data := make(st.EventsCBData)
	data["sid"] = sess.ID

	id, err := s.mfolders.SThg.Events.Register(args.Name, cbFunc, args.ProjectID, &data)
	*/

	var cbFunc FolderEventCB
	cbFunc = func(cfg *xsapiv1.FolderConfig, data *FolderEventCBData) {
		ssid := (*data)["sid"].(string)
		so := s.sessions.IOSocketGet(ssid)
		if so == nil {
			//s.log.Infof("Event %s not emitted - sid: %s", ev.Type, ssid)

			// Consider that client disconnected, so unregister this event
			// SEB FIXMEs.mfolders.RegisterEventChange(ev.Type)
			return
		}

		msg := xsapiv1.EventMsg{
			Time:   time.Now().String(),
			Type:   evType,
			Folder: *cfg,
		}

		s.Log.Debugf("WS Emit %s - Status=%10s, IsInSync=%6v, ID=%s",
			xsapiv1.EventTypePrefix+evType, cfg.Status, cfg.IsInSync, cfg.ID)

		if err := (*so).Emit(xsapiv1.EventTypePrefix+evType, msg); err != nil {
			s.Log.Errorf("WS Emit Folder StateChanged event : %v", err)
		}
	}
	data := make(FolderEventCBData)
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
	var args xsapiv1.EventUnRegisterArgs

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
