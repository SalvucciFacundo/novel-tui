package model_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SalvucciFacundo/novel-tui/internal/repository"
	"github.com/SalvucciFacundo/novel-tui/internal/service"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/model"
)

func TestRootModel_Init_NoPanicWhenNoInitialDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-init-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateLauncher, "")

	// Init should not panic
	cmd := root.Init()
	if cmd != nil {
		// Bubble tea batch command execution simulation
		_ = cmd()
	}
}

func TestRootModel_InitializationAndResizing(t *testing.T) {
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

	// 3. Normal window layout in editor view
	normalSizeMsg := tea.WindowSizeMsg{Width: 100, Height: 30}
	updatedModel, _ = updatedModel.Update(normalSizeMsg)
	normalView := updatedModel.View()
	if strings.Contains(normalView, "Terminal size too small") {
		t.Errorf("expected normal view, but got warning: %s", normalView)
	}
}

func TestRootModel_LauncherModeAndTransitions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-launcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateLauncher, "")

	// Initialize window size
	m, _ := root.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	view := m.View()

	if !strings.Contains(view, "Acciones Rápidas") {
		t.Errorf("expected launcher in view: %s", view)
	}

	// Transition to LLM Config view
	m, _ = m.Update(messages.ChangeViewMsg{View: messages.ViewStateLLMConfig})
	llmView := m.View()
	if !strings.Contains(llmView, "Configuración de Inteligencia Artificial") {
		t.Errorf("expected LLM config view, got: %s", llmView)
	}

	// Transition back to Launcher
	m, _ = m.Update(messages.ChangeViewMsg{View: messages.ViewStateLauncher})
	launcherView := m.View()
	if !strings.Contains(launcherView, "Acciones Rápidas") {
		t.Errorf("expected Launcher view, got: %s", launcherView)
	}

	// Open Modal for New Novel
	m, _ = m.Update(messages.ShowModalMsg{
		Purpose: messages.ModalPurposeNewNovel,
		Title:   "Nueva Novela",
		Prompt:  "Título de la novela:",
	})
	modalView := m.View()
	if !strings.Contains(modalView, "Nueva Novela") {
		t.Errorf("expected modal view overlay, got: %s", modalView)
	}

	// Close Modal
	m, _ = m.Update(messages.HideModalMsg{})
	afterClose := m.View()
	if strings.Contains(afterClose, "Título de la novela:") {
		t.Errorf("expected modal to be hidden")
	}
}

func TestRootModel_CreateNovelAndTransitionToEditor(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-create-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	cfg, _ := configRepo.Load()
	cfg.RootDir = tempDir
	_ = configRepo.Save(cfg)

	workspaceMgr := service.NewWorkspaceManager()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateLauncher, "")
	m, _ := root.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Dispatch CreateNovelMsg
	m, cmd := m.Update(messages.CreateNovelMsg{Title: "Las Crónicas de Noria"})
	if cmd == nil {
		t.Fatalf("expected command after CreateNovelMsg")
	}
	msg := cmd()
	openMsg, ok := msg.(messages.OpenNovelMsg)
	if !ok {
		t.Fatalf("expected OpenNovelMsg, got: %+v", msg)
	}

	// Update with OpenNovelMsg
	m, _ = m.Update(openMsg)

	// Verify view is now in Editor mode
	editorView := m.View()
	if strings.Contains(editorView, "Acciones Rápidas") {
		t.Errorf("expected Editor view, but still on Launcher")
	}
}
