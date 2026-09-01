package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

func TestDefaultLLMConfig(t *testing.T) {
	cfg := domain.DefaultLLMConfig()

	if cfg.Provider != "ollama" {
		t.Errorf("expected provider ollama, got %s", cfg.Provider)
	}
	if cfg.BaseURL != "http://localhost:11434" {
		t.Errorf("expected default baseURL, got %s", cfg.BaseURL)
	}
	if cfg.Model != "qwen2.5:7b" {
		t.Errorf("expected default model qwen2.5:7b, got %s", cfg.Model)
	}
	if cfg.Temperature != 0.7 {
		t.Errorf("expected temperature 0.7, got %f", cfg.Temperature)
	}
	if len(cfg.GenrePrompts) < 4 {
		t.Errorf("expected at least 4 genre prompts, got %d", len(cfg.GenrePrompts))
	}
}

func TestDefaultAppConfig(t *testing.T) {
	appCfg := domain.DefaultAppConfig()

	if appCfg == nil {
		t.Fatal("expected non-nil default app config")
	}
	if !strings.Contains(appCfg.RootDir, "Novelas") {
		t.Errorf("expected rootDir containing 'Novelas', got %s", appCfg.RootDir)
	}
	if appCfg.RecentNovels == nil {
		t.Errorf("expected non-nil recent novels slice")
	}
	if appCfg.LLM.Provider != "ollama" {
		t.Errorf("expected default LLM provider in AppConfig, got %s", appCfg.LLM.Provider)
	}
}

func TestAppConfigJSONSerialization(t *testing.T) {
	appCfg := domain.DefaultAppConfig()
	appCfg.RecentNovels = []string{"/path/to/novel1", "/path/to/novel2"}

	data, err := json.Marshal(appCfg)
	if err != nil {
		t.Fatalf("failed to marshal AppConfig: %v", err)
	}

	var parsed domain.AppConfig
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal AppConfig: %v", err)
	}

	if parsed.RootDir != appCfg.RootDir {
		t.Errorf("rootDir mismatch: got %s, want %s", parsed.RootDir, appCfg.RootDir)
	}
	if len(parsed.RecentNovels) != 2 {
		t.Errorf("expected 2 recent novels, got %d", len(parsed.RecentNovels))
	}
	if parsed.LLM.Model != "qwen2.5:7b" {
		t.Errorf("expected model qwen2.5:7b, got %s", parsed.LLM.Model)
	}
}
