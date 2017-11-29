package xsapiv1

// SDK Define a cross tool chain used to build application
type SDK struct {
	ID      string `json:"id" binding:"required"`
	Name    string `json:"name"`
	Profile string `json:"profile"`
	Version string `json:"version"`
	Arch    string `json:"arch"`
	Path    string `json:"path"`

	// Not exported fields
	EnvFile string `json:"-"`
}
