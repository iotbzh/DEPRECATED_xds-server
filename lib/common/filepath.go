package common

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
)

// Exists returns whether the given file or directory exists or not
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// ResolveEnvVar Resolved environment variable regarding the syntax ${MYVAR}
// or $MYVAR following by a slash or a backslash
func ResolveEnvVar(s string) (string, error) {

	// Resolved tilde : ~/
	if s[:2] == "~/" {
		if usr, err := user.Current(); err == nil {
			s = filepath.Join(usr.HomeDir, s[2:])
		}
	}

	// Resolved ${MYVAR}
	re := regexp.MustCompile("\\${([^}]+)}")
	vars := re.FindAllStringSubmatch(s, -1)
	res := s
	for _, v := range vars {
		val := os.Getenv(v[1])
		if val == "" {
			return res, fmt.Errorf("ERROR: %s env variable not defined", v[1])
		}

		rer := regexp.MustCompile("\\${" + v[1] + "}")
		res = rer.ReplaceAllString(res, val)
	}

	// Resolved $MYVAR following by a slash (or a backslash for Windows)
	// TODO
	//re := regexp.MustCompile("\\$([^\\/])+/")

	return path.Clean(res), nil
}
