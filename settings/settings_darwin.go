package settings

import (
	"os"
	"path/filepath"
)

var localSettingsFolder = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
