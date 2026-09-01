package components

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

const AsciiBanner = `
 ███╗   ██╗ ██████╗ ██╗   ██╗███████╗██╗         ████████╗██╗   ██╗██╗
 ████╗  ██║██╔═══██╗██║   ██║██╔════╝██║         ╚══██╔══╝██║   ██║██║
 ██╔██╗ ██║██║   ██║██║   ██║█████╗  ██║  █████╗    ██║   ██║   ██║██║
 ██║╚██╗██║██║   ██║╚██╗ ██╔╝██╔══╝  ██║  ╚════╝    ██║   ██║   ██║██║
 ██║ ╚████║╚██████╔╝ ╚████╔╝ ███████╗███████╗       ██║   ╚██████╔╝██║
 ╚═╝  ╚═══╝ ╚═════╝   ╚═══╝  ╚══════╝╚══════╝       ╚═╝    ╚═════╝ ╚═╝`

// LauncherKeyMap defines single-key navigation and actions in the launcher.
type LauncherKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Select   key.Binding
	Continue key.Binding
	New      key.Binding
	Open     key.Binding
	LLM      key.Binding
	RootDir  key.Binding
	Quit     key.Binding
}

// DefaultLauncherKeyMap returns standard launcher key bindings.
func DefaultLauncherKeyMap() LauncherKeyMap {
	return LauncherKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "arriba"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "abajo"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "abrir"),
		),
		Continue: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "continuar última"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "nueva novela"),
		),
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "abrir carpeta"),
		),
		LLM: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "configurar IA"),
		),
		RootDir: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "cambiar directorio"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "salir"),
		),
	}
}

// LauncherModel manages the Home Dashboard view.
type LauncherModel struct {
	Novels       []domain.NovelMetadata
	SelectedIndex int
	RootDir      string
	Notification string
	NotificationErr bool

	width  int
	height int
	styles theme.Styles
	keys   LauncherKeyMap
}

// NewLauncherModel creates a new LauncherModel.
func NewLauncherModel(styles theme.Styles) LauncherModel {
	return LauncherModel{
		Novels:        []domain.NovelMetadata{},
		SelectedIndex: 0,
		RootDir:       "~/Novelas",
		styles:        styles,
		keys:          DefaultLauncherKeyMap(),
	}
}

// SetNovels updates the list of discovered novels.
func (m *LauncherModel) SetNovels(novels []domain.NovelMetadata) {
	m.Novels = novels
	if m.SelectedIndex >= len(novels) && len(novels) > 0 {
		m.SelectedIndex = len(novels) - 1
	}
	if len(novels) == 0 {
		m.SelectedIndex = 0
	}
}

// SetRootDir updates the current root directory path.
func (m *LauncherModel) SetRootDir(dir string) {
	m.RootDir = dir
}

// SetSize updates layout dimensions.
func (m *LauncherModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Init initializes the launcher component.
func (m LauncherModel) Init() tea.Cmd {
	return nil
}

// Update handles user input on the launcher dashboard.
func (m LauncherModel) Update(msg tea.Msg) (LauncherModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.NovelListRefreshedMsg:
		m.SetNovels(msg.Novels)
		return m, nil

	case messages.NotificationMsg:
		m.Notification = msg.Message
		m.NotificationErr = msg.IsError
		return m, nil

	case tea.KeyMsg:
		m.Notification = "" // Clear notification on user action

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Up):
			if m.SelectedIndex > 0 {
				m.SelectedIndex--
			}

		case key.Matches(msg, m.keys.Down):
			if m.SelectedIndex < len(m.Novels)-1 {
				m.SelectedIndex++
			}

		case key.Matches(msg, m.keys.Select):
			if len(m.Novels) > 0 && m.SelectedIndex < len(m.Novels) {
				selected := m.Novels[m.SelectedIndex]
				return m, func() tea.Msg {
					return messages.OpenNovelMsg{Path: selected.AbsolutePath}
				}
			}

		case key.Matches(msg, m.keys.Continue):
			if len(m.Novels) > 0 {
				mostRecent := m.Novels[0]
				return m, func() tea.Msg {
					return messages.OpenNovelMsg{Path: mostRecent.AbsolutePath}
				}
			}
			m.Notification = "No se encontraron novelas recientes en el directorio"
			m.NotificationErr = false
			return m, nil

		case key.Matches(msg, m.keys.New):
			return m, func() tea.Msg {
				return messages.ShowModalMsg{
					Purpose: messages.ModalPurposeNewNovel,
					Title:   "Nueva Novela",
					Prompt:  "Título de la nueva novela:",
				}
			}

		case key.Matches(msg, m.keys.Open):
			return m, func() tea.Msg {
				return messages.ShowModalMsg{
					Purpose: messages.ModalPurposeOpenFolder,
					Title:   "Abrir Carpeta de Novela",
					Prompt:  "Ruta absoluta o relativa:",
				}
			}

		case key.Matches(msg, m.keys.LLM):
			return m, func() tea.Msg {
				return messages.ChangeViewMsg{View: messages.ViewStateLLMConfig}
			}

		case key.Matches(msg, m.keys.RootDir):
			return m, func() tea.Msg {
				return messages.ShowModalMsg{
					Purpose:      messages.ModalPurposeSetRootDir,
					Title:        "Configurar Directorio Raíz",
					Prompt:       "Directorio base para novelas:",
					InitialValue: m.RootDir,
				}
			}
		}
	}

	return m, nil
}

// View renders the Home Dashboard.
func (m LauncherModel) View() string {
	contentWidth := m.width
	if contentWidth <= 0 {
		contentWidth = 80
	}

	// 1. Banner
	bannerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.CurrentTheme.Highlight).
		Align(lipgloss.Center)
	banner := bannerStyle.Render(AsciiBanner)

	taglineStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme.Secondary).
		Italic(true).
		MarginBottom(1).
		Align(lipgloss.Center)
	tagline := taglineStyle.Render("Terminal Novel Writing Environment")

	headerBlock := lipgloss.JoinVertical(lipgloss.Center, banner, tagline)

	// 2. Notification bar
	var notifBlock string
	if m.Notification != "" {
		color := theme.CurrentTheme.Accent
		if m.NotificationErr {
			color = theme.CurrentTheme.Error
		}
		notifStyle := lipgloss.NewStyle().
			Foreground(color).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)
		notifBlock = notifStyle.Render("ℹ " + m.Notification)
	}

	// 3. Left / Top Section: Action Menu
	menuTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.CurrentTheme.Accent).
		MarginBottom(1)
	menuTitle := menuTitleStyle.Render("Acciones Rápidas")

	menuItems := []struct {
		key  string
		desc string
	}{
		{"c", "Continuar última novela"},
		{"n", "Nueva novela"},
		{"o", "Abrir otra carpeta"},
		{"l", "Configurar IA / LLM"},
		{"d", "Cambiar directorio raíz"},
		{"q", "Salir"},
	}

	var menuLines []string
	menuLines = append(menuLines, menuTitle)
	for _, item := range menuItems {
		kStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.CurrentTheme.Highlight).
			Background(theme.CurrentTheme.CardBg).
			Padding(0, 1)
		dStyle := lipgloss.NewStyle().
			Foreground(theme.CurrentTheme.Foreground).
			PaddingLeft(1)
		line := lipgloss.JoinHorizontal(lipgloss.Center, kStyle.Render("["+item.key+"]"), dStyle.Render(item.desc))
		menuLines = append(menuLines, line)
	}

	rootPathStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme.Muted).
		MarginTop(1)
	menuLines = append(menuLines, rootPathStyle.Render("Directorio raíz: "+m.RootDir))

	menuBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.CurrentTheme.BorderBlurred).
		Padding(1, 2).
		Width(38)
	menuBox := menuBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, menuLines...))

	// 4. Right / Main Section: Recent Novels List
	novelsTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.CurrentTheme.Accent).
		MarginBottom(1)
	novelsTitle := novelsTitleStyle.Render("Novelas Recientes")

	var novelListLines []string
	novelListLines = append(novelListLines, novelsTitle)

	if len(m.Novels) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(theme.CurrentTheme.Muted).
			Padding(1, 0)
		novelListLines = append(novelListLines, emptyStyle.Render("No hay novelas en este directorio.\nPresiona [n] para crear tu primera novela."))
	} else {
		for i, novel := range m.Novels {
			prefix := "  "
			if i == m.SelectedIndex {
				prefix = "▶ "
			}

			title := novel.Title
			if len(title) > 28 {
				title = title[:25] + "..."
			}

			itemHeader := fmt.Sprintf("%s%s", prefix, title)
			var renderedHeader string
			if i == m.SelectedIndex {
				renderedHeader = lipgloss.NewStyle().
					Bold(true).
					Foreground(theme.CurrentTheme.Background).
					Background(theme.CurrentTheme.BorderFocused).
					Padding(0, 1).
					Render(itemHeader)
			} else {
				renderedHeader = lipgloss.NewStyle().
					Foreground(theme.CurrentTheme.Foreground).
					Padding(0, 1).
					Render(itemHeader)
			}

			timeAgo := formatTimeAgo(novel.LastModified)
			metaText := fmt.Sprintf("    %d capítulos • modificado %s", novel.ChapterCount, timeAgo)
			renderedMeta := lipgloss.NewStyle().
				Foreground(theme.CurrentTheme.Muted).
				Render(metaText)

			novelListLines = append(novelListLines, renderedHeader, renderedMeta)
		}
	}

	novelsBoxWidth := 46
	if contentWidth > 90 {
		novelsBoxWidth = contentWidth - 44
	}
	novelsBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.CurrentTheme.BorderFocused).
		Padding(1, 2).
		Width(novelsBoxWidth)
	novelsBox := novelsBoxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, novelListLines...))

	// Join Action Menu and Novels List side by side
	columns := lipgloss.JoinHorizontal(lipgloss.Top, menuBox, " ", novelsBox)

	// Combine all elements vertically centered
	var bodyElements []string
	bodyElements = append(bodyElements, headerBlock)
	if notifBlock != "" {
		bodyElements = append(bodyElements, notifBlock)
	}
	bodyElements = append(bodyElements, columns)

	allContent := lipgloss.JoinVertical(lipgloss.Center, bodyElements...)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, allContent)
	}

	return allContent
}

func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "nunca"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "hace un momento"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "hace 1 minuto"
		}
		return fmt.Sprintf("hace %d minutos", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "hace 1 hora"
		}
		return fmt.Sprintf("hace %d horas", hours)
	case d < 30*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "ayer"
		}
		return fmt.Sprintf("hace %d días", days)
	default:
		return t.Format("02/01/2006")
	}
}
