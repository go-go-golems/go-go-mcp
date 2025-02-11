package config

import (
	"os"
	"path/filepath"
)

// GetDefaultProfilesPath returns the default path for the profiles configuration file
func GetDefaultProfilesPath() (string, error) {
	xdgConfigPath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(xdgConfigPath, "go-go-mcp", "profiles.yaml"), nil
}

// GetProfilesPath returns the profiles path from either the provided path or the default location
func GetProfilesPath(configFile string) (string, error) {
	if configFile != "" {
		return configFile, nil
	}
	return GetDefaultProfilesPath()
}
