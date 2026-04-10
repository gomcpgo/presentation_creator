package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the configuration for the Presentation Creator
type Config struct {
	RootDir string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	rootDir := os.Getenv("PRESENTATION_CREATOR_ROOT_DIR")
	if rootDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		rootDir = filepath.Join(homeDir, ".savant_presentations")
	}

	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory %s: %w", rootDir, err)
	}

	return &Config{
		RootDir: rootDir,
	}, nil
}
