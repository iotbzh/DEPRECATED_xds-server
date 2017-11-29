package xsapiv1

// APIConfig parameters (json format) of /config command
type APIConfig struct {
	ServerUID        string          `json:"id"`
	Version          string          `json:"version"`
	APIVersion       string          `json:"apiVersion"`
	VersionGitTag    string          `json:"gitTag"`
	SupportedSharing map[string]bool `json:"supportedSharing"`
	Builder          BuilderConfig   `json:"builder"`
}

// BuilderConfig represents the builder container configuration
type BuilderConfig struct {
	IP          string `json:"ip"`
	Port        string `json:"port"`
	SyncThingID string `json:"syncThingID"`
}
