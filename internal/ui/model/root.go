package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/components"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

const (
	MinWidth  = 60
	MinHeight = 15

	SidebarDefaultWidth = 28
)

// RootKeyMap defines global keybindings.
type RootKeyMap struct {
	Quit     key.Binding
	NextTab  key.Binding
	PrevTab  key.Binding
	Save     key.Binding
}

// DefaultRootKeyMap returns standard root navigation keys.
func DefaultRootKeyMap() RootKeyMap {
	return RootKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch focus"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "switch focus back"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
	}
}

// RootModel represents the root TUI state.
type RootModel struct {
	chapterRepo   domain.ChapterRepository
	characterRepo domain.CharacterRepository

	sidebar   components.SidebarModel
	editor    components.EditorModel
	statusbar components.StatusBarModel

	activeFocus messages.FocusState
	width       int
	height      int
	ready       bool

	styles theme.Styles
	keys   RootKeyMap
}

// NewRootModel constructs the RootModel with all child components.
func NewRootModel(
	chapterRepo domain.ChapterRepository,
	characterRepo domain.CharacterRepository,
) RootModel {
	styles := theme.DefaultStyles
	return RootModel{
		chapterRepo:   chapterRepo,
		characterRepo: characterRepo,
		sidebar:       components.NewSidebarModel(chapterRepo, characterRepo, styles),
		editor:        components.NewEditorModel(styles),
		statusbar:     components.NewStatusBarModel(styles),
		activeFocus:   messages.FocusSidebar,
		styles:        styles,
		keys:          DefaultRootKeyMap(),
	}
}

// Init initializes the sub-components.
func (m RootModel) Init() tea.Cmd {
	return tea.Batch(
		m.sidebar.Init(),
		m.editor.Init(),
		m.statusbar.Init(),
		func() tea.Msg { return messages.FocusMsg{Target: m.activeFocus} },
	)
}

// Update coordinates events and dispatches messages to child components.
func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.recalculateLayout()

	case messages.SaveRequestedMsg:
		chapterID := msg.ChapterID
		content := msg.Content
		repo := m.chapterRepo
		return m, func() tea.Msg {
			err := repo.SaveContent(chapterID, content)
			return messages.SaveCompletedMsg{
				ChapterID: chapterID,
				Success:   err == nil,
				Error:     err,
			}
		}

	case messages.ChapterSelectedMsg:
		// When a chapter is selected, auto-focus editor if content loaded
		var editorCmd tea.Cmd
		m.editor, editorCmd = m.editor.Update(msg)
		cmds = append(cmds, editorCmd)

		var statusCmd tea.Cmd
		m.statusbar, statusCmd = m.statusbar.Update(msg)
		cmds = append(cmds, statusCmd)

		return m, tea.Batch(cmds...)

	case messages.ChapterCreatedMsg:
		var sCmd tea.Cmd
		m.sidebar, sCmd = m.sidebar.Update(msg)
		cmds = append(cmds, sCmd)

		// Focus editor on new chapter creation
		m.activeFocus = messages.FocusEditor
		cmds = append(cmds, func() tea.Msg { return messages.FocusMsg{Target: messages.FocusEditor} })

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.NextTab):
			if m.activeFocus == messages.FocusSidebar {
				m.activeFocus = messages.FocusEditor
			} else {
				m.activeFocus = messages.FocusSidebar
			}
			focusMsg := messages.FocusMsg{Target: m.activeFocus}
			var sCmd, eCmd tea.Cmd
			m.sidebar, sCmd = m.sidebar.Update(focusMsg)
			m.editor, eCmd = m.editor.Update(focusMsg)
			return m, tea.Batch(sCmd, eCmd)

		case key.Matches(msg, m.keys.PrevTab):
			if m.activeFocus == messages.FocusEditor {
				m.activeFocus = messages.FocusSidebar
			} else {
				m.activeFocus = messages.FocusEditor
			}
			focusMsg := messages.FocusMsg{Target: m.activeFocus}
			var sCmd, eCmd tea.Cmd
			m.sidebar, sCmd = m.sidebar.Update(focusMsg)
			m.editor, eCmd = m.editor.Update(focusMsg)
			return m, tea.Batch(sCmd, eCmd)
		}
	}

	// Forward messages to active components
	var sidebarCmd, editorCmd, statusCmd tea.Cmd
	m.sidebar, sidebarCmd = m.sidebar.Update(msg)
	m.editor, editorCmd = m.editor.Update(msg)
	m.statusbar, statusCmd = m.statusbar.Update(msg)

	cmds = append(cmds, sidebarCmd, editorCmd, statusCmd)
	return m, tea.Batch(cmds...)
}

func (m *RootModel) recalculateLayout() {
	if m.width < MinWidth || m.height < MinHeight {
		return
	}

	statusHeight := 1
	mainHeight := m.height - statusHeight

	sidebarWidth := SidebarDefaultWidth
	if sidebarWidth > m.width/2 {
		sidebarWidth = m.width / 3
	}
	editorWidth := m.width - sidebarWidth

	m.sidebar.SetSize(sidebarWidth, mainHeight)
	m.editor.SetSize(editorWidth, mainHeight)
	m.statusbar.SetWidth(m.width)
}

// View renders the multi-panel TUI.
func (m RootModel) View() string {
	if !m.ready {
		return "Initializing Novel TUI..."
	}

	// 1. Terminal window too small check
	if m.width < MinWidth || m.height < MinHeight {
		warning := fmt.Sprintf(
			"Terminal size too small: (%dx%d)\nPlease resize your window to at least %dx%d.",
			m.width, m.height, MinWidth, MinHeight,
		)
		return m.styles.WarningView.
			Width(m.width).
			Height(m.height).
			Render(warning)
	}

	// 2. Main content panels (Sidebar + Editor side-by-side)
	sidebarView := m.sidebar.View()
	editorView := m.editor.View()
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, editorView)

	// 3. Status bar at the bottom
	statusView := m.statusbar.View()

	// 4. Combined full application view
	return lipgloss.JoinVertical(lipgloss.Left, mainView, statusView)
}
