package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// NavbarPillID identifies clickable pills in the top navbar.
type NavbarPillID int

const (
	PillHome NavbarPillID = iota
	PillTabChapters
	PillTabCharacters
	PillTabNotes
	PillTabBrain
	PillToggleChat
)

type pillHitZone struct {
	id     NavbarPillID
	startX int
	endX   int
}

// NavbarModel coordinates top navigation bar rendering, breadcrumbs, and action pills.
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

func (m NavbarModel) getHitZones() []pillHitZone {
	w := m.Width
	if w <= 0 {
		w = 100
	}

	var zones []pillHitZone

	// 1. Home Pill
	homePill := m.styles.NavbarHomePill.Render("[← Inicio (Ctrl+H)]")
	homePillW := lipgloss.Width(homePill)
	zones = append(zones, pillHitZone{
		id:     PillHome,
		startX: 0,
		endX:   homePillW,
	})

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

	// 3. Right Action Pills
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

	rightStartX := w - rightWidth
	if rightStartX < leftWidth {
		rightStartX = leftWidth
	}

	currentX := rightStartX
	wChap := lipgloss.Width(pillChap)
	zones = append(zones, pillHitZone{
		id:     PillTabChapters,
		startX: currentX,
		endX:   currentX + wChap,
	})
	currentX += wChap + 1

	wChar := lipgloss.Width(pillChar)
	zones = append(zones, pillHitZone{
		id:     PillTabCharacters,
		startX: currentX,
		endX:   currentX + wChar,
	})
	currentX += wChar + 1

	wNotes := lipgloss.Width(pillNotes)
	zones = append(zones, pillHitZone{
		id:     PillTabNotes,
		startX: currentX,
		endX:   currentX + wNotes,
	})
	currentX += wNotes + 1

	wBrain := lipgloss.Width(pillBrain)
	zones = append(zones, pillHitZone{
		id:     PillTabBrain,
		startX: currentX,
		endX:   currentX + wBrain,
	})
	currentX += wBrain + 1

	wChat := lipgloss.Width(pillChat)
	zones = append(zones, pillHitZone{
		id:     PillToggleChat,
		startX: currentX,
		endX:   currentX + wChat,
	})

	return zones
}

// Update processes incoming messages such as chapter selection and mouse clicks.
func (m NavbarModel) Update(msg tea.Msg) (NavbarModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.ChapterSelectedMsg:
		m.ChapterTitle = msg.Chapter.Title
		return m, nil

	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			zones := m.getHitZones()
			for _, zone := range zones {
				if msg.X >= zone.startX && msg.X < zone.endX {
					switch zone.id {
					case PillHome:
						return m, func() tea.Msg {
							return messages.ChangeViewMsg{View: messages.ViewStateLauncher}
						}
					case PillTabChapters:
						return m, func() tea.Msg {
							return messages.SelectSidebarTabMsg{Tab: 0}
						}
					case PillTabCharacters:
						return m, func() tea.Msg {
							return messages.SelectSidebarTabMsg{Tab: 1}
						}
					case PillTabNotes:
						return m, func() tea.Msg {
							return messages.SelectSidebarTabMsg{Tab: 2}
						}
					case PillTabBrain:
						return m, func() tea.Msg {
							return messages.SelectSidebarTabMsg{Tab: 3}
						}
					case PillToggleChat:
						return m, func() tea.Msg {
							return messages.ToggleChatDrawerMsg{}
						}
					}
				}
			}
		}
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

	// 3. Right Action Pills
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
