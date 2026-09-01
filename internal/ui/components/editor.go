package components

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/service"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// EditorKeyMap defines keybindings for the editor.
type EditorKeyMap struct {
	Save key.Binding
}

// DefaultEditorKeyMap returns standard editor keybindings.
func DefaultEditorKeyMap() EditorKeyMap {
	return EditorKeyMap{
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save chapter"),
		),
	}
}

// EditorModel wraps textarea.Model and provides rich novel writing features.
type EditorModel struct {
	textarea textarea.Model

	ActiveChapter domain.Chapter
	Metrics       domain.EditorMetrics
	IsDirty       bool
	Focused       bool

	Width  int
	Height int

	styles theme.Styles
	keys   EditorKeyMap
}

// NewEditorModel constructs an EditorModel.
func NewEditorModel(styles theme.Styles) EditorModel {
	ta := textarea.New()
	ta.Placeholder = "Begin writing your story here..."
	ta.ShowLineNumbers = true
	ta.CharLimit = 0 // unlimited

	// Custom styles for textarea
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(styles.AppContainer.GetBackground())
	ta.FocusedStyle.LineNumber = lipgloss.NewStyle().Foreground(theme.CurrentTheme.Muted)
	ta.BlurredStyle.LineNumber = lipgloss.NewStyle().Foreground(theme.CurrentTheme.BorderBlurred)

	return EditorModel{
		textarea: ta,
		styles:   styles,
		keys:     DefaultEditorKeyMap(),
	}
}

// Init initializes the editor.
func (m EditorModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages for the editor component.
func (m EditorModel) Update(msg tea.Msg) (EditorModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messages.FocusMsg:
		m.Focused = (msg.Target == messages.FocusEditor)
		if m.Focused {
			cmds = append(cmds, m.textarea.Focus())
		} else {
			m.textarea.Blur()
		}

	case messages.ChapterSelectedMsg:
		m.ActiveChapter = msg.Chapter
		m.textarea.SetValue(msg.Chapter.Content)
		m.IsDirty = false
		m.Metrics = service.CalculateMetrics(msg.Chapter.Content, false)

		return m, func() tea.Msg {
			return messages.TextChangedMsg{
				ChapterID: m.ActiveChapter.ID,
				Content:   msg.Chapter.Content,
				Metrics:   m.Metrics,
			}
		}

	case messages.SaveCompletedMsg:
		if msg.Success && msg.ChapterID == m.ActiveChapter.ID {
			m.IsDirty = false
			m.Metrics.IsDirty = false
		}

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Save) {
			content := m.textarea.Value()
			chapterID := m.ActiveChapter.ID
			if chapterID != "" {
				return m, func() tea.Msg {
					return messages.SaveRequestedMsg{
						ChapterID: chapterID,
						Content:   content,
					}
				}
			}
		}

		if !m.Focused {
			return m, nil
		}

		prevValue := m.textarea.Value()
		var taCmd tea.Cmd
		m.textarea, taCmd = m.textarea.Update(msg)
		cmds = append(cmds, taCmd)

		curValue := m.textarea.Value()
		if curValue != prevValue {
			m.IsDirty = true
			m.Metrics = service.CalculateMetrics(curValue, true)
			cmds = append(cmds, func() tea.Msg {
				return messages.TextChangedMsg{
					ChapterID: m.ActiveChapter.ID,
					Content:   curValue,
					Metrics:   m.Metrics,
				}
			})
		}
		return m, tea.Batch(cmds...)
	}

	if m.Focused {
		var taCmd tea.Cmd
		m.textarea, taCmd = m.textarea.Update(msg)
		cmds = append(cmds, taCmd)
	}

	return m, tea.Batch(cmds...)
}

// SetSize sets allocated dimensions for the editor component.
func (m *EditorModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h

	// Internal textarea size within border
	innerW := w - 4
	innerH := h - 2
	if innerW < 10 {
		innerW = 10
	}
	if innerH < 5 {
		innerH = 5
	}
	m.textarea.SetWidth(innerW)
	m.textarea.SetHeight(innerH)
}

// Value returns the current text in the buffer.
func (m EditorModel) Value() string {
	return m.textarea.Value()
}

// View renders the editor panel.
func (m EditorModel) View() string {
	panelStyle := m.styles.BlurredPanel
	if m.Focused {
		panelStyle = m.styles.FocusedPanel
	}

	return panelStyle.
		Width(m.Width).
		Height(m.Height).
		Render(m.textarea.View())
}
