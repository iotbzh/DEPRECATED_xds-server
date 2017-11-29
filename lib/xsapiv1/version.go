package xsapiv1

// XDS server Version
type Version struct {
	ID            string `json:"id"`
	Version       string `json:"version"`
	APIVersion    string `json:"apiVersion"`
	VersionGitTag string `json:"gitTag"`
}
