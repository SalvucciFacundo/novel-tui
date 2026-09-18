package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// NavbarModel coordinates top navigation bar rendering, breadcrumbs, and action pills.
//
// The pills are a visual shortcut legend only: they are not clickable.
// Hit-testing them proved unreliable across terminal widths (narrow screens
// shift the right-aligned pills away from their computed zones), so all
// navigation stays keyboard-driven (Ctrl+H, Alt+1..4, Ctrl+A).
type NavbarModel struct {
	NovelTitle   string
	ChapterTitle string

	Width  int
	styles theme.Styles
}

// NewNavbarModel creates a new NavbarModel.
func NewNavbarModel(styles theme.Styles) NavbarModel {
	return NavbarModel{
		NovelTitle:   "",
		ChapterTitle: "",
		styles:       styles,
	}
}

// Init initializes the navbar component.
func (m NavbarModel) Init() tea.Cmd {
	return nil
}

// SetWidth sets the allocated terminal width for the navbar.
func (m *NavbarModel) SetWidth(w int) {
	m.Width = w
}

// SetNovelTitle updates the active novel's display title.
func (m *NavbarModel) SetNovelTitle(title string) {
	m.NovelTitle = title
}

// SetChapterTitle updates the active chapter's display title.
func (m *NavbarModel) SetChapterTitle(title string) {
	m.ChapterTitle = title
}

// Update processes incoming messages such as chapter selection.
func (m NavbarModel) Update(msg tea.Msg) (NavbarModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.ChapterSelectedMsg:
		m.ChapterTitle = msg.Chapter.Title
		return m, nil
	}
	return m, nil
}

// View renders the horizontal top navigation bar.
func (m NavbarModel) View() string {
	if m.Width <= 0 {
		return ""
	}

	// 1. Home Pill
	homePill := m.styles.NavbarHomePill.Render("[← Inicio (Ctrl+H)]")

	// 2. Breadcrumbs
	novelName := m.NovelTitle
	if strings.TrimSpace(novelName) == "" {
		novelName = "Novela"
	}

	chapterName := m.ChapterTitle
	if strings.TrimSpace(chapterName) == "" {
		chapterName = "Ningún capítulo seleccionado"
	}

	breadcrumbText := " 📖 " + novelName + " › 📑 " + chapterName
	renderedBreadcrumb := m.styles.NavbarBreadcrumb.Render(breadcrumbText)

	leftSection := lipgloss.JoinHorizontal(lipgloss.Center, homePill, renderedBreadcrumb)
	leftWidth := lipgloss.Width(leftSection)

	// 3. Right Action Pills (visual legend; navigation is keyboard-driven:
	// Alt+1..4 works from anywhere, plain digits outside text inputs)
	pillChap := m.styles.NavbarActionPill.Render("[1: Capítulos]")
	pillChar := m.styles.NavbarActionPill.Render("[2: Personajes]")
	pillNotes := m.styles.NavbarActionPill.Render("[3: Notas]")
	pillBrain := m.styles.NavbarActionPill.Render("[4: Brain]")
	pillChat := m.styles.NavbarActionPill.Render("[🤖 Asistente IA (Ctrl+A)]")

	rightSection := lipgloss.JoinHorizontal(lipgloss.Center,
		pillChap, " ",
		pillChar, " ",
		pillNotes, " ",
		pillBrain, " ",
		pillChat, " ",
	)
	rightWidth := lipgloss.Width(rightSection)

	// Spacer between left and right sections
	spaceWidth := m.Width - leftWidth - rightWidth
	if spaceWidth < 1 {
		spaceWidth = 1
	}

	spacer := lipgloss.NewStyle().
		Background(m.styles.NavbarContainer.GetBackground()).
		Width(spaceWidth).
		Render("")

	barContent := lipgloss.JoinHorizontal(lipgloss.Center, leftSection, spacer, rightSection)

	return m.styles.NavbarContainer.
		Width(m.Width).
		Render(barContent)
}
