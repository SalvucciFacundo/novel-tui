package model_test

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SalvucciFacundo/novel-tui/internal/repository"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/model"
)

func TestRootModel(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-root-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	chapterRepo, err := repository.NewFileChapterRepository(tempDir)
	if err != nil {
		t.Fatalf("NewFileChapterRepository failed: %v", err)
	}
	characterRepo := repository.NewFileCharacterRepository(tempDir)

	root := model.NewRootModel(chapterRepo, characterRepo)

	// 1. Initial view before resize message
	initView := root.View()
	if !strings.Contains(initView, "Initializing Novel TUI") {
		t.Errorf("expected Initializing state, got %s", initView)
	}

	// 2. Small window warning
	smallSizeMsg := tea.WindowSizeMsg{Width: 40, Height: 10}
	updatedModel, _ := root.Update(smallSizeMsg)
	smallView := updatedModel.View()
	if !strings.Contains(smallView, "Terminal size too small") {
		t.Errorf("expected small window warning, got: %s", smallView)
	}

	// 3. Normal window layout
	normalSizeMsg := tea.WindowSizeMsg{Width: 100, Height: 30}
	updatedModel, _ = updatedModel.Update(normalSizeMsg)
	normalView := updatedModel.View()
	if strings.Contains(normalView, "Terminal size too small") {
		t.Errorf("expected normal view, but got warning: %s", normalView)
	}
}
