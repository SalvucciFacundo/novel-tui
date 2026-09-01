package repository_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/repository"
)

func TestFileConfigRepository_LoadNonExistent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "sub", "config.json")
	repo := repository.NewFileConfigRepository(configPath)

	// Load when file does not exist -> should initialize default config
	cfg, err := repo.Load()
	if err != nil {
		t.Fatalf("Load failed on non-existent config: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.LLM.Provider != "ollama" {
		t.Errorf("expected provider ollama, got %s", cfg.LLM.Provider)
	}

	// Verify file was written to disk
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("expected config file to be created at %s", configPath)
	}
}

func TestFileConfigRepository_SaveAndLoad(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	repo := repository.NewFileConfigRepository(configPath)

	customCfg := &domain.AppConfig{
		RootDir:      filepath.Join(tempDir, "MyNovels"),
		RecentNovels: []string{filepath.Join(tempDir, "MyNovels", "NovelA")},
		LLM: domain.LLMConfig{
			Provider:    "ollama",
			BaseURL:     "http://localhost:11434",
			Model:       "mistral:latest",
			Temperature: 0.85,
			GenrePrompts: map[string]string{
				"Fantasy": "Custom fantasy prompt",
			},
		},
	}

	if err := repo.Save(customCfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := repo.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.RootDir != customCfg.RootDir {
		t.Errorf("RootDir mismatch: got %s, want %s", loaded.RootDir, customCfg.RootDir)
	}
	if len(loaded.RecentNovels) != 1 || loaded.RecentNovels[0] != customCfg.RecentNovels[0] {
		t.Errorf("RecentNovels mismatch: %+v", loaded.RecentNovels)
	}
	if loaded.LLM.Model != "mistral:latest" {
		t.Errorf("LLM Model mismatch: got %s, want mistral:latest", loaded.LLM.Model)
	}
	if loaded.LLM.Temperature != 0.85 {
		t.Errorf("LLM Temperature mismatch: got %f, want 0.85", loaded.LLM.Temperature)
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("UserHomeDir not available")
	}

	expanded := repository.ExpandHome("~/Novelas")
	expected := filepath.Join(home, "Novelas")
	if expanded != expected {
		t.Errorf("ExpandHome failed: got %s, want %s", expanded, expected)
	}

	expandedExact := repository.ExpandHome("~")
	if expandedExact != home {
		t.Errorf("ExpandHome('~') failed: got %s, want %s", expandedExact, home)
	}

	regularPath := "/tmp/novels"
	if repository.ExpandHome(regularPath) != regularPath {
		t.Errorf("ExpandHome should not modify non-home path")
	}
}
