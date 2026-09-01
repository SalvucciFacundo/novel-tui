package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// ModalKeyMap defines keybindings for the input modal.
type ModalKeyMap struct {
	Submit key.Binding
	Cancel key.Binding
}

// DefaultModalKeyMap returns standard modal keys.
func DefaultModalKeyMap() ModalKeyMap {
	return ModalKeyMap{
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// ModalModel is a reusable centered floating dialog with a text input.
type ModalModel struct {
	Active   bool
	Purpose  messages.ModalPurpose
	Title    string
	Prompt   string
	ErrorMsg string

	input textinput.Model
	keys  ModalKeyMap

	width  int
	height int
	styles theme.Styles
}

// NewModalModel creates a new ModalModel.
func NewModalModel(styles theme.Styles) ModalModel {
	ti := textinput.New()
	ti.Focus()
	ti.Prompt = "❯ "
	ti.CharLimit = 120
	ti.Width = 40

	return ModalModel{
		Active: false,
		input:  ti,
		keys:   DefaultModalKeyMap(),
		styles: styles,
	}
}

// Show opens the modal with the specified context and pre-filled value.
func (m *ModalModel) Show(purpose messages.ModalPurpose, title, prompt, initialValue string) {
	m.Active = true
	m.Purpose = purpose
	m.Title = title
	m.Prompt = prompt
	m.ErrorMsg = ""

	m.input.SetValue(initialValue)
	m.input.Focus()
	m.input.CursorEnd()
}

// Hide closes and resets the modal.
func (m *ModalModel) Hide() {
	m.Active = false
	m.ErrorMsg = ""
	m.input.Blur()
}

// SetError displays an error message inside the modal dialog.
func (m *ModalModel) SetError(err string) {
	m.ErrorMsg = err
}

// SetSize updates terminal dimensions for centering calculations.
func (m *ModalModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Init initializes the modal component.
func (m ModalModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles keyboard and input events for the modal dialog.
func (m ModalModel) Update(msg tea.Msg) (ModalModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.ShowModalMsg:
		m.Active = true
		m.Purpose = msg.Purpose
		m.Title = msg.Title
		m.Prompt = msg.Prompt
		m.ErrorMsg = msg.ErrorMsg
		m.input.SetValue(msg.InitialValue)
		m.input.Focus()
		m.input.CursorEnd()
		return m, textinput.Blink

	case messages.HideModalMsg:
		m.Active = false
		m.ErrorMsg = ""
		m.input.Blur()
		return m, nil

	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			var tiCmd tea.Cmd
			m.input, tiCmd = m.input.Update(msg)
			return m, tiCmd
		}
		return m, nil
	}

	if !m.Active {
		return m, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel):
			m.Active = false
			m.ErrorMsg = ""
			m.input.Blur()
			return m, func() tea.Msg { return messages.HideModalMsg{} }

		case key.Matches(msg, m.keys.Submit):
			val := strings.TrimSpace(m.input.Value())
			if val == "" {
				m.ErrorMsg = "El campo no puede estar vacío"
				return m, nil
			}

			purpose := m.Purpose
			m.Active = false
			m.ErrorMsg = ""
			m.input.Blur()

			// Dispatch both generic SubmitModalMsg and contextual typed messages
			return m, func() tea.Msg {
				switch purpose {
				case messages.ModalPurposeNewNovel:
					return messages.CreateNovelMsg{Title: val}
				case messages.ModalPurposeNewChapter:
					return messages.CreateChapterMsg{Title: val}
				case messages.ModalPurposeSetRootDir:
					return messages.SetRootDirMsg{Path: val}
				case messages.ModalPurposeOpenFolder:
					return messages.OpenNovelMsg{Path: val}
				default:
					return messages.SubmitModalMsg{Purpose: purpose, Value: val}
				}
			}
		}
	}

	var tiCmd tea.Cmd
	m.input, tiCmd = m.input.Update(msg)
	cmds = append(cmds, tiCmd)

	return m, tea.Batch(cmds...)
}

// View renders the centered modal dialog overlay.
func (m ModalModel) View() string {
	if !m.Active {
		return ""
	}

	modalWidth := 50
	if m.width > 0 && modalWidth > m.width-4 {
		modalWidth = m.width - 4
	}
	if modalWidth < 30 {
		modalWidth = 30
	}

	// 1. Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.CurrentTheme.Highlight).
		MarginBottom(1)
	titleText := titleStyle.Render(m.Title)

	// 2. Prompt
	promptStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme.Foreground).
		MarginBottom(1)
	var promptText string
	if m.Prompt != "" {
		promptText = promptStyle.Render(m.Prompt)
	}

	// 3. Text input box
	inputBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(theme.CurrentTheme.BorderFocused).
		Padding(0, 1).
		Width(modalWidth - 6)
	inputBox := inputBoxStyle.Render(m.input.View())

	// 4. Error message if present
	var errBox string
	if m.ErrorMsg != "" {
		errStyle := lipgloss.NewStyle().
			Foreground(theme.CurrentTheme.Error).
			Bold(true).
			MarginTop(1)
		errBox = errStyle.Render("⚠ " + m.ErrorMsg)
	}

	// 5. Help / Footer hints
	helpStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme.Muted).
		MarginTop(1)
	helpText := helpStyle.Render("[Enter] Confirmar   [Esc] Cancelar")

	// Assemble contents
	var elements []string
	elements = append(elements, titleText)
	if promptText != "" {
		elements = append(elements, promptText)
	}
	elements = append(elements, inputBox)
	if errBox != "" {
		elements = append(elements, errBox)
	}
	elements = append(elements, helpText)

	content := lipgloss.JoinVertical(lipgloss.Left, elements...)

	// Outer card
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.CurrentTheme.BorderFocused).
		Background(theme.CurrentTheme.CardBg).
		Padding(1, 2).
		Width(modalWidth)

	dialog := cardStyle.Render(content)

	if m.width == 0 || m.height == 0 {
		return dialog
	}

	// Center dialog within terminal bounds
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}
