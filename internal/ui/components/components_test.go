package components_test

import (
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/repository"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/components"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

func TestSidebarAndEditorAndStatusBar(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-components-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	chapRepo, _ := repository.NewFileChapterRepository(tempDir)
	charRepo := repository.NewFileCharacterRepository(tempDir)

	styles := theme.DefaultStyles

	// 1. SidebarModel test
	sidebar := components.NewSidebarModel(chapRepo, charRepo, styles)
	sidebar.SetSize(30, 20)

	// Create sample chapter
	c1, _ := chapRepo.Create("Chapter 1: The Call")
	sidebar, _ = sidebar.Update(messages.ChapterCreatedMsg{Chapter: c1})

	sidebarView := sidebar.View()
	if sidebarView == "" {
		t.Errorf("sidebar view should not be empty")
	}

	// 2. EditorModel test
	editor := components.NewEditorModel(styles)
	editor.SetSize(60, 20)

	// Select chapter in editor
	editor, _ = editor.Update(messages.ChapterSelectedMsg{
		Chapter: domain.Chapter{
			ID:      "chap-1",
			Title:   "Chapter 1",
			Content: "Once upon a time in a distant land...",
		},
	})
	if editor.Value() != "Once upon a time in a distant land..." {
		t.Errorf("expected editor content loaded, got %s", editor.Value())
	}

	// 3. StatusBarModel test
	statusBar := components.NewStatusBarModel(styles)
	statusBar.SetWidth(80)
	statusBar, _ = statusBar.Update(messages.ChapterSelectedMsg{
		Chapter: domain.Chapter{
			ID:      "chap-1",
			Title:   "Chapter 1",
			Content: "Once upon a time...",
		},
	})
	statusView := statusBar.View()
	if statusView == "" {
		t.Errorf("statusBar view should not be empty")
	}
}

func TestModalComponent(t *testing.T) {
	styles := theme.DefaultStyles
	modal := components.NewModalModel(styles)
	modal.SetSize(80, 24)

	// Initially inactive
	if modal.Active {
		t.Errorf("modal should not be active initially")
	}
	if modal.View() != "" {
		t.Errorf("modal view should be empty when inactive")
	}

	// Show modal
	modal.Show(messages.ModalPurposeNewNovel, "Nueva Novela", "Título:", "")
	if !modal.Active {
		t.Errorf("modal should be active after Show()")
	}

	view := modal.View()
	if !strings.Contains(view, "Nueva Novela") {
		t.Errorf("expected title in modal view, got: %s", view)
	}

	// Submit empty input -> should trigger error
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if modal.ErrorMsg == "" {
		t.Errorf("expected validation error on empty submit")
	}

	// Type input and submit
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Mi Novela")})
	modal, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if modal.Active {
		t.Errorf("modal should be closed after valid submission")
	}
	if cmd == nil {
		t.Fatalf("expected command emitted on valid modal submit")
	}
	msg := cmd()
	createMsg, ok := msg.(messages.CreateNovelMsg)
	if !ok || createMsg.Title != "Mi Novela" {
		t.Errorf("unexpected message from modal submit: %+v", msg)
	}
}

func TestLauncherComponent(t *testing.T) {
	styles := theme.DefaultStyles
	launcher := components.NewLauncherModel(styles)
	launcher.SetSize(100, 30)

	novels := []domain.NovelMetadata{
		{
			Title:        "Novela de Prueba",
			AbsolutePath: "/tmp/novela-prueba",
			ChapterCount: 3,
			LastModified: time.Now(),
		},
	}
	launcher.SetNovels(novels)

	view := launcher.View()
	if !strings.Contains(view, "Novela de Prueba") {
		t.Errorf("expected novel title in launcher view: %s", view)
	}

	// Press 'n' -> should emit ShowModalMsg
	_, cmd := launcher.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if cmd == nil {
		t.Fatalf("expected command on pressing 'n'")
	}
	msg := cmd()
	showMsg, ok := msg.(messages.ShowModalMsg)
	if !ok || showMsg.Purpose != messages.ModalPurposeNewNovel {
		t.Errorf("unexpected message on 'n': %+v", msg)
	}

	// Press 'l' -> should emit ChangeViewMsg for LLM config
	_, cmd = launcher.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if cmd == nil {
		t.Fatalf("expected command on pressing 'l'")
	}
	msg = cmd()
	viewMsg, ok := msg.(messages.ChangeViewMsg)
	if !ok || viewMsg.View != messages.ViewStateLLMConfig {
		t.Errorf("unexpected view change message: %+v", msg)
	}
}

func TestLLMConfigComponent(t *testing.T) {
	styles := theme.DefaultStyles
	llmCfg := components.NewLLMConfigModel(styles)
	llmCfg.SetSize(100, 30)

	view := llmCfg.View()
	if !strings.Contains(view, "Configuración de Inteligencia Artificial") {
		t.Errorf("expected LLM config header in view: %s", view)
	}

	// Press 'Esc' -> should emit ChangeViewMsg for Launcher
	_, cmd := llmCfg.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatalf("expected command on pressing 'Esc'")
	}
	msg := cmd()
	viewMsg, ok := msg.(messages.ChangeViewMsg)
	if !ok || viewMsg.View != messages.ViewStateLauncher {
		t.Errorf("unexpected view change message on Esc: %+v", msg)
	}

	// Press 'Ctrl+S' -> should emit SaveLLMConfigMsg
	_, cmd = llmCfg.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatalf("expected command on pressing 'Ctrl+S'")
	}
	msg = cmd()
	saveMsg, ok := msg.(messages.SaveLLMConfigMsg)
	if !ok || saveMsg.Config.Provider != "ollama" {
		t.Errorf("unexpected save message: %+v", msg)
	}
}
