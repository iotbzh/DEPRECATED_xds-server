package xsapiv1

// EventRegisterArgs Parameters (json format) of /events/register command
type EventRegisterArgs struct {
	Name      string `json:"name"`
	ProjectID string `json:"filterProjectID"`
}

// EventUnRegisterArgs Parameters of /events/unregister command
type EventUnRegisterArgs struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// EventMsg Message send
type EventMsg struct {
	Time   string       `json:"time"`
	Type   string       `json:"type"`
	Folder FolderConfig `json:"folder"`
}

// EventEvent Event send in WS when an internal event (eg. Syncthing event is received)
const (
	// EventTypePrefix Used as event prefix
	EventTypePrefix = "event:" // following by event type

	// Supported Events type
	EVTAll               = EventTypePrefix + "all"
	EVTFolderChange      = EventTypePrefix + "folder-change"       // type EventMsg with Data type xsapiv1.???
	EVTFolderStateChange = EventTypePrefix + "folder-state-change" // type EventMsg with Data type xsapiv1.???
)
