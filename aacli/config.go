package main

import (
	"errors"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var (
	// ErrNoConfigFile is returned by LoadConfig() when no config file was found.
	ErrNoConfigFile = errors.New("no config file")
)
var (
	configDirPath  string
	configFilePath string
)

func init() {
	// get user directory
	u, err := user.Current()
	if err != nil || u == nil {
		configDirPath = "."
		configFilePath = filepath.Join(configDirPath, "aaconfig")
		return
	}
	// build config dir and file paths
	configDirPath = filepath.Join(u.HomeDir, ".config", "audioaddict")
	configFilePath = filepath.Join(configDirPath, "aaconfig")
}

// Config holds the config variables that are persisted to disk
type Config struct {
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

// NewConfig creates a new config with default settings
func NewConfig() *Config {
	c := &Config{}
	c.Player.Volume = 50
	return c
}

// LoadConfig loads config variables from disk
func LoadConfig() (*Config, error) {
	configBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoConfigFile
		}
		return nil, err
	}
	c := &Config{}
	err = yaml.Unmarshal(configBytes, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Save saves the config to disk in yaml format
func (c *Config) Save() error {
	configBytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	// create config dir
	err = os.MkdirAll(configDirPath, 0755)
	if err != nil {
		return err
	}
	// create new config file
	err = ioutil.WriteFile(configFilePath, configBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}
