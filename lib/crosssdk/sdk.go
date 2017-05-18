package crosssdk

import (
	"fmt"
	"path/filepath"
)

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

// NewCrossSDK creates a new instance of Syncthing
func NewCrossSDK(path string) (*SDK, error) {
	// Assume that we have .../<profile>/<version>/<arch>
	s := SDK{Path: path}

	s.Arch = filepath.Base(path)

	d := filepath.Dir(path)
	s.Version = filepath.Base(d)

	d = filepath.Dir(d)
	s.Profile = filepath.Base(d)

	s.ID = s.Profile + "_" + s.Arch + "_" + s.Version
	s.Name = s.Arch + "   (" + s.Version + ")"

	envFile := filepath.Join(path, "environment-setup*")
	ef, err := filepath.Glob(envFile)
	if err != nil {
		return nil, fmt.Errorf("Cannot retrieve environment setup file: %v", err)
	}
	if len(ef) != 1 {
		return nil, fmt.Errorf("No environment setup file found match %s", envFile)
	}
	s.EnvFile = ef[0]

	return &s, nil
}

// GetEnvCmd returns the command used to initialized the environment
func (s *SDK) GetEnvCmd() string {
	return ". " + s.EnvFile
}
