package xdsconfig

import (
	"errors"
	"net"
)

// BuilderConfig represents the builder container configuration
type BuilderConfig struct {
	IP          string `json:"ip"`
	Port        string `json:"port"`
	SyncThingID string `json:"syncThingID"`
}

// NewBuilderConfig creates a new BuilderConfig instance
func NewBuilderConfig(stID string) (BuilderConfig, error) {
	// Do we really need it ? may be not accessible from client side
	ip, err := getLocalIP()
	if err != nil {
		return BuilderConfig{}, err
	}

	b := BuilderConfig{
		IP:          ip, // TODO currently not used
		Port:        "", // TODO currently not used
		SyncThingID: stID,
	}
	return b, nil
}

/*** Private ***/

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("Cannot determined local IP")
}
