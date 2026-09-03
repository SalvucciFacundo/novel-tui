package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/repository"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// ChatDrawerModel implements the AI writing assistant panel with viewport, textarea, and session manager.
type ChatDrawerModel struct {
	viewport textareaViewport
	textarea textarea.Model

	session           domain.ChatSession
	sessions          []domain.ChatSession
	sessionPickerOpen bool
	sessionCursor     int

	isGenerating bool
	effortLevel  domain.LLMEffortLevel

	novelDir    string
	sessionRepo domain.ChatSessionRepository

	width   int
	height  int
	Focused bool

	brainNotification string

	styles theme.Styles
}

type textareaViewport struct {
	viewport.Model
}

// NewChatDrawerModel constructs a new ChatDrawerModel.
func NewChatDrawerModel(sessionRepo domain.ChatSessionRepository, styles theme.Styles) ChatDrawerModel {
	if sessionRepo == nil {
		sessionRepo = repository.NewFileChatSessionRepository()
	}

	vp := viewport.New(30, 10)
	vp.SetContent("Inicia una conversación con tu asistente de escritura...")

	ta := textarea.New()
	ta.Placeholder = "Escribe una pregunta o petición (ej: 'Revisa este diálogo')..."
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetHeight(3)

	return ChatDrawerModel{
		viewport:    textareaViewport{vp},
		textarea:    ta,
		sessionRepo: sessionRepo,
		effortLevel: domain.EffortMedium,
		styles:      styles,
		session: domain.ChatSession{
			EffortLevel: domain.EffortMedium,
		},
	}
}

// Init initializes the chat drawer component.
func (m ChatDrawerModel) Init() tea.Cmd {
	return textarea.Blink
}

// SetNovelDir sets the working novel directory and loads sessions.
func (m *ChatDrawerModel) SetNovelDir(novelDir string) {
	m.novelDir = novelDir
	m.session.NovelPath = novelDir
	m.RefreshSessions()

	if len(m.sessions) > 0 {
		m.LoadSession(m.sessions[0].ID)
	} else {
		m.NewSession()
	}
}

// RefreshSessions reloads the list of sessions from the repository.
func (m *ChatDrawerModel) RefreshSessions() {
	if m.novelDir == "" || m.sessionRepo == nil {
		return
	}
	list, err := m.sessionRepo.List(m.novelDir)
	if err == nil {
		m.sessions = list
	}
}

// LoadSession loads a specific session by ID.
func (m *ChatDrawerModel) LoadSession(sessionID string) {
	if m.novelDir == "" || m.sessionRepo == nil {
		return
	}
	sess, err := m.sessionRepo.Get(m.novelDir, sessionID)
	if err != nil {
		return
	}
	m.session = sess
	if sess.EffortLevel != "" {
		m.effortLevel = sess.EffortLevel
	}
	m.renderHistory()
}

// NewSession creates a fresh, blank chat session.
func (m *ChatDrawerModel) NewSession() {
	m.session = domain.ChatSession{
		ID:          fmt.Sprintf("session_%s", time.Now().UTC().Format("20060102_150405")),
		Title:       "Nueva Conversación",
		NovelPath:   m.novelDir,
		EffortLevel: m.effortLevel,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Messages:    []domain.ChatMessage{},
	}
	m.renderHistory()
}

// ActiveSession returns the current active ChatSession.
func (m ChatDrawerModel) ActiveSession() domain.ChatSession {
	return m.session
}

// LastAssistantMessage returns the content of the most recent assistant message.
func (m ChatDrawerModel) LastAssistantMessage() string {
	for i := len(m.session.Messages) - 1; i >= 0; i-- {
		if m.session.Messages[i].Role == "assistant" && strings.TrimSpace(m.session.Messages[i].Content) != "" {
			return m.session.Messages[i].Content
		}
	}
	return ""
}

// IsGenerating returns whether a stream response is currently being received.
func (m ChatDrawerModel) IsGenerating() bool {
	return m.isGenerating
}

// EffortLevel returns the active effort level.
func (m ChatDrawerModel) EffortLevel() domain.LLMEffortLevel {
	return m.effortLevel
}

// SetEffortLevel updates the active effort level.
func (m *ChatDrawerModel) SetEffortLevel(effort domain.LLMEffortLevel) {
	m.effortLevel = effort
	m.session.EffortLevel = effort
	if m.novelDir != "" && len(m.session.Messages) > 0 {
		_ = m.sessionRepo.Save(m.session)
	}
}

// CycleEffortLevel rotates through Low -> Medium -> High -> Low.
func (m *ChatDrawerModel) CycleEffortLevel() domain.LLMEffortLevel {
	switch m.effortLevel {
	case domain.EffortLow:
		m.SetEffortLevel(domain.EffortMedium)
	case domain.EffortMedium:
		m.SetEffortLevel(domain.EffortHigh)
	case domain.EffortHigh:
		fallthrough
	default:
		m.SetEffortLevel(domain.EffortLow)
	}
	return m.effortLevel
}

// SetDimensions updates width and height of inner components.
func (m *ChatDrawerModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height

	contentWidth := width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	headerHeight := 3
	inputHeight := 4
	footerHeight := 2
	vpHeight := height - headerHeight - inputHeight - footerHeight - 2
	if vpHeight < 4 {
		vpHeight = 4
	}

	m.viewport.Width = contentWidth
	m.viewport.Height = vpHeight
	m.textarea.SetWidth(contentWidth)
	m.textarea.SetHeight(3)

	m.renderHistory()
}

// Update handles messages and keypresses for the chat drawer.
func (m ChatDrawerModel) Update(msg tea.Msg) (ChatDrawerModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messages.FocusMsg:
		m.Focused = (msg.Target == messages.FocusChat)
		if m.Focused {
			cmds = append(cmds, m.textarea.Focus())
		} else {
			m.textarea.Blur()
		}

	case messages.TokenReceivedMsg:
		if len(m.session.Messages) > 0 && m.session.Messages[len(m.session.Messages)-1].Role == "assistant" {
			lastIdx := len(m.session.Messages) - 1
			m.session.Messages[lastIdx].Content += msg.Content
		} else {
			m.session.Messages = append(m.session.Messages, domain.ChatMessage{
				Role:      "assistant",
				Content:   msg.Content,
				Timestamp: time.Now().UTC(),
			})
		}
		m.renderHistory()
		m.viewport.GotoBottom()

	case messages.StreamFinishedMsg:
		m.isGenerating = false
		if m.novelDir != "" {
			_ = m.sessionRepo.Save(m.session)
			m.RefreshSessions()
		}
		m.renderHistory()
		m.viewport.GotoBottom()

	case messages.StreamErrorMsg:
		m.isGenerating = false
		errMsg := "\n\n⚠️ Error: " + msg.Err.Error()
		if len(m.session.Messages) > 0 && m.session.Messages[len(m.session.Messages)-1].Role == "assistant" {
			m.session.Messages[len(m.session.Messages)-1].Content += errMsg
		} else {
			m.session.Messages = append(m.session.Messages, domain.ChatMessage{
				Role:      "assistant",
				Content:   errMsg,
				Timestamp: time.Now().UTC(),
			})
		}
		if m.novelDir != "" {
			_ = m.sessionRepo.Save(m.session)
		}
		m.renderHistory()
		m.viewport.GotoBottom()

	case messages.SelectSessionMsg:
		m.LoadSession(msg.SessionID)
		m.sessionPickerOpen = false

	case messages.CreateSessionMsg:
		m.NewSession()
		m.sessionPickerOpen = false

	case messages.SetEffortLevelMsg:
		m.SetEffortLevel(msg.EffortLevel)

	case messages.BrainActivityMsg:
		m.brainNotification = msg.Event.Description
		m.renderHistory()
		m.viewport.GotoBottom()

	case tea.KeyMsg:
		if !m.Focused {
			return m, nil
		}

		if m.sessionPickerOpen {
			switch msg.String() {
			case "up", "k":
				if m.sessionCursor > 0 {
					m.sessionCursor--
				}
				return m, nil
			case "down", "j":
				if m.sessionCursor < len(m.sessions)-1 {
					m.sessionCursor++
				}
				return m, nil
			case "enter":
				if len(m.sessions) > 0 && m.sessionCursor >= 0 && m.sessionCursor < len(m.sessions) {
					targetID := m.sessions[m.sessionCursor].ID
					m.LoadSession(targetID)
				}
				m.sessionPickerOpen = false
				return m, nil
			case "ctrl+d":
				if len(m.sessions) > 0 && m.sessionCursor >= 0 && m.sessionCursor < len(m.sessions) {
					toDelete := m.sessions[m.sessionCursor].ID
					_ = m.sessionRepo.Delete(m.novelDir, toDelete)
					m.RefreshSessions()
					if m.session.ID == toDelete {
						if len(m.sessions) > 0 {
							m.LoadSession(m.sessions[0].ID)
						} else {
							m.NewSession()
						}
					}
					if m.sessionCursor >= len(m.sessions) && len(m.sessions) > 0 {
						m.sessionCursor = len(m.sessions) - 1
					}
				}
				return m, nil
			case "esc", "ctrl+s":
				m.sessionPickerOpen = false
				return m, nil
			}
			return m, nil
		}

		// Handle normal chat shortcuts
		switch msg.String() {
		case "ctrl+e":
			m.CycleEffortLevel()
			return m, nil
		case "ctrl+s":
			m.RefreshSessions()
			m.sessionPickerOpen = true
			m.sessionCursor = 0
			for i, s := range m.sessions {
				if s.ID == m.session.ID {
					m.sessionCursor = i
					break
				}
			}
			return m, nil
		case "ctrl+n":
			m.NewSession()
			return m, nil
		case "enter":
			val := strings.TrimSpace(m.textarea.Value())
			if val != "" && !m.isGenerating {
				// Add user message
				userMsg := domain.ChatMessage{
					Role:      "user",
					Content:   val,
					Timestamp: time.Now().UTC(),
				}
				m.session.Messages = append(m.session.Messages, userMsg)
				m.textarea.Reset()
				m.isGenerating = true

				// Prepare empty assistant message for streaming
				m.session.Messages = append(m.session.Messages, domain.ChatMessage{
					Role:      "assistant",
					Content:   "",
					Timestamp: time.Now().UTC(),
				})

				if m.novelDir != "" {
					_ = m.sessionRepo.Save(m.session)
				}

				m.renderHistory()
				m.viewport.GotoBottom()

				return m, func() tea.Msg {
					return messages.SendChatMessageMsg{
						Content:     val,
						EffortLevel: m.effortLevel,
					}
				}
			}
			return m, nil
		}

		var taCmd tea.Cmd
		m.textarea, taCmd = m.textarea.Update(msg)
		cmds = append(cmds, taCmd)

		var vpCmd tea.Cmd
		m.viewport.Model, vpCmd = m.viewport.Model.Update(msg)
		cmds = append(cmds, vpCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *ChatDrawerModel) renderHistory() {
	if len(m.session.Messages) == 0 {
		baseText := "Inicia una conversación con tu asistente de escritura...\n\nPresiona Enter para enviar."
		if m.brainNotification != "" {
			brainStyle := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Highlight).Italic(true)
			baseText += "\n\n" + brainStyle.Render(m.brainNotification)
		}
		m.viewport.SetContent(m.styles.ListSubtitle.Render(baseText))
		return
	}

	userHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.CurrentTheme.Secondary)
	asstHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.CurrentTheme.Highlight)
	timeStyle := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Muted)
	dividerStyle := lipgloss.NewStyle().Foreground(theme.CurrentTheme.BorderBlurred)

	var sb strings.Builder
	for i, msg := range m.session.Messages {
		if i > 0 {
			sb.WriteString("\n" + dividerStyle.Render(strings.Repeat("─", m.viewport.Width)) + "\n\n")
		}

		tStr := ""
		if !msg.Timestamp.IsZero() {
			tStr = " " + timeStyle.Render(msg.Timestamp.Format("15:04"))
		}

		if msg.Role == "user" {
			sb.WriteString(userHeaderStyle.Render("👤 Autor") + tStr + "\n")
		} else {
			sb.WriteString(asstHeaderStyle.Render("🤖 Asistente") + tStr + "\n")
		}

		content := msg.Content
		if msg.Role == "assistant" && content == "" && m.isGenerating {
			content = "Pensando..."
		}
		sb.WriteString(content + "\n")
	}

	if m.brainNotification != "" {
		brainStyle := lipgloss.NewStyle().
			Foreground(theme.CurrentTheme.Highlight).
			Italic(true)
		sb.WriteString("\n" + brainStyle.Render(m.brainNotification) + "\n")
	}

	m.viewport.SetContent(sb.String())
}

// View renders the Chat Drawer UI.
func (m ChatDrawerModel) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	// 1. Header
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.CurrentTheme.Foreground)
	headerTitle := "🤖 Asistente IA"
	if m.session.Title != "" && m.session.Title != "Nueva Conversación" {
		runes := []rune(m.session.Title)
		if len(runes) > 15 {
			headerTitle = "🤖 " + string(runes[:12]) + ".."
		} else {
			headerTitle = "🤖 " + m.session.Title
		}
	}

	var effortBadge string
	switch m.effortLevel {
	case domain.EffortLow:
		effortBadge = lipgloss.NewStyle().Bold(true).Foreground(theme.CurrentTheme.Success).Render("[Bajo]")
	case domain.EffortHigh:
		effortBadge = lipgloss.NewStyle().Bold(true).Foreground(theme.CurrentTheme.Warning).Render("[Alto]")
	case domain.EffortMedium:
		fallthrough
	default:
		effortBadge = lipgloss.NewStyle().Bold(true).Foreground(theme.CurrentTheme.Accent).Render("[Medio]")
	}

	var statusIndicator string
	if m.isGenerating {
		statusIndicator = lipgloss.NewStyle().Foreground(theme.CurrentTheme.Highlight).Render(" ⚡ escribiendo...")
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Center,
		titleStyle.Render(headerTitle),
		" ",
		effortBadge,
		statusIndicator,
	)

	// 2. Middle View (Session Picker OR Viewport)
	var middleView string
	if m.sessionPickerOpen {
		middleView = m.renderSessionPicker()
	} else {
		middleView = m.viewport.View()
	}

	// 3. Separator
	sep := lipgloss.NewStyle().Foreground(theme.CurrentTheme.BorderBlurred).Render(strings.Repeat("─", m.width-4))

	// 4. Textarea Input
	inputView := m.textarea.View()

	// 5. Help Footer
	footerText := "[Enter] Enviar  [Ctrl+E] Nivel  [Ctrl+S] Sesiones  [Ctrl+N] Nueva"
	if m.sessionPickerOpen {
		footerText = "[Enter] Elegir  [Ctrl+D] Borrar  [Esc] Volver"
	}
	footer := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Muted).Render(footerText)

	// Combine all vertically
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		sep,
		middleView,
		sep,
		inputView,
		footer,
	)

	panelStyle := m.styles.BlurredPanel
	if m.Focused {
		panelStyle = m.styles.FocusedPanel
	}

	return panelStyle.
		Width(m.width - 2).
		Height(m.height - 2).
		Render(content)
}

func (m ChatDrawerModel) renderSessionPicker() string {
	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.CurrentTheme.Accent).Render("📁 Sesiones Guardadas") + "\n\n")

	if len(m.sessions) == 0 {
		sb.WriteString(m.styles.ListSubtitle.Render("No hay sesiones previas.\nPresiona [Esc] para volver."))
		return sb.String()
	}

	for i, s := range m.sessions {
		cursor := "  "
		itemStyle := m.styles.ListItem
		if i == m.sessionCursor {
			cursor = "▶ "
			itemStyle = m.styles.ListItemActive
		}

		title := s.Title
		if title == "" {
			title = "Sin título"
		}
		runes := []rune(title)
		if len(runes) > 22 {
			title = string(runes[:20]) + ".."
		}

		timeStr := s.UpdatedAt.Format("02/01 15:04")
		sb.WriteString(fmt.Sprintf("%s%s %s\n", cursor, itemStyle.Render(title), lipgloss.NewStyle().Foreground(theme.CurrentTheme.Muted).Render("("+timeStr+")")))
	}

	return sb.String()
}
