package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

// ConfigRepository defines the contract for loading and persisting application configuration.
type ConfigRepository interface {
	Load() (*domain.AppConfig, error)
	Save(cfg *domain.AppConfig) error
}

// FileConfigRepository implements ConfigRepository using a JSON file on disk.
type FileConfigRepository struct {
	configPath string
}

// ExpandHome expands the leading "~" in a file path to the user's home directory.
func ExpandHome(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
		return path
	}
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// DefaultConfigPath returns the canonical path ~/.config/novel-tui/config.json.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, ".config", "novel-tui", "config.json")
	}
	configDir, err := os.UserConfigDir()
	if err == nil && configDir != "" {
		return filepath.Join(configDir, "novel-tui", "config.json")
	}
	return ".config/novel-tui/config.json"
}

// NewFileConfigRepository creates a new FileConfigRepository.
// If customPath is empty, DefaultConfigPath() is used.
func NewFileConfigRepository(customPath string) *FileConfigRepository {
	if customPath == "" {
		customPath = DefaultConfigPath()
	} else {
		customPath = ExpandHome(customPath)
	}
	return &FileConfigRepository{configPath: customPath}
}

// Path returns the resolved path to the config file.
func (r *FileConfigRepository) Path() string {
	return r.configPath
}

// Load reads and unmarshals the configuration file.
// If the configuration file does not exist, it initializes default configuration and saves it.
func (r *FileConfigRepository) Load() (*domain.AppConfig, error) {
	data, err := os.ReadFile(r.configPath)
	if os.IsNotExist(err) {
		defaultCfg := domain.DefaultAppConfig()
		defaultCfg.RootDir = ExpandHome(defaultCfg.RootDir)
		if err := r.Save(defaultCfg); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		// Ensure root directory exists
		_ = os.MkdirAll(defaultCfg.RootDir, 0755)
		return defaultCfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config file at %s: %w", r.configPath, err)
	}

	var cfg domain.AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	cfg.RootDir = ExpandHome(cfg.RootDir)
	for i, path := range cfg.RecentNovels {
		cfg.RecentNovels[i] = ExpandHome(path)
	}

	return &cfg, nil
}

// Save marshals and writes the configuration file atomically.
func (r *FileConfigRepository) Save(cfg *domain.AppConfig) error {
	dir := filepath.Dir(r.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	tmpPath := filepath.Join(dir, fmt.Sprintf(".%s.tmp", filepath.Base(r.configPath)))
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp config file: %w", err)
	}

	if err := os.Rename(tmpPath, r.configPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp config file to %s: %w", r.configPath, err)
	}

	return nil
}
