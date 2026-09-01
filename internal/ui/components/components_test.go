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

func TestLauncher_MouseInteractions(t *testing.T) {
	styles := theme.DefaultStyles
	launcher := components.NewLauncherModel(styles)
	launcher.SetSize(100, 30)

	novels := []domain.NovelMetadata{
		{
			Title:        "Novela 1",
			AbsolutePath: "/tmp/novela-1",
			ChapterCount: 5,
			LastModified: time.Now(),
		},
		{
			Title:        "Novela 2",
			AbsolutePath: "/tmp/novela-2",
			ChapterCount: 2,
			LastModified: time.Now(),
		},
	}
	launcher.SetNovels(novels)

	// 1. Mouse wheel down -> increment SelectedIndex
	if launcher.SelectedIndex != 0 {
		t.Fatalf("expected initial index 0, got %d", launcher.SelectedIndex)
	}
	launcher, _ = launcher.Update(tea.MouseMsg{Type: tea.MouseWheelDown})
	if launcher.SelectedIndex != 1 {
		t.Errorf("expected SelectedIndex 1 after wheel down, got %d", launcher.SelectedIndex)
	}

	// 2. Mouse wheel up -> decrement SelectedIndex
	launcher, _ = launcher.Update(tea.MouseMsg{Type: tea.MouseWheelUp})
	if launcher.SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex 0 after wheel up, got %d", launcher.SelectedIndex)
	}

	// Calculate layout bounds to hit menu buttons
	// In size (100, 30):
	// totalW = 85, totalH = 23 (approx), startX = 7, startY = 3
	// headerHeight = 9 -> columnsY = 12
	// Button [c] is at Y = 12 + 4 = 16, X in [7, 45]
	menuX := 15

	// Click [n] (buttonIdx 1, Y = 17)
	launcher, cmd := launcher.Update(tea.MouseMsg{
		X:    menuX,
		Y:    17,
		Type: tea.MouseLeft,
	})
	if cmd == nil {
		t.Fatalf("expected cmd when clicking [n] button")
	}
	msg := cmd()
	showMsg, ok := msg.(messages.ShowModalMsg)
	if !ok || showMsg.Purpose != messages.ModalPurposeNewNovel {
		t.Errorf("expected ModalPurposeNewNovel on [n] click, got: %+v", msg)
	}

	// Click [l] (buttonIdx 3, Y = 19)
	launcher, cmd = launcher.Update(tea.MouseMsg{
		X:    menuX,
		Y:    19,
		Type: tea.MouseLeft,
	})
	if cmd == nil {
		t.Fatalf("expected cmd when clicking [l] button")
	}
	msg = cmd()
	viewMsg, ok := msg.(messages.ChangeViewMsg)
	if !ok || viewMsg.View != messages.ViewStateLLMConfig {
		t.Errorf("expected ViewStateLLMConfig on [l] click, got: %+v", msg)
	}

	// Click [d] (buttonIdx 4, Y = 20)
	launcher, cmd = launcher.Update(tea.MouseMsg{
		X:    menuX,
		Y:    20,
		Type: tea.MouseLeft,
	})
	if cmd == nil {
		t.Fatalf("expected cmd when clicking [d] button")
	}
	msg = cmd()
	showMsg, ok = msg.(messages.ShowModalMsg)
	if !ok || showMsg.Purpose != messages.ModalPurposeSetRootDir {
		t.Errorf("expected ModalPurposeSetRootDir on [d] click, got: %+v", msg)
	}

	// Click [o] (buttonIdx 2, Y = 18)
	launcher, cmd = launcher.Update(tea.MouseMsg{
		X:    menuX,
		Y:    18,
		Type: tea.MouseLeft,
	})
	if cmd == nil {
		t.Fatalf("expected cmd when clicking [o] button")
	}
	msg = cmd()
	showMsg, ok = msg.(messages.ShowModalMsg)
	if !ok || showMsg.Purpose != messages.ModalPurposeOpenFolder {
		t.Errorf("expected ModalPurposeOpenFolder on [o] click, got: %+v", msg)
	}

	// Click [c] (buttonIdx 0, Y = 16) -> continues most recent novel
	launcher, cmd = launcher.Update(tea.MouseMsg{
		X:    menuX,
		Y:    16,
		Type: tea.MouseLeft,
	})
	if cmd == nil {
		t.Fatalf("expected cmd when clicking [c] button")
	}
	msg = cmd()
	openMsg, ok := msg.(messages.OpenNovelMsg)
	if !ok || openMsg.Path != "/tmp/novela-1" {
		t.Errorf("expected OpenNovelMsg for /tmp/novela-1, got: %+v", msg)
	}

	// Click [q] (buttonIdx 5, Y = 21) -> quits
	_, qCmd := launcher.Update(tea.MouseMsg{
		X:    menuX,
		Y:    21,
		Type: tea.MouseLeft,
	})
	if qCmd == nil {
		t.Errorf("expected quit cmd on [q] click")
	}

	// Hit-test clicking recent novels:
	// novelsBox starts around X = 7 + 39 = 46
	novelsX := 55
	// Novel 1 is at Y = 12 + 4 + 1*2 = 18
	launcher.SelectedIndex = 0
	launcher, _ = launcher.Update(tea.MouseMsg{
		X:    novelsX,
		Y:    18,
		Type: tea.MouseLeft,
	})
	if launcher.SelectedIndex != 1 {
		t.Errorf("expected SelectedIndex 1 after clicking Novel 2, got %d", launcher.SelectedIndex)
	}

	// Click again on Novel 2 -> should open it
	launcher, cmd = launcher.Update(tea.MouseMsg{
		X:    novelsX,
		Y:    18,
		Type: tea.MouseLeft,
	})
	if cmd == nil {
		t.Fatalf("expected OpenNovelMsg when clicking selected novel")
	}
	msg = cmd()
	openMsg, ok = msg.(messages.OpenNovelMsg)
	if !ok || openMsg.Path != "/tmp/novela-2" {
		t.Errorf("expected OpenNovelMsg for /tmp/novela-2, got: %+v", msg)
	}
}

func TestSidebar_MouseInteractions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-sidebar-mouse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	chapRepo, _ := repository.NewFileChapterRepository(tempDir)
	c1, _ := chapRepo.Create("Chapter 1")
	c2, _ := chapRepo.Create("Chapter 2")
	charRepo := repository.NewFileCharacterRepository(tempDir)
	_ = charRepo.SaveAll([]domain.Character{{ID: "char-1", Name: "Alice", Role: "Protagonist"}})

	styles := theme.DefaultStyles
	sidebar := components.NewSidebarModel(chapRepo, charRepo, styles)
	sidebar.SetSize(30, 20)
	sidebar.Chapters = []domain.Chapter{c1, c2}
	sidebar.Characters = []domain.Character{{ID: "char-1", Name: "Alice", Role: "Protagonist"}}

	// 1. Mouse wheel in sidebar
	sidebar.SelectedChapter = 0
	sidebar, _ = sidebar.Update(tea.MouseMsg{Type: tea.MouseWheelDown})
	if sidebar.SelectedChapter != 1 {
		t.Errorf("expected SelectedChapter 1 after wheel down, got %d", sidebar.SelectedChapter)
	}
	sidebar, _ = sidebar.Update(tea.MouseMsg{Type: tea.MouseWheelUp})
	if sidebar.SelectedChapter != 0 {
		t.Errorf("expected SelectedChapter 0 after wheel up, got %d", sidebar.SelectedChapter)
	}

	// 2. Click Tab 2 (Lore) at Y = 1, X = 18
	sidebar, _ = sidebar.Update(tea.MouseMsg{
		X:    18,
		Y:    1,
		Type: tea.MouseLeft,
	})
	if sidebar.ActiveTab != components.TabLore {
		t.Errorf("expected ActiveTab to be TabLore, got %v", sidebar.ActiveTab)
	}

	// 3. Click Tab 1 (Chapters) at Y = 1, X = 5
	sidebar, _ = sidebar.Update(tea.MouseMsg{
		X:    5,
		Y:    1,
		Type: tea.MouseLeft,
	})
	if sidebar.ActiveTab != components.TabChapters {
		t.Errorf("expected ActiveTab to be TabChapters, got %v", sidebar.ActiveTab)
	}

	// 4. Click chapter row 1 (Chapter 2) at Y = 5 (since row 0 is at Y=3,4, row 1 is at Y=5,6)
	sidebar, cmd := sidebar.Update(tea.MouseMsg{
		X:    5,
		Y:    5,
		Type: tea.MouseLeft,
	})
	if sidebar.SelectedChapter != 1 {
		t.Errorf("expected SelectedChapter 1, got %d", sidebar.SelectedChapter)
	}
	if cmd == nil {
		t.Fatalf("expected ChapterSelectedMsg when clicking chapter row")
	}
	msg := cmd()
	chapMsg, ok := msg.(messages.ChapterSelectedMsg)
	if !ok || chapMsg.Chapter.Title != "Chapter 2" {
		t.Errorf("unexpected chapter selected msg: %+v", msg)
	}
}

func TestEditor_MouseInteractions(t *testing.T) {
	styles := theme.DefaultStyles
	editor := components.NewEditorModel(styles)
	editor.SetSize(60, 20)
	editor, _ = editor.Update(messages.ChapterSelectedMsg{
		Chapter: domain.Chapter{
			ID:      "chap-1",
			Title:   "Chapter 1",
			Content: "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8",
		},
	})

	// 1. Left click focuses editor
	editor.Focused = false
	editor, cmd := editor.Update(tea.MouseMsg{
		X:    10,
		Y:    5,
		Type: tea.MouseLeft,
	})
	if !editor.Focused {
		t.Errorf("expected editor to be focused after left click")
	}
	if cmd == nil {
		t.Fatalf("expected FocusMsg cmd on editor click")
	}

	// 2. Wheel scroll up and down
	editor, _ = editor.Update(tea.MouseMsg{Type: tea.MouseWheelDown})
	editor, _ = editor.Update(tea.MouseMsg{Type: tea.MouseWheelUp})
}

func TestModal_MouseInteractions(t *testing.T) {
	styles := theme.DefaultStyles
	modal := components.NewModalModel(styles)
	modal.SetSize(80, 24)
	modal.Show(messages.ModalPurposeNewNovel, "Nueva Novela", "Título:", "Valor inicial")

	// Left click inside modal
	modal, _ = modal.Update(tea.MouseMsg{
		X:    40,
		Y:    12,
		Type: tea.MouseLeft,
	})
	if !modal.Active {
		t.Errorf("modal should remain active")
	}
}
