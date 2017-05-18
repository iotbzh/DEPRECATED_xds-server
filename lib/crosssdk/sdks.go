package crosssdk

import (
	"path"
	"path/filepath"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/iotbzh/xds-server/lib/common"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
)

// SDKs List of installed SDK
type SDKs struct {
	Sdks []SDK

	mutex sync.Mutex
}

// Init creates a new instance of Syncthing
func Init(cfg *xdsconfig.Config, log *logrus.Logger) (*SDKs, error) {
	s := SDKs{}

	// Retrieve installed sdks
	sdkRD := cfg.FileConf.SdkRootDir

	if common.Exists(sdkRD) {

		// Assume that SDK install tree is <rootdir>/<profile>/<version>/<arch>
		dirs, err := filepath.Glob(path.Join(sdkRD, "*", "*", "*"))
		if err != nil {
			log.Debugf("Error while retrieving SDKs: dir=%s, error=%s", sdkRD, err.Error())
			return &s, err
		}
		s.mutex.Lock()
		defer s.mutex.Unlock()

		for _, d := range dirs {
			sdk, err := NewCrossSDK(d)
			if err != nil {
				log.Debugf("Error while processing SDK dir=%s, err=%s", d, err.Error())
			}
			s.Sdks = append(s.Sdks, *sdk)
		}
	}

	log.Debugf("SDKs: %d cross sdks found", len(s.Sdks))

	return &s, nil
}

// GetAll returns all existing SDKs
func (s *SDKs) GetAll() []SDK {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	res := s.Sdks
	return res
}

// Get returns an SDK from id
func (s *SDKs) Get(id int) SDK {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if id < 0 || id > len(s.Sdks) {
		return SDK{}
	}
	res := s.Sdks[id]
	return res
}

// GetEnvCmd returns the command used to initialized the environment for an SDK
func (s *SDKs) GetEnvCmd(id string, defaultID string) string {
	if id == "" && defaultID == "" {
		// no env cmd
		return ""
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	defaultEnv := ""
	for _, sdk := range s.Sdks {
		if sdk.ID == id {
			return sdk.GetEnvCmd()
		}
		if sdk.ID == defaultID {
			defaultEnv = sdk.GetEnvCmd()
		}
	}
	// Return default env that may be empty
	return defaultEnv
}
