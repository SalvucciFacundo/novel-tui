package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
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
	TabCharacters
	TabNotes

	// TabLore is an alias for TabCharacters for backward compatibility.
	TabLore = TabCharacters
)

// SidebarKeyMap defines keybindings for the sidebar.
type SidebarKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Select  key.Binding
	New     key.Binding
	PrevTab key.Binding
	NextTab key.Binding
	Save    key.Binding
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
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save notes"),
		),
	}
}

// SidebarModel manages the left sidebar panel state, 3-tab navigation, and notes editing.
type SidebarModel struct {
	chapterRepo   domain.ChapterRepository
	characterRepo domain.CharacterRepository
	novelPath     string

	ActiveTab       SidebarTab
	Chapters        []domain.Chapter
	Characters      []domain.Character
	SelectedChapter int
	SelectedChar    int

	notesTextarea textarea.Model

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
	ta := textarea.New()
	ta.Placeholder = "Notas de la novela..."
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(styles.AppContainer.GetBackground())

	return SidebarModel{
		chapterRepo:   chapterRepo,
		characterRepo: characterRepo,
		ActiveTab:     TabChapters,
		notesTextarea: ta,
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

// SetNovelPath updates the active novel root directory and loads notas.txt if present.
func (m *SidebarModel) SetNovelPath(path string) {
	m.novelPath = path
	if path == "" {
		m.notesTextarea.SetValue("")
		return
	}

	notesFile := filepath.Join(path, "notas.txt")
	data, err := os.ReadFile(notesFile)
	if err == nil {
		m.notesTextarea.SetValue(string(data))
	} else {
		m.notesTextarea.SetValue("")
	}
}

// SaveNotes persists current notesTextarea contents to notas.txt.
func (m *SidebarModel) SaveNotes() error {
	if m.novelPath == "" {
		return nil
	}
	notesFile := filepath.Join(m.novelPath, "notas.txt")
	return os.WriteFile(notesFile, []byte(m.notesTextarea.Value()), 0644)
}

// NotesValue returns current notes text buffer.
func (m SidebarModel) NotesValue() string {
	return m.notesTextarea.Value()
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
		if m.Focused && m.ActiveTab == TabNotes {
			cmds = append(cmds, m.notesTextarea.Focus())
		} else {
			m.notesTextarea.Blur()
		}

	case messages.SelectSidebarTabMsg:
		tab := SidebarTab(msg.Tab)
		if tab >= TabChapters && tab <= TabNotes {
			m.ActiveTab = tab
			if m.ActiveTab == TabNotes && m.Focused {
				cmds = append(cmds, m.notesTextarea.Focus())
			}
		}

	case messages.SaveRequestedMsg:
		_ = m.SaveNotes()

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			if m.ActiveTab == TabChapters {
				if m.SelectedChapter > 0 {
					m.SelectedChapter--
				}
			} else if m.ActiveTab == TabCharacters {
				if m.SelectedChar > 0 {
					m.SelectedChar--
				}
			} else {
				m.notesTextarea.CursorUp()
			}
			return m, nil

		case tea.MouseWheelDown:
			if m.ActiveTab == TabChapters {
				if m.SelectedChapter < len(m.Chapters)-1 {
					m.SelectedChapter++
				}
			} else if m.ActiveTab == TabCharacters {
				if m.SelectedChar < len(m.Characters)-1 {
					m.SelectedChar++
				}
			} else {
				m.notesTextarea.CursorDown()
			}
			return m, nil

		case tea.MouseLeft:
			m.Focused = true
			if msg.Y <= 2 {
				// Header tabs click detection
				third := m.Width / 3
				if third <= 0 {
					third = 10
				}
				if msg.X < third {
					m.ActiveTab = TabChapters
					m.notesTextarea.Blur()
				} else if msg.X < third*2 {
					m.ActiveTab = TabCharacters
					m.notesTextarea.Blur()
				} else {
					m.ActiveTab = TabNotes
					cmds = append(cmds, m.notesTextarea.Focus())
				}
				return m, tea.Batch(cmds...)
			}

			if msg.Y >= 3 {
				if m.ActiveTab == TabChapters {
					chapIdx := (msg.Y - 3) / 2
					if chapIdx >= 0 && chapIdx < len(m.Chapters) {
						m.SelectedChapter = chapIdx
						selected := m.Chapters[chapIdx]
						return m, func() tea.Msg {
							return messages.ChapterSelectedMsg{Chapter: selected}
						}
					}
				} else if m.ActiveTab == TabCharacters {
					charIdx := msg.Y - 3
					if charIdx >= 0 && charIdx < len(m.Characters) {
						m.SelectedChar = charIdx
						return m, nil
					}
				} else if m.ActiveTab == TabNotes {
					var taCmd tea.Cmd
					m.notesTextarea, taCmd = m.notesTextarea.Update(msg)
					cmds = append(cmds, taCmd)
					return m, tea.Batch(cmds...)
				}
			}
			return m, nil
		}

	case tea.KeyMsg:
		if !m.Focused {
			return m, nil
		}

		if m.ActiveTab == TabNotes {
			if key.Matches(msg, m.keys.Save) {
				_ = m.SaveNotes()
				return m, nil
			}
			// Allow tab cycling shortcuts even in notes
			if key.Matches(msg, m.keys.PrevTab) && msg.String() == "[" {
				m.ActiveTab = TabCharacters
				m.notesTextarea.Blur()
				return m, nil
			}
			if key.Matches(msg, m.keys.NextTab) && msg.String() == "]" {
				m.ActiveTab = TabChapters
				m.notesTextarea.Blur()
				return m, nil
			}

			var taCmd tea.Cmd
			m.notesTextarea, taCmd = m.notesTextarea.Update(msg)
			cmds = append(cmds, taCmd)
			return m, tea.Batch(cmds...)
		}

		// In TabChapters or TabCharacters:
		switch {
		case msg.String() == "1":
			m.ActiveTab = TabChapters
		case msg.String() == "2":
			m.ActiveTab = TabCharacters
		case msg.String() == "3":
			m.ActiveTab = TabNotes
			cmds = append(cmds, m.notesTextarea.Focus())
		case key.Matches(msg, m.keys.PrevTab):
			if m.ActiveTab == TabNotes {
				m.ActiveTab = TabCharacters
			} else if m.ActiveTab == TabCharacters {
				m.ActiveTab = TabChapters
			} else {
				m.ActiveTab = TabNotes
				cmds = append(cmds, m.notesTextarea.Focus())
			}
		case key.Matches(msg, m.keys.NextTab):
			if m.ActiveTab == TabChapters {
				m.ActiveTab = TabCharacters
			} else if m.ActiveTab == TabCharacters {
				m.ActiveTab = TabNotes
				cmds = append(cmds, m.notesTextarea.Focus())
			} else {
				m.ActiveTab = TabChapters
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

// SetSize updates the dimensions allocated to the sidebar panel.
func (m *SidebarModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h

	innerW := w - 4
	innerH := h - 4
	if innerW < 10 {
		innerW = 10
	}
	if innerH < 5 {
		innerH = 5
	}
	m.notesTextarea.SetWidth(innerW)
	m.notesTextarea.SetHeight(innerH)
}

// View renders the 3-tab sidebar panel.
func (m SidebarModel) View() string {
	contentWidth := m.Width - 2 // account for borders
	if contentWidth < 0 {
		contentWidth = 0
	}

	// 1. Header with 3 Tabs
	var tab1, tab2, tab3 string
	if m.ActiveTab == TabChapters {
		tab1 = m.styles.TabActive.Render("1: Capítulos")
		tab2 = m.styles.TabInactive.Render("2: Personajes")
		tab3 = m.styles.TabInactive.Render("3: Notas")
	} else if m.ActiveTab == TabCharacters {
		tab1 = m.styles.TabInactive.Render("1: Capítulos")
		tab2 = m.styles.TabActive.Render("2: Personajes")
		tab3 = m.styles.TabInactive.Render("3: Notas")
	} else {
		tab1 = m.styles.TabInactive.Render("1: Capítulos")
		tab2 = m.styles.TabInactive.Render("2: Personajes")
		tab3 = m.styles.TabActive.Render("3: Notas")
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top, tab1, " ", tab2, " ", tab3)
	header = m.styles.SidebarHeader.Width(contentWidth).Render(header)

	// 2. Tab Content
	var body string
	switch m.ActiveTab {
	case TabChapters:
		body = m.renderChaptersList(contentWidth)
	case TabCharacters:
		body = m.renderLoreView(contentWidth)
	case TabNotes:
		body = m.renderNotesView(contentWidth)
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
		return m.styles.ListSubtitle.Render("\n  No hay capítulos.\n  Presiona 'n' para crear uno.")
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

		subText := fmt.Sprintf("    %d palabras", chap.WordCount)
		renderedSub := m.styles.ListSubtitle.Render(subText)

		items = append(items, renderedItem, renderedSub)
	}

	return strings.Join(items, "\n")
}

func (m SidebarModel) renderLoreView(width int) string {
	if len(m.Characters) == 0 {
		return m.styles.ListSubtitle.Render("\n  No hay personajes.\n  Edita personajes.json para agregar.")
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
			m.styles.CardRole.Render("Rol: "+selected.Role),
			m.styles.CardDescription.Width(width-4).Render(selected.Description),
			m.styles.CardNotes.Width(width-4).Render("Notas: "+selected.Notes),
		)
		card := m.styles.CardContainer.Width(width - 2).Render(cardContent)
		return lipgloss.JoinVertical(lipgloss.Left, listSection, card)
	}

	return listSection
}

func (m SidebarModel) renderNotesView(width int) string {
	return lipgloss.NewStyle().
		Padding(0, 1).
		Width(width).
		Render(m.notesTextarea.View())
}
