package common

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strings"
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
	if s == "" {
		return s, nil
	}

	// Resolved tilde : ~/
	if len(s) > 2 && s[:2] == "~/" {
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

// PathNormalize
func PathNormalize(p string) string {
	sep := string(filepath.Separator)
	if sep != "/" {
		return p
	}
	// Replace drive like C: by C/
	res := p
	if p[1:2] == ":" {
		res = p[0:1] + sep + p[2:]
	}
	res = strings.Replace(res, "\\", "/", -1)
	return filepath.Clean(res)
}
