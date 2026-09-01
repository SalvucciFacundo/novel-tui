package components

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// LLMConfigKeyMap defines keys in the LLM settings view.
type LLMConfigKeyMap struct {
	NextField key.Binding
	PrevField key.Binding
	Save      key.Binding
	Back      key.Binding
}

// DefaultLLMConfigKeyMap returns standard LLM config keys.
func DefaultLLMConfigKeyMap() LLMConfigKeyMap {
	return LLMConfigKeyMap{
		NextField: key.NewBinding(
			key.WithKeys("tab", "down"),
			key.WithHelp("tab/↓", "siguiente campo"),
		),
		PrevField: key.NewBinding(
			key.WithKeys("shift+tab", "up"),
			key.WithHelp("shift+tab/↑", "campo anterior"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "guardar"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "volver al inicio"),
		),
	}
}

// LLMConfigModel provides an interactive configuration form for LLM / AI parameters.
type LLMConfigModel struct {
	config domain.LLMConfig

	inputs       []textinput.Model
	focusedIndex int

	errorMsg string
	savedMsg string

	width  int
	height int
	styles theme.Styles
	keys   LLMConfigKeyMap
}

const (
	fieldProvider = iota
	fieldBaseURL
	fieldModel
	fieldTemperature
	fieldFantasyPrompt
	fieldSciFiPrompt
	fieldMysteryPrompt
	fieldRomancePrompt
	fieldCount
)

// NewLLMConfigModel creates a new LLMConfigModel.
func NewLLMConfigModel(styles theme.Styles) LLMConfigModel {
	inputs := make([]textinput.Model, fieldCount)

	labels := []string{
		"Proveedor (Provider)",
		"URL Base (Base URL)",
		"Nombre del Modelo (Model)",
		"Temperatura (0.0 - 2.0)",
		"Prompt Fantasía",
		"Prompt Sci-Fi",
		"Prompt Misterio",
		"Prompt Romance",
	}

	for i := range inputs {
		ti := textinput.New()
		ti.Prompt = "❯ "
		ti.CharLimit = 300
		ti.Width = 60
		ti.Placeholder = labels[i]
		inputs[i] = ti
	}

	m := LLMConfigModel{
		config:       domain.DefaultLLMConfig(),
		inputs:       inputs,
		focusedIndex: 0,
		styles:       styles,
		keys:         DefaultLLMConfigKeyMap(),
	}
	m.populateFields()
	return m
}

// SetConfig updates the LLM configuration and populates input fields.
func (m *LLMConfigModel) SetConfig(cfg domain.LLMConfig) {
	m.config = cfg
	m.populateFields()
}

func (m *LLMConfigModel) populateFields() {
	if len(m.inputs) < fieldCount {
		return
	}

	m.inputs[fieldProvider].SetValue(m.config.Provider)
	m.inputs[fieldBaseURL].SetValue(m.config.BaseURL)
	m.inputs[fieldModel].SetValue(m.config.Model)
	m.inputs[fieldTemperature].SetValue(fmt.Sprintf("%.2f", m.config.Temperature))

	if m.config.GenrePrompts == nil {
		m.config.GenrePrompts = domain.DefaultLLMConfig().GenrePrompts
	}

	m.inputs[fieldFantasyPrompt].SetValue(m.config.GenrePrompts["Fantasy"])
	m.inputs[fieldSciFiPrompt].SetValue(m.config.GenrePrompts["Sci-Fi"])
	m.inputs[fieldMysteryPrompt].SetValue(m.config.GenrePrompts["Mystery"])
	m.inputs[fieldRomancePrompt].SetValue(m.config.GenrePrompts["Romance"])

	m.focusCurrentField()
}

func (m *LLMConfigModel) focusCurrentField() {
	for i := range m.inputs {
		if i == m.focusedIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

// SetSize updates dimensions.
func (m *LLMConfigModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Init initializes the model.
func (m LLMConfigModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles interactions in the LLM settings form.
func (m LLMConfigModel) Update(msg tea.Msg) (LLMConfigModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.savedMsg = ""

		switch {
		case key.Matches(msg, m.keys.Back):
			m.errorMsg = ""
			m.savedMsg = ""
			return m, func() tea.Msg {
				return messages.ChangeViewMsg{View: messages.ViewStateLauncher}
			}

		case key.Matches(msg, m.keys.Save):
			return m, m.saveCmd()

		case key.Matches(msg, m.keys.NextField):
			m.focusedIndex = (m.focusedIndex + 1) % fieldCount
			m.focusCurrentField()
			return m, nil

		case key.Matches(msg, m.keys.PrevField):
			m.focusedIndex--
			if m.focusedIndex < 0 {
				m.focusedIndex = fieldCount - 1
			}
			m.focusCurrentField()
			return m, nil
		}
	}

	// Forward key to currently focused input
	if m.focusedIndex >= 0 && m.focusedIndex < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focusedIndex], cmd = m.inputs[m.focusedIndex].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *LLMConfigModel) saveCmd() tea.Cmd {
	provider := strings.TrimSpace(m.inputs[fieldProvider].Value())
	if provider == "" {
		m.errorMsg = "El proveedor no puede estar vacío"
		return nil
	}

	baseURL := strings.TrimSpace(m.inputs[fieldBaseURL].Value())
	if baseURL == "" {
		m.errorMsg = "La URL base no puede estar vacía"
		return nil
	}

	model := strings.TrimSpace(m.inputs[fieldModel].Value())
	if model == "" {
		m.errorMsg = "El modelo no puede estar vacío"
		return nil
	}

	tempStr := strings.TrimSpace(m.inputs[fieldTemperature].Value())
	temp, err := strconv.ParseFloat(tempStr, 64)
	if err != nil || temp < 0.0 || temp > 2.0 {
		m.errorMsg = "La temperatura debe ser un número decimal entre 0.0 y 2.0"
		return nil
	}

	genrePrompts := map[string]string{
		"Fantasy": strings.TrimSpace(m.inputs[fieldFantasyPrompt].Value()),
		"Sci-Fi":  strings.TrimSpace(m.inputs[fieldSciFiPrompt].Value()),
		"Mystery": strings.TrimSpace(m.inputs[fieldMysteryPrompt].Value()),
		"Romance": strings.TrimSpace(m.inputs[fieldRomancePrompt].Value()),
	}

	updatedConfig := domain.LLMConfig{
		Provider:     provider,
		BaseURL:      baseURL,
		Model:        model,
		Temperature:  temp,
		GenrePrompts: genrePrompts,
	}

	m.config = updatedConfig
	m.errorMsg = ""
	m.savedMsg = "✓ Configuración guardada exitosamente"

	return func() tea.Msg {
		return messages.SaveLLMConfigMsg{Config: updatedConfig}
	}
}

// View renders the LLM configuration form.
func (m LLMConfigModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.CurrentTheme.Highlight).
		MarginBottom(1)
	title := titleStyle.Render("⚙ Configuración de Inteligencia Artificial (LLM)")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme.Secondary).
		MarginBottom(1)
	subtitle := subtitleStyle.Render("Ajusta los parámetros de conexión con Ollama y las plantillas de prompts por género.")

	labels := []string{
		"Proveedor:",
		"URL Base:",
		"Modelo:",
		"Temperatura (0.0 - 2.0):",
		"Prompt Fantasía:",
		"Prompt Ciencia Ficción:",
		"Prompt Misterio:",
		"Prompt Romance:",
	}

	var formLines []string
	formLines = append(formLines, title, subtitle)

	for i := range m.inputs {
		lblStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.CurrentTheme.Accent)
		if i == m.focusedIndex {
			lblStyle = lblStyle.Foreground(theme.CurrentTheme.Highlight)
		}
		lbl := lblStyle.Render(labels[i])

		inputBox := m.inputs[i].View()
		formLines = append(formLines, lbl, inputBox)
	}

	if m.errorMsg != "" {
		errStyle := lipgloss.NewStyle().
			Foreground(theme.CurrentTheme.Error).
			Bold(true).
			MarginTop(1)
		formLines = append(formLines, errStyle.Render("⚠ "+m.errorMsg))
	}

	if m.savedMsg != "" {
		savedStyle := lipgloss.NewStyle().
			Foreground(theme.CurrentTheme.Success).
			Bold(true).
			MarginTop(1)
		formLines = append(formLines, savedStyle.Render(m.savedMsg))
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme.Muted).
		MarginTop(1)
	help := helpStyle.Render("[Tab/↓] Siguiente   [Shift+Tab/↑] Anterior   [Ctrl+S] Guardar   [Esc] Volver")
	formLines = append(formLines, help)

	cardWidth := 74
	if m.width > 0 && cardWidth > m.width-4 {
		cardWidth = m.width - 4
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.CurrentTheme.BorderFocused).
		Background(theme.CurrentTheme.CardBg).
		Padding(1, 2).
		Width(cardWidth)

	card := cardStyle.Render(lipgloss.JoinVertical(lipgloss.Left, formLines...))

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, card)
	}

	return card
}
