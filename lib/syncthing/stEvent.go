package st

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

// Events .
type Events struct {
	MonitorTime time.Duration
	Debug       bool

	stop  chan bool
	st    *SyncThing
	log   *logrus.Logger
	cbArr map[string][]cbMap
}

type Event struct {
	Type string            `json:"type"`
	Time time.Time         `json:"time"`
	Data map[string]string `json:"data"`
}

type EventsCBData map[string]interface{}
type EventsCB func(ev Event, cbData *EventsCBData)

const (
	EventFolderCompletion string = "FolderCompletion"
	EventFolderSummary    string = "FolderSummary"
	EventFolderPaused     string = "FolderPaused"
	EventFolderResumed    string = "FolderResumed"
	EventFolderErrors     string = "FolderErrors"
	EventStateChanged     string = "StateChanged"
)

var EventsAll string = EventFolderCompletion + "|" +
	EventFolderSummary + "|" +
	EventFolderPaused + "|" +
	EventFolderResumed + "|" +
	EventFolderErrors + "|" +
	EventStateChanged

type STEvent struct {
	// Per-subscription sequential event ID. Named "id" for backwards compatibility with the REST API
	SubscriptionID int `json:"id"`
	// Global ID of the event across all subscriptions
	GlobalID int                    `json:"globalID"`
	Time     time.Time              `json:"time"`
	Type     string                 `json:"type"`
	Data     map[string]interface{} `json:"data"`
}

type cbMap struct {
	id       int
	cb       EventsCB
	filterID string
	data     *EventsCBData
}

// NewEventListener Create a new instance of Event listener
func (s *SyncThing) NewEventListener() *Events {
	_, dbg := os.LookupEnv("XDS_DEBUG_STEVENTS") // set to add more debug log
	return &Events{
		MonitorTime: 100, // in Milliseconds
		Debug:       dbg,
		stop:        make(chan bool, 1),
		st:          s,
		log:         s.log,
		cbArr:       make(map[string][]cbMap),
	}
}

// Start starts event monitoring loop
func (e *Events) Start() error {
	go e.monitorLoop()
	return nil
}

// Stop stops event monitoring loop
func (e *Events) Stop() {
	e.stop <- true
}

// Register Add a listener on an event
func (e *Events) Register(evName string, cb EventsCB, filterID string, data *EventsCBData) (int, error) {
	if evName == "" || !strings.Contains(EventsAll, evName) {
		return -1, fmt.Errorf("Unknown event name")
	}
	if data == nil {
		data = &EventsCBData{}
	}

	cbList := []cbMap{}
	if _, ok := e.cbArr[evName]; ok {
		cbList = e.cbArr[evName]
	}

	id := len(cbList)
	(*data)["id"] = strconv.Itoa(id)

	e.cbArr[evName] = append(cbList, cbMap{id: id, cb: cb, filterID: filterID, data: data})

	return id, nil
}

// UnRegister Remove a listener event
func (e *Events) UnRegister(evName string, id int) error {
	cbKey, ok := e.cbArr[evName]
	if !ok {
		return fmt.Errorf("No event registered to such name")
	}

	// FIXME - NOT TESTED
	if id >= len(cbKey) {
		return fmt.Errorf("Invalid id")
	} else if id == len(cbKey) {
		e.cbArr[evName] = cbKey[:id-1]
	} else {
		e.cbArr[evName] = cbKey[id : id+1]
	}

	return nil
}

// GetEvents returns the Syncthing events
func (e *Events) getEvents(since int) ([]STEvent, error) {
	var data []byte
	ev := []STEvent{}
	url := "events"
	if since != -1 {
		url += "?since=" + strconv.Itoa(since)
	}
	if err := e.st.client.HTTPGet(url, &data); err != nil {
		return ev, err
	}
	err := json.Unmarshal(data, &ev)
	return ev, err
}

// Loop to monitor Syncthing events
func (e *Events) monitorLoop() {
	e.log.Infof("Event monitoring running...")
	since := 0
	for {
		select {
		case <-e.stop:
			e.log.Infof("Event monitoring exited")
			return

		case <-time.After(e.MonitorTime * time.Millisecond):
			stEvArr, err := e.getEvents(since)
			if err != nil {
				e.log.Errorf("Syncthing Get Events: %v", err)
				continue
			}
			// Process events
			for _, stEv := range stEvArr {
				since = stEv.SubscriptionID
				if e.Debug {
					e.log.Warnf("ST EVENT: %d %s\n  %v", stEv.GlobalID, stEv.Type, stEv)
				}

				cbKey, ok := e.cbArr[stEv.Type]
				if !ok {
					continue
				}

				evData := Event{
					Type: stEv.Type,
					Time: stEv.Time,
				}

				// Decode Events
				// FIXME: re-define data struct for each events
				// instead of map of string and use JSON marshing/unmarshing
				fID := ""
				evData.Data = make(map[string]string)
				switch stEv.Type {

				case EventFolderCompletion:
					fID = convString(stEv.Data["folder"])
					evData.Data["completion"] = convFloat64(stEv.Data["completion"])

				case EventFolderSummary:
					fID = convString(stEv.Data["folder"])
					evData.Data["needBytes"] = convInt64(stEv.Data["needBytes"])
					evData.Data["state"] = convString(stEv.Data["state"])

				case EventFolderPaused, EventFolderResumed:
					fID = convString(stEv.Data["id"])
					evData.Data["label"] = convString(stEv.Data["label"])

				case EventFolderErrors:
					fID = convString(stEv.Data["folder"])
					// TODO decode array evData.Data["errors"] = convString(stEv.Data["errors"])

				case EventStateChanged:
					fID = convString(stEv.Data["folder"])
					evData.Data["from"] = convString(stEv.Data["from"])
					evData.Data["to"] = convString(stEv.Data["to"])

				default:
					e.log.Warnf("Unsupported event type")
				}

				if fID != "" {
					evData.Data["id"] = fID
				}

				// Call all registered callbacks
				for _, c := range cbKey {
					if e.Debug {
						e.log.Warnf("EVENT CB fID=%s, filterID=%s", fID, c.filterID)
					}
					// Call when filterID is not set or when it matches
					if c.filterID == "" || (fID != "" && fID == c.filterID) {
						c.cb(evData, c.data)
					}
				}
			}
		}
	}
}

func convString(d interface{}) string {
	return d.(string)
}

func convFloat64(d interface{}) string {
	return strconv.FormatFloat(d.(float64), 'f', -1, 64)
}

func convInt64(d interface{}) string {
	return strconv.FormatInt(d.(int64), 10)
}
