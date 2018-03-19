package settings

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

const (
	settingsDirApplicationName = "tune"
	settingsFileName           = "settings.toml"
)

func settingsPath() string {
	return filepath.Join(localSettingsFolder, "tune", "settings.toml")
}

// Settings holds the settings variables that are persisted to disk
type Settings struct {
	Account struct {
		APIKey string
	}

	Settings struct {
		StreamlistKey string
	}

	Player struct {
		Volume            int
		LastPlayedChannel string
	}
}

// newDefault creates a new settings with safe defaults
func newDefault() *Settings {
	c := &Settings{}
	c.Player.Volume = 50
	return c
}

// Load loads settings variables from disk
func Load() (*Settings, error) {
	settingsFile, err := os.Open(settingsPath())
	if err != nil {
		if os.IsNotExist(err) {
			// create new default settings and make sure we can save to disk
			c := newDefault()
			err = c.Save()
			if err != nil {
				return nil, errors.Wrap(err, "failed to create new settings file")
			}
			return c, err
		}
	}
	defer settingsFile.Close()

	// read all bytes from file
	settingsBytes, err := ioutil.ReadAll(settingsFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read from settings file")
	}
	c := &Settings{}
	err = toml.Unmarshal(settingsBytes, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Save saves the settings to disk in yaml format
func (c *Settings) Save() error {
	// TODO: This is not an atomic file write, instead file should be created with a random suffix and
	// moved over existing settings file when completed.

	// create parent path
	err := os.MkdirAll(filepath.Dir(settingsPath()), 0775)
	if err != nil {
		return errors.Wrap(err, "failed to create directory for settings")
	}

	// create settings file
	settingsFile, err := os.Create(settingsPath())
	if err != nil {
		return err
	}
	defer settingsFile.Close()

	// encode settings to file
	err = toml.NewEncoder(settingsFile).Encode(c)
	if err != nil {
		return err
	}

	return nil
}
