package crosssdk

import (
	"path"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/iotbzh/xds-server/lib/common"
	"github.com/iotbzh/xds-server/lib/xdsconfig"
)

// SDKs List of installed SDK
type SDKs []*SDK

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
		for _, d := range dirs {
			sdk, err := NewCrossSDK(d)
			if err != nil {
				log.Debugf("Error while processing SDK dir=%s, err=%s", d, err.Error())
			}
			s = append(s, sdk)
		}
	}
	return &s, nil
}
