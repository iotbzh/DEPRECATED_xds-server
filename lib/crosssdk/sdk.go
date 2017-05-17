package crosssdk

import (
	"fmt"
	"path"
	"path/filepath"
)

// SDK Define a cross tool chain used to build application
type SDK struct {
	Profile string
	Version string
	Arch    string
	Path    string
	EnvFile string
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

// GetEnvCmd returns the command to initialized the environment to use a cross SDK
func (s *SDK) GetEnvCmd() string {
	return ". " + path.Join(s.Path, s.EnvFile)
}
