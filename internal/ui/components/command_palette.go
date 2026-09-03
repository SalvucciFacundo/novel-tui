package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// CommandPaletteKeyMap defines keyboard shortcuts for the Command Palette modal.
type CommandPaletteKeyMap struct {
	Close  key.Binding
	Submit key.Binding
	Up     key.Binding
	Down   key.Binding
}

// DefaultCommandPaletteKeyMap returns standard command palette keybindings.
func DefaultCommandPaletteKeyMap() CommandPaletteKeyMap {
	return CommandPaletteKeyMap{
		Close: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cerrar"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "ejecutar"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "ctrl+p"),
			key.WithHelp("↑", "anterior"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "ctrl+n"),
			key.WithHelp("↓", "siguiente"),
		),
	}
}

// CommandPaletteModel is a centered modal for searching and executing actions/shortcuts.
type CommandPaletteModel struct {
	Active      bool
	SearchInput textinput.Model
	Commands    []domain.CommandItem
	Filtered    []domain.CommandItem
	CursorIndex int

	width  int
	height int
	styles theme.Styles
	keys   CommandPaletteKeyMap
}

// NewCommandPaletteModel constructs a new CommandPaletteModel.
func NewCommandPaletteModel(styles theme.Styles) CommandPaletteModel {
	ti := textinput.New()
	ti.Placeholder = "Buscar comando o atajo... (Ctrl+P)"
	ti.Prompt = "⌨️  "
	ti.CharLimit = 100
	ti.Width = 60
	ti.Focus()

	defaultCmds := domain.DefaultCommands()
	filtered := make([]domain.CommandItem, len(defaultCmds))
	copy(filtered, defaultCmds)

	return CommandPaletteModel{
		Active:      false,
		SearchInput: ti,
		Commands:    defaultCmds,
		Filtered:    filtered,
		CursorIndex: 0,
		styles:      styles,
		keys:        DefaultCommandPaletteKeyMap(),
	}
}

// Init initializes the command palette textinput.
func (m CommandPaletteModel) Init() tea.Cmd {
	return textinput.Blink
}

// SetSize updates the viewport dimensions for centering.
func (m *CommandPaletteModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetCommands replaces the available commands and re-applies filtering.
func (m *CommandPaletteModel) SetCommands(commands []domain.CommandItem) {
	m.Commands = commands
	m.PerformFilter()
}

// Show opens and focuses the command palette.
func (m *CommandPaletteModel) Show() {
	m.Active = true
	m.SearchInput.SetValue("")
	m.SearchInput.Focus()
	m.PerformFilter()
	m.CursorIndex = 0
}

// Hide closes and blurs the command palette.
func (m *CommandPaletteModel) Hide() {
	m.Active = false
	m.SearchInput.Blur()
}

// PerformFilter filters commands based on the search query.
func (m *CommandPaletteModel) PerformFilter() {
	query := strings.ToLower(strings.TrimSpace(m.SearchInput.Value()))
	if query == "" {
		m.Filtered = make([]domain.CommandItem, len(m.Commands))
		copy(m.Filtered, m.Commands)
	} else {
		var matched []domain.CommandItem
		for _, cmd := range m.Commands {
			if strings.Contains(strings.ToLower(cmd.Title), query) ||
				strings.Contains(strings.ToLower(cmd.Category), query) ||
				strings.Contains(strings.ToLower(cmd.Shortcut), query) ||
				strings.Contains(strings.ToLower(cmd.Description), query) {
				matched = append(matched, cmd)
			}
		}
		m.Filtered = matched
	}

	if len(m.Filtered) == 0 {
		m.CursorIndex = 0
	} else if m.CursorIndex >= len(m.Filtered) {
		m.CursorIndex = len(m.Filtered) - 1
	} else if m.CursorIndex < 0 {
		m.CursorIndex = 0
	}
}

// Update handles messages and keybindings for the command palette.
func (m CommandPaletteModel) Update(msg tea.Msg) (CommandPaletteModel, tea.Cmd) {
	switch msg.(type) {
	case messages.OpenCommandPaletteMsg:
		m.Show()
		return m, textinput.Blink

	case messages.CloseCommandPaletteMsg:
		m.Hide()
		return m, nil
	}

	if !m.Active {
		return m, nil
	}

	var cmds []tea.Cmd

	switch kMsg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(kMsg, m.keys.Close):
			m.Hide()
			return m, func() tea.Msg { return messages.CloseCommandPaletteMsg{} }

		case key.Matches(kMsg, m.keys.Up):
			if m.CursorIndex > 0 {
				m.CursorIndex--
			}
			return m, nil

		case key.Matches(kMsg, m.keys.Down):
			if m.CursorIndex < len(m.Filtered)-1 {
				m.CursorIndex++
			}
			return m, nil

		case key.Matches(kMsg, m.keys.Submit):
			if len(m.Filtered) > 0 && m.CursorIndex >= 0 && m.CursorIndex < len(m.Filtered) {
				selected := m.Filtered[m.CursorIndex]
				m.Hide()
				return m, func() tea.Msg {
					return messages.ExecuteCommandMsg{Command: selected}
				}
			}
			return m, nil
		}
	}


	// Update search input
	prevVal := m.SearchInput.Value()
	var tiCmd tea.Cmd
	m.SearchInput, tiCmd = m.SearchInput.Update(msg)
	cmds = append(cmds, tiCmd)

	if m.SearchInput.Value() != prevVal {
		m.PerformFilter()
	}

	return m, tea.Batch(cmds...)
}

// View renders the centered floating Command Palette modal.
func (m CommandPaletteModel) View() string {
	if !m.Active {
		return ""
	}

	modalWidth := 78
	if m.width > 0 && modalWidth > m.width-4 {
		modalWidth = m.width - 4
	}
	if modalWidth < 44 {
		modalWidth = 44
	}

	innerWidth := modalWidth - 6
	if innerWidth < 34 {
		innerWidth = 34
	}

	cardBg := theme.CurrentTheme.CardBg

	// 1. Header
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.CurrentTheme.Highlight).
		Background(cardBg).
		Width(innerWidth)
	title := titleStyle.Render("⌨️  Paleta de Comandos y Atajos")

	// 2. Search Input Box
	searchBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.CurrentTheme.BorderFocused).
		Background(cardBg).
		Padding(0, 1).
		Width(innerWidth).
		Render(m.SearchInput.View())

	// 3. Command List / Results
	const maxItems = 7
	var listContent string

	if len(m.Filtered) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(theme.CurrentTheme.Muted).
			Background(cardBg).
			Padding(1, 1).
			Width(innerWidth).
			Align(lipgloss.Center)

		query := m.SearchInput.Value()
		if query != "" {
			listContent = emptyStyle.Render(fmt.Sprintf("No se encontraron comandos para \"%s\"", query))
		} else {
			listContent = emptyStyle.Render("No hay comandos disponibles")
		}
	} else {
		// Calculate viewport window around CursorIndex
		start := 0
		if m.CursorIndex >= maxItems {
			start = m.CursorIndex - maxItems + 1
		}
		end := start + maxItems
		if end > len(m.Filtered) {
			end = len(m.Filtered)
		}

		var rows []string
		for i := start; i < end; i++ {
			cmd := m.Filtered[i]
			isSelected := i == m.CursorIndex

			// Category pill
			categoryStyle := lipgloss.NewStyle().
				Foreground(theme.CurrentTheme.Secondary).
				Background(theme.CurrentTheme.CardBg).
				Bold(false)
			catText := categoryStyle.Render(fmt.Sprintf("[%s]", cmd.Category))

			// Title
			titleWidth := innerWidth - len(cmd.Shortcut) - len(cmd.Category) - 10
			if titleWidth < 18 {
				titleWidth = 18
			}

			itemTitleStyle := lipgloss.NewStyle().
				Foreground(theme.CurrentTheme.Foreground)
			if isSelected {
				itemTitleStyle = itemTitleStyle.
					Foreground(theme.CurrentTheme.Highlight).
					Bold(true)
			}

			truncatedTitle := cmd.Title
			if len(truncatedTitle) > titleWidth {
				truncatedTitle = truncatedTitle[:titleWidth-3] + "..."
			}
			titleText := itemTitleStyle.Render(truncatedTitle)

			// Shortcut pill
			shortcutStyle := lipgloss.NewStyle().
				Foreground(theme.CurrentTheme.Accent).
				Background(theme.CurrentTheme.Background).
				Bold(true).
				Padding(0, 1)
			shortcutText := shortcutStyle.Render(cmd.Shortcut)

			// Description subtitle
			descStyle := lipgloss.NewStyle().
				Foreground(theme.CurrentTheme.Muted).
				Width(innerWidth - 6)
			truncatedDesc := cmd.Description
			if len(truncatedDesc) > innerWidth-6 {
				truncatedDesc = truncatedDesc[:innerWidth-9] + "..."
			}
			descText := descStyle.Render(truncatedDesc)

			// Top line of row: Category + Title + [spacing] + Shortcut
			leftSide := fmt.Sprintf("%s %s", catText, titleText)
			leftWidth := lipgloss.Width(leftSide)
			rightWidth := lipgloss.Width(shortcutText)
			gap := innerWidth - leftWidth - rightWidth - 4
			if gap < 1 {
				gap = 1
			}
			topLine := fmt.Sprintf(" %s%s%s", leftSide, strings.Repeat(" ", gap), shortcutText)
			bottomLine := fmt.Sprintf("   %s", descText)

			rowBlock := lipgloss.JoinVertical(lipgloss.Left, topLine, bottomLine)

			rowStyle := lipgloss.NewStyle().
				Width(innerWidth).
				Padding(0, 0)

			if isSelected {
				rowStyle = rowStyle.
					Background(theme.CurrentTheme.Background).
					Border(lipgloss.NormalBorder(), false, false, false, true).
					BorderForeground(theme.CurrentTheme.Highlight)
			} else {
				rowStyle = rowStyle.
					Background(cardBg).
					Border(lipgloss.NormalBorder(), false, false, false, true).
					BorderForeground(cardBg)
			}

			rows = append(rows, rowStyle.Render(rowBlock))
		}

		listContent = lipgloss.JoinVertical(lipgloss.Left, rows...)
	}

	listBox := lipgloss.NewStyle().
		Background(cardBg).
		Padding(0, 1).
		Width(innerWidth).
		Render(listContent)

	// 4. Footer
	footerStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme.Muted).
		Background(cardBg).
		Padding(0, 1).
		Width(innerWidth)
	footer := footerStyle.Render("[Enter] Ejecutar  [Esc] Cerrar  [↑/↓] Navegar")

	// Assemble modal box
	modalContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		searchBox,
		listBox,
		footer,
	)

	modalDialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.CurrentTheme.BorderFocused).
		Background(cardBg).
		Padding(1, 2).
		Width(modalWidth).
		Render(modalContent)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			modalDialog,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(theme.CurrentTheme.Background),
		)
	}

	return modalDialog
}
