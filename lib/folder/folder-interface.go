package folder

// FolderType definition
type FolderType string

const (
	TypePathMap   = "PathMap"
	TypeCloudSync = "CloudSync"
	TypeCifsSmb   = "CIFS"
)

// Folder Status definition
const (
	StatusErrorConfig = "ErrorConfig"
	StatusDisable     = "Disable"
	StatusEnable      = "Enable"
	StatusPause       = "Pause"
	StatusSyncing     = "Syncing"
)

type EventCBData map[string]interface{}
type EventCB func(cfg *FolderConfig, data *EventCBData)

// IFOLDER Folder interface
type IFOLDER interface {
	NewUID(suffix string) string                              // Get a new folder UUID
	Add(cfg FolderConfig) (*FolderConfig, error)              // Add a new folder
	GetConfig() FolderConfig                                  // Get folder public configuration
	GetFullPath(dir string) string                            // Get folder full path
	ConvPathCli2Svr(s string) string                          // Convert path from Client to Server
	ConvPathSvr2Cli(s string) string                          // Convert path from Server to Client
	Remove() error                                            // Remove a folder
	RegisterEventChange(cb *EventCB, data *EventCBData) error // Request events registration (sent through WS)
	UnRegisterEventChange() error                             // Un-register events
	Sync() error                                              // Force folder files synchronization
	IsInSync() (bool, error)                                  // Check if folder files are in-sync
}

// FolderConfig is the config for one folder
type FolderConfig struct {
	ID         string     `json:"id"`
	Label      string     `json:"label"`
	ClientPath string     `json:"path"`
	Type       FolderType `json:"type"`
	Status     string     `json:"status"`
	IsInSync   bool       `json:"isInSync"`
	DefaultSdk string     `json:"defaultSdk"`

	// Not exported fields from REST API point of view
	RootPath string `json:"-"`

	// FIXME: better to define an equivalent to union data and then implement
	// UnmarshalJSON/MarshalJSON to decode/encode according to Type value
	// Data interface{} `json:"data"`

	// Specific data depending on which Type is used
	DataPathMap   PathMapConfig   `json:"dataPathMap,omitempty"`
	DataCloudSync CloudSyncConfig `json:"dataCloudSync,omitempty"`
}

// PathMapConfig Path mapping specific data
type PathMapConfig struct {
	ServerPath string `json:"serverPath"`

	// Don't keep temporary file name (IOW we don't want to save it and reuse it)
	CheckFile    string `json:"checkFile" xml:"-"`
	CheckContent string `json:"checkContent" xml:"-"`
}

// CloudSyncConfig CloudSync (AKA Syncthing) specific data
type CloudSyncConfig struct {
	SyncThingID string `json:"syncThingID"`
}
