package pwd

import (
	"os"
	"strings"
)

func IsDebug() bool {
	return false
}

func NormalizePath(path string) string {
	// add '/' char at the beginning if missing
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// remove last '/' char if exists
	if len(path) > 0 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-len("/")]
	}
	return path
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
