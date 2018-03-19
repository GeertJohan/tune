// +build !windows,!darwin

package settings

import (
	"os"
	"path/filepath"
)

var localSettingsFolder string

func init() {
	// https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		localSettingsFolder = os.Getenv("XDG_CONFIG_HOME")
	} else {
		localSettingsFolder = filepath.Join(os.Getenv("HOME"), ".config")
	}
}
