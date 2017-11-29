package xdsconfig

import (
	"errors"
	"net"

	"github.com/iotbzh/xds-server/lib/xsapiv1"
)

// NewBuilderConfig creates a new BuilderConfig instance
func NewBuilderConfig(stID string) (xsapiv1.BuilderConfig, error) {
	// Do we really need it ? may be not accessible from client side
	ip, err := getLocalIP()
	if err != nil {
		return xsapiv1.BuilderConfig{}, err
	}

	b := xsapiv1.BuilderConfig{
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
