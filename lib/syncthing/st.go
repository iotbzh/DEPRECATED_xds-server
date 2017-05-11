package st

import (
	"encoding/json"

	"strings"

	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/iotbzh/xds-server/lib/common"
	"github.com/syncthing/syncthing/lib/config"
)

// SyncThing .
type SyncThing struct {
	BaseURL string
	client  *common.HTTPClient
	log     *logrus.Logger
}

// NewSyncThing creates a new instance of Syncthing
func NewSyncThing(url string, apikey string, log *logrus.Logger) *SyncThing {
	cl, err := common.HTTPNewClient(url,
		common.HTTPClientConfig{
			URLPrefix:           "/rest",
			HeaderClientKeyName: "X-Syncthing-ID",
		})
	if err != nil {
		msg := ": " + err.Error()
		if strings.Contains(err.Error(), "connection refused") {
			msg = fmt.Sprintf("(url: %s)", url)
		}
		log.Debugf("ERROR: cannot connect to Syncthing %s", msg)
		return nil
	}

	s := SyncThing{
		BaseURL: url,
		client:  cl,
		log:     log,
	}

	return &s
}

// IDGet returns the Syncthing ID of Syncthing instance running locally
func (s *SyncThing) IDGet() (string, error) {
	var data []byte
	if err := s.client.HTTPGet("system/status", &data); err != nil {
		return "", err
	}
	status := make(map[string]interface{})
	json.Unmarshal(data, &status)
	return status["myID"].(string), nil
}

// ConfigGet returns the current Syncthing configuration
func (s *SyncThing) ConfigGet() (config.Configuration, error) {
	var data []byte
	config := config.Configuration{}
	if err := s.client.HTTPGet("system/config", &data); err != nil {
		return config, err
	}
	err := json.Unmarshal(data, &config)
	return config, err
}

// ConfigSet set Syncthing configuration
func (s *SyncThing) ConfigSet(cfg config.Configuration) error {
	body, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return s.client.HTTPPost("system/config", string(body))
}
