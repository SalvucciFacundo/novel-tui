package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// SidebarTab represents the active view mode in the sidebar.
type SidebarTab int

const (
	TabChapters SidebarTab = iota
	TabLore
)

// SidebarKeyMap defines keybindings for the sidebar.
type SidebarKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Select   key.Binding
	New      key.Binding
	PrevTab  key.Binding
	NextTab  key.Binding
}

// DefaultSidebarKeyMap returns standard sidebar keybindings.
func DefaultSidebarKeyMap() SidebarKeyMap {
	return SidebarKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open"),
		),
		New: key.NewBinding(
			key.WithKeys("n", "ctrl+n"),
			key.WithHelp("n", "new chapter"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("[", "h"),
			key.WithHelp("[/h", "prev tab"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("]", "l"),
			key.WithHelp("]/l", "next tab"),
		),
	}
}

// SidebarModel manages the left sidebar panel state and rendering.
type SidebarModel struct {
	chapterRepo   domain.ChapterRepository
	characterRepo domain.CharacterRepository

	ActiveTab       SidebarTab
	Chapters        []domain.Chapter
	Characters      []domain.Character
	SelectedChapter int
	SelectedChar    int

	Focused bool
	Width   int
	Height  int

	styles theme.Styles
	keys   SidebarKeyMap
}

// NewSidebarModel creates a new SidebarModel.
func NewSidebarModel(
	chapterRepo domain.ChapterRepository,
	characterRepo domain.CharacterRepository,
	styles theme.Styles,
) SidebarModel {
	return SidebarModel{
		chapterRepo:   chapterRepo,
		characterRepo: characterRepo,
		ActiveTab:     TabChapters,
		styles:        styles,
		keys:          DefaultSidebarKeyMap(),
	}
}

// Init loads initial chapters and characters.
func (m SidebarModel) Init() tea.Cmd {
	if m.chapterRepo == nil && m.characterRepo == nil {
		return nil
	}
	return m.ReloadDataCmd()
}

// ReloadDataCmd fetches chapters and characters from repositories.
func (m *SidebarModel) ReloadDataCmd() tea.Cmd {
	if m.chapterRepo == nil && m.characterRepo == nil {
		return nil
	}
	chapterRepo := m.chapterRepo
	characterRepo := m.characterRepo
	return func() tea.Msg {
		var chapters []domain.Chapter
		var chars []domain.Character
		if chapterRepo != nil {
			chapters, _ = chapterRepo.ListAll()
		}
		if characterRepo != nil {
			chars, _ = characterRepo.ListAll()
		}
		return dataLoadedMsg{chapters: chapters, characters: chars}
	}
}

type dataLoadedMsg struct {
	chapters   []domain.Chapter
	characters []domain.Character
}

// Update handles incoming events for the sidebar.
func (m SidebarModel) Update(msg tea.Msg) (SidebarModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case dataLoadedMsg:
		m.Chapters = msg.chapters
		m.Characters = msg.characters
		if m.SelectedChapter >= len(m.Chapters) && len(m.Chapters) > 0 {
			m.SelectedChapter = len(m.Chapters) - 1
		}
		if m.SelectedChar >= len(m.Characters) && len(m.Characters) > 0 {
			m.SelectedChar = len(m.Characters) - 1
		}

	case messages.ReloadChaptersMsg:
		cmds = append(cmds, m.ReloadDataCmd())

	case messages.ChapterCreatedMsg:
		m.Chapters = append(m.Chapters, msg.Chapter)
		m.SelectedChapter = len(m.Chapters) - 1

	case messages.FocusMsg:
		m.Focused = (msg.Target == messages.FocusSidebar)

	case tea.KeyMsg:
		if !m.Focused {
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.PrevTab):
			if m.ActiveTab == TabLore {
				m.ActiveTab = TabChapters
			}
		case key.Matches(msg, m.keys.NextTab):
			if m.ActiveTab == TabChapters {
				m.ActiveTab = TabLore
			}
		case key.Matches(msg, m.keys.Up):
			if m.ActiveTab == TabChapters {
				if m.SelectedChapter > 0 {
					m.SelectedChapter--
				}
			} else {
				if m.SelectedChar > 0 {
					m.SelectedChar--
				}
			}
		case key.Matches(msg, m.keys.Down):
			if m.ActiveTab == TabChapters {
				if m.SelectedChapter < len(m.Chapters)-1 {
					m.SelectedChapter++
				}
			} else {
				if m.SelectedChar < len(m.Characters)-1 {
					m.SelectedChar++
				}
			}
		case key.Matches(msg, m.keys.Select):
			if m.ActiveTab == TabChapters && len(m.Chapters) > 0 && m.SelectedChapter < len(m.Chapters) {
				selected := m.Chapters[m.SelectedChapter]
				return m, func() tea.Msg {
					return messages.ChapterSelectedMsg{Chapter: selected}
				}
			}
		case key.Matches(msg, m.keys.New):
			if m.ActiveTab == TabChapters {
				return m, func() tea.Msg {
					return messages.ShowModalMsg{
						Purpose: messages.ModalPurposeNewChapter,
						Title:   "Nuevo Capítulo",
						Prompt:  "Título del nuevo capítulo:",
					}
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// SetRepositories updates the active repositories and reloads data.
func (m *SidebarModel) SetRepositories(chapterRepo domain.ChapterRepository, characterRepo domain.CharacterRepository) tea.Cmd {
	m.chapterRepo = chapterRepo
	m.characterRepo = characterRepo
	m.SelectedChapter = 0
	m.SelectedChar = 0
	return m.ReloadDataCmd()
}
func (m *SidebarModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

// View renders the sidebar panel.
func (m SidebarModel) View() string {
	contentWidth := m.Width - 2 // account for borders
	if contentWidth < 0 {
		contentWidth = 0
	}

	// 1. Header with Tabs
	var tab1, tab2 string
	if m.ActiveTab == TabChapters {
		tab1 = m.styles.TabActive.Render("1: Chapters")
		tab2 = m.styles.TabInactive.Render("2: Lore")
	} else {
		tab1 = m.styles.TabInactive.Render("1: Chapters")
		tab2 = m.styles.TabActive.Render("2: Lore")
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top, tab1, " ", tab2)
	header = m.styles.SidebarHeader.Width(contentWidth).Render(header)

	// 2. Tab Content
	var body string
	if m.ActiveTab == TabChapters {
		body = m.renderChaptersList(contentWidth)
	} else {
		body = m.renderLoreView(contentWidth)
	}

	// Combine header and body
	fullContent := lipgloss.JoinVertical(lipgloss.Left, header, body)

	// Apply panel style based on focus
	panelStyle := m.styles.BlurredPanel
	if m.Focused {
		panelStyle = m.styles.FocusedPanel
	}

	return panelStyle.
		Width(m.Width).
		Height(m.Height).
		Render(fullContent)
}

func (m SidebarModel) renderChaptersList(width int) string {
	if len(m.Chapters) == 0 {
		return m.styles.ListSubtitle.Render("\n  No chapters found.\n  Press 'n' to create one.")
	}

	var items []string
	for i, chap := range m.Chapters {
		prefix := "  "
		if i == m.SelectedChapter {
			prefix = "▶ "
		}

		title := chap.Title
		if len(title) > width-10 {
			if width > 13 {
				title = title[:width-13] + "..."
			}
		}

		itemText := fmt.Sprintf("%s%s", prefix, title)
		var renderedItem string
		if i == m.SelectedChapter {
			renderedItem = m.styles.ListItemActive.Width(width).Render(itemText)
		} else {
			renderedItem = m.styles.ListItem.Width(width).Render(itemText)
		}

		subText := fmt.Sprintf("    %d words", chap.WordCount)
		renderedSub := m.styles.ListSubtitle.Render(subText)

		items = append(items, renderedItem, renderedSub)
	}

	return strings.Join(items, "\n")
}

func (m SidebarModel) renderLoreView(width int) string {
	if len(m.Characters) == 0 {
		return m.styles.ListSubtitle.Render("\n  No characters in lore.\n  Edit characters.json to add.")
	}

	var items []string
	// List of character names
	for i, char := range m.Characters {
		prefix := "  "
		if i == m.SelectedChar {
			prefix = "◆ "
		}

		itemText := fmt.Sprintf("%s%s (%s)", prefix, char.Name, char.Role)
		if i == m.SelectedChar {
			items = append(items, m.styles.ListItemActive.Width(width).Render(itemText))
		} else {
			items = append(items, m.styles.ListItem.Width(width).Render(itemText))
		}
	}

	listSection := strings.Join(items, "\n")

	// Character details card
	if m.SelectedChar < len(m.Characters) {
		selected := m.Characters[m.SelectedChar]
		cardContent := lipgloss.JoinVertical(
			lipgloss.Left,
			m.styles.CardName.Render(selected.Name),
			m.styles.CardRole.Render("Role: "+selected.Role),
			m.styles.CardDescription.Width(width-4).Render(selected.Description),
			m.styles.CardNotes.Width(width-4).Render("Notes: "+selected.Notes),
		)
		card := m.styles.CardContainer.Width(width - 2).Render(cardContent)
		return lipgloss.JoinVertical(lipgloss.Left, listSection, card)
	}

	return listSection
}
