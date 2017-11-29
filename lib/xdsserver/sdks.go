package xdsserver

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"

	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
)

// SDKs List of installed SDK
type SDKs struct {
	*Context
	Sdks map[string]*CrossSDK

	mutex sync.Mutex
}

// NewSDKs creates a new instance of SDKs
func NewSDKs(ctx *Context) (*SDKs, error) {
	s := SDKs{
		Context: ctx,
		Sdks:    make(map[string]*CrossSDK),
	}

	// Retrieve installed sdks
	sdkRD := ctx.Config.FileConf.SdkRootDir

	if common.Exists(sdkRD) {

		// Assume that SDK install tree is <rootdir>/<profile>/<version>/<arch>
		dirs, err := filepath.Glob(path.Join(sdkRD, "*", "*", "*"))
		if err != nil {
			ctx.Log.Debugf("Error while retrieving SDKs: dir=%s, error=%s", sdkRD, err.Error())
			return &s, err
		}
		s.mutex.Lock()
		defer s.mutex.Unlock()

		for _, d := range dirs {
			if !common.IsDir(d) {
				continue
			}
			cSdk, err := NewCrossSDK(d)
			if err != nil {
				ctx.Log.Debugf("Error while processing SDK dir=%s, err=%s", d, err.Error())
				continue
			}
			s.Sdks[cSdk.sdk.ID] = cSdk
		}
	}

	ctx.Log.Debugf("SDKs: %d cross sdks found", len(s.Sdks))

	return &s, nil
}

// ResolveID Complete an SDK ID (helper for user that can use partial ID value)
func (s *SDKs) ResolveID(id string) (string, error) {
	if id == "" {
		return "", nil
	}

	match := []string{}
	for iid := range s.Sdks {
		if strings.HasPrefix(iid, id) {
			match = append(match, iid)
		}
	}

	if len(match) == 1 {
		return match[0], nil
	} else if len(match) == 0 {
		return id, fmt.Errorf("Unknown sdk id")
	}
	return id, fmt.Errorf("Multiple sdk IDs found with provided prefix: " + id)
}

// Get returns an SDK from id
func (s *SDKs) Get(id string) *xsapiv1.SDK {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sc, exist := s.Sdks[id]
	if !exist {
		return nil
	}
	return (*sc).Get()
}

// GetAll returns all existing SDKs
func (s *SDKs) GetAll() []xsapiv1.SDK {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	res := []xsapiv1.SDK{}
	for _, v := range s.Sdks {
		res = append(res, *(*v).Get())
	}
	return res
}

// GetEnvCmd returns the command used to initialized the environment for an SDK
func (s *SDKs) GetEnvCmd(id string, defaultID string) []string {
	if id == "" && defaultID == "" {
		// no env cmd
		return []string{}
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if iid, err := s.ResolveID(id); err == nil {
		if sdk, exist := s.Sdks[iid]; exist {
			return sdk.GetEnvCmd()
		}
	}

	if sdk, exist := s.Sdks[defaultID]; defaultID != "" && exist {
		return sdk.GetEnvCmd()
	}

	// Return default env that may be empty
	return []string{}
}
