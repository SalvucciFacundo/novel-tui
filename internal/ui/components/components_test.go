package components_test

import (
	"os"
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/repository"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/components"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

func TestComponents(t *testing.T) {
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
