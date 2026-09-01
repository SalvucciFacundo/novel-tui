package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/repository"
	"github.com/SalvucciFacundo/novel-tui/internal/service"
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
	Quit       key.Binding
	NextTab    key.Binding
	PrevTab    key.Binding
	Save       key.Binding
	NewChapter key.Binding
	Launcher   key.Binding
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
		NewChapter: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "nuevo capítulo"),
		),
		Launcher: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "volver al launcher"),
		),
	}
}

// RootModel coordinates the top-level application state and active view.
type RootModel struct {
	viewState messages.ViewState

	configRepo   repository.ConfigRepository
	workspaceMgr service.WorkspaceManager
	config       *domain.AppConfig

	activeNovelPath string
	chapterRepo     domain.ChapterRepository
	characterRepo   domain.CharacterRepository

	launcher  components.LauncherModel
	llmConfig components.LLMConfigModel
	modal     components.ModalModel
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

// NewRootModel constructs the RootModel with default repositories for backward compatibility and testing.
func NewRootModel(
	chapterRepo domain.ChapterRepository,
	characterRepo domain.CharacterRepository,
) RootModel {
	styles := theme.DefaultStyles
	cfgRepo := repository.NewFileConfigRepository("")
	cfg, _ := cfgRepo.Load()
	if cfg == nil {
		cfg = domain.DefaultAppConfig()
	}
	wsMgr := service.NewWorkspaceManager()

	m := RootModel{
		viewState:     messages.ViewStateEditor,
		configRepo:    cfgRepo,
		workspaceMgr:  wsMgr,
		config:        cfg,
		chapterRepo:   chapterRepo,
		characterRepo: characterRepo,
		launcher:      components.NewLauncherModel(styles),
		llmConfig:     components.NewLLMConfigModel(styles),
		modal:         components.NewModalModel(styles),
		sidebar:       components.NewSidebarModel(chapterRepo, characterRepo, styles),
		editor:        components.NewEditorModel(styles),
		statusbar:     components.NewStatusBarModel(styles),
		activeFocus:   messages.FocusSidebar,
		styles:        styles,
		keys:          DefaultRootKeyMap(),
	}
	m.launcher.SetRootDir(cfg.RootDir)
	m.llmConfig.SetConfig(cfg.LLM)
	return m
}

// NewRootModelWithConfig creates a fully initialized RootModel with custom configuration.
func NewRootModelWithConfig(
	cfgRepo repository.ConfigRepository,
	wsMgr service.WorkspaceManager,
	initialView messages.ViewState,
	initialDir string,
) RootModel {
	styles := theme.DefaultStyles
	cfg, err := cfgRepo.Load()
	if err != nil || cfg == nil {
		cfg = domain.DefaultAppConfig()
	}

	var chapRepo domain.ChapterRepository
	var charRepo domain.CharacterRepository

	if initialDir != "" {
		initialDir = repository.ExpandHome(initialDir)
		chapRepo, _ = repository.NewFileChapterRepository(initialDir)
		charRepo = repository.NewFileCharacterRepository(initialDir)
	}

	m := RootModel{
		viewState:       initialView,
		configRepo:      cfgRepo,
		workspaceMgr:    wsMgr,
		config:          cfg,
		activeNovelPath: initialDir,
		chapterRepo:     chapRepo,
		characterRepo:   charRepo,
		launcher:        components.NewLauncherModel(styles),
		llmConfig:       components.NewLLMConfigModel(styles),
		modal:           components.NewModalModel(styles),
		sidebar:         components.NewSidebarModel(chapRepo, charRepo, styles),
		editor:          components.NewEditorModel(styles),
		statusbar:       components.NewStatusBarModel(styles),
		activeFocus:     messages.FocusSidebar,
		styles:          styles,
		keys:            DefaultRootKeyMap(),
	}

	m.launcher.SetRootDir(cfg.RootDir)
	m.llmConfig.SetConfig(cfg.LLM)
	return m
}

// Init initializes the active components.
func (m RootModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds,
		m.modal.Init(),
		m.launcher.Init(),
		m.llmConfig.Init(),
		m.sidebar.Init(),
		m.editor.Init(),
		m.statusbar.Init(),
		func() tea.Msg { return messages.FocusMsg{Target: m.activeFocus} },
	)

	// Scan recent novels on startup
	if m.config != nil && m.config.RootDir != "" {
		rootDir := m.config.RootDir
		mgr := m.workspaceMgr
		cmds = append(cmds, func() tea.Msg {
			novels, err := mgr.ListRecentNovels(rootDir)
			if err == nil {
				return messages.NovelListRefreshedMsg{Novels: novels}
			}
			return nil
		})
	}

	return tea.Batch(cmds...)
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

	case messages.ChangeViewMsg:
		m.viewState = msg.View
		if msg.View == messages.ViewStateLLMConfig && m.config != nil {
			m.llmConfig.SetConfig(m.config.LLM)
		}
		if msg.View == messages.ViewStateLauncher && m.config != nil {
			m.launcher.SetRootDir(m.config.RootDir)
			rootDir := m.config.RootDir
			mgr := m.workspaceMgr
			cmds = append(cmds, func() tea.Msg {
				novels, err := mgr.ListRecentNovels(rootDir)
				if err == nil {
					return messages.NovelListRefreshedMsg{Novels: novels}
				}
				return nil
			})
		}
		return m, tea.Batch(cmds...)

	case messages.SaveLLMConfigMsg:
		if m.config != nil {
			m.config.LLM = msg.Config
			_ = m.configRepo.Save(m.config)
		}
		return m, nil

	case messages.SetRootDirMsg:
		if m.config != nil {
			newPath := repository.ExpandHome(msg.Path)
			m.config.RootDir = newPath
			_ = m.configRepo.Save(m.config)
			m.launcher.SetRootDir(newPath)
			mgr := m.workspaceMgr
			cmds = append(cmds, func() tea.Msg {
				novels, err := mgr.ListRecentNovels(newPath)
				if err == nil {
					return messages.NovelListRefreshedMsg{Novels: novels}
				}
				return nil
			})
		}
		return m, tea.Batch(cmds...)

	case messages.CreateNovelMsg:
		if m.config == nil {
			return m, nil
		}
		meta, err := m.workspaceMgr.CreateNovel(m.config.RootDir, msg.Title)
		if err != nil {
			// Show error in modal
			m.modal.Show(messages.ModalPurposeNewNovel, "Nueva Novela", "Título de la novela:", msg.Title)
			m.modal.SetError(err.Error())
			return m, nil
		}

		// Update recent novels list in config
		m.updateRecentNovels(meta.AbsolutePath)
		return m, func() tea.Msg {
			return messages.OpenNovelMsg{Path: meta.AbsolutePath}
		}

	case messages.OpenNovelMsg:
		targetPath := repository.ExpandHome(msg.Path)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			m.launcher.Notification = fmt.Sprintf("La ruta %s no existe", targetPath)
			m.launcher.NotificationErr = true
			return m, nil
		}

		m.activeNovelPath = targetPath
		m.updateRecentNovels(targetPath)

		chapRepo, err := repository.NewFileChapterRepository(targetPath)
		if err != nil {
			m.launcher.Notification = fmt.Sprintf("Error al abrir capítulos: %v", err)
			m.launcher.NotificationErr = true
			return m, nil
		}
		charRepo := repository.NewFileCharacterRepository(targetPath)

		m.chapterRepo = chapRepo
		m.characterRepo = charRepo

		sCmd := m.sidebar.SetRepositories(chapRepo, charRepo)
		cmds = append(cmds, sCmd)

		// Switch to editor view
		m.viewState = messages.ViewStateEditor
		m.activeFocus = messages.FocusSidebar

		// Select first chapter if available
		chapters, _ := chapRepo.ListAll()
		if len(chapters) > 0 {
			firstChap := chapters[0]
			cmds = append(cmds, func() tea.Msg {
				return messages.ChapterSelectedMsg{Chapter: firstChap}
			})
		}

		return m, tea.Batch(cmds...)

	case messages.CreateChapterMsg:
		if m.activeNovelPath == "" && m.chapterRepo == nil {
			return m, nil
		}

		var newChap domain.Chapter
		var err error

		if m.workspaceMgr != nil && m.activeNovelPath != "" {
			var chapPath string
			chapPath, err = m.workspaceMgr.CreateChapter(m.activeNovelPath, msg.Title)
			if err == nil {
				// Reload repository chapters to find the newly created chapter
				chapters, _ := m.chapterRepo.ListAll()
				for _, ch := range chapters {
					if filepath.Base(ch.FilePath) == filepath.Base(chapPath) {
						newChap = ch
						break
					}
				}
				if newChap.ID == "" && len(chapters) > 0 {
					newChap = chapters[len(chapters)-1]
				}
			}
		} else {
			newChap, err = m.chapterRepo.Create(msg.Title)
		}

		if err != nil {
			m.modal.Show(messages.ModalPurposeNewChapter, "Nuevo Capítulo", "Título del capítulo:", msg.Title)
			m.modal.SetError(err.Error())
			return m, nil
		}

		// Notify sidebar and editor
		cmds = append(cmds,
			func() tea.Msg { return messages.ChapterCreatedMsg{Chapter: newChap} },
			func() tea.Msg { return messages.ChapterSelectedMsg{Chapter: newChap} },
		)
		return m, tea.Batch(cmds...)

	case messages.ShowModalMsg:
		var mCmd tea.Cmd
		m.modal, mCmd = m.modal.Update(msg)
		return m, mCmd

	case messages.HideModalMsg:
		var mCmd tea.Cmd
		m.modal, mCmd = m.modal.Update(msg)
		return m, mCmd

	case messages.NovelListRefreshedMsg:
		m.launcher.SetNovels(msg.Novels)
		return m, nil

	case messages.SaveRequestedMsg:
		chapterID := msg.ChapterID
		content := msg.Content
		repo := m.chapterRepo
		if repo != nil {
			return m, func() tea.Msg {
				err := repo.SaveContent(chapterID, content)
				return messages.SaveCompletedMsg{
					ChapterID: chapterID,
					Success:   err == nil,
					Error:     err,
				}
			}
		}
		return m, nil

	case messages.ChapterSelectedMsg:
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

		m.activeFocus = messages.FocusEditor
		cmds = append(cmds, func() tea.Msg { return messages.FocusMsg{Target: messages.FocusEditor} })
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		// 1. Modal dialog intercepts keys when open
		if m.modal.Active {
			var mCmd tea.Cmd
			m.modal, mCmd = m.modal.Update(msg)
			return m, mCmd
		}

		// 2. Global Quit
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

		// 3. View-specific key routing
		switch m.viewState {
		case messages.ViewStateLauncher:
			var lCmd tea.Cmd
			m.launcher, lCmd = m.launcher.Update(msg)
			return m, lCmd

		case messages.ViewStateLLMConfig:
			var cfgCmd tea.Cmd
			m.llmConfig, cfgCmd = m.llmConfig.Update(msg)
			return m, cfgCmd

		case messages.ViewStateEditor:
			switch {
			case key.Matches(msg, m.keys.Launcher):
				return m, func() tea.Msg {
					return messages.ChangeViewMsg{View: messages.ViewStateLauncher}
				}

			case key.Matches(msg, m.keys.NewChapter):
				return m, func() tea.Msg {
					return messages.ShowModalMsg{
						Purpose: messages.ModalPurposeNewChapter,
						Title:   "Nuevo Capítulo",
						Prompt:  "Título del nuevo capítulo:",
					}
				}

			case key.Matches(msg, m.keys.NextTab), key.Matches(msg, m.keys.PrevTab):
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
			}

			// Forward keys to active focused sub-panel in editor
			var sCmd, eCmd, statusCmd tea.Cmd
			m.sidebar, sCmd = m.sidebar.Update(msg)
			m.editor, eCmd = m.editor.Update(msg)
			m.statusbar, statusCmd = m.statusbar.Update(msg)
			return m, tea.Batch(sCmd, eCmd, statusCmd)
		}
	}

	// Forward non-key messages to child components
	var lCmd, cfgCmd, mCmd, sCmd, eCmd, stCmd tea.Cmd
	m.launcher, lCmd = m.launcher.Update(msg)
	m.llmConfig, cfgCmd = m.llmConfig.Update(msg)
	m.modal, mCmd = m.modal.Update(msg)
	m.sidebar, sCmd = m.sidebar.Update(msg)
	m.editor, eCmd = m.editor.Update(msg)
	m.statusbar, stCmd = m.statusbar.Update(msg)

	cmds = append(cmds, lCmd, cfgCmd, mCmd, sCmd, eCmd, stCmd)
	return m, tea.Batch(cmds...)
}

func (m *RootModel) updateRecentNovels(novelPath string) {
	if m.config == nil {
		return
	}
	var updated []string
	updated = append(updated, novelPath)
	for _, p := range m.config.RecentNovels {
		if p != novelPath && strings.TrimSpace(p) != "" {
			updated = append(updated, p)
		}
	}
	if len(updated) > 10 {
		updated = updated[:10]
	}
	m.config.RecentNovels = updated
	_ = m.configRepo.Save(m.config)
}

func (m *RootModel) recalculateLayout() {
	if m.width < MinWidth || m.height < MinHeight {
		return
	}

	m.modal.SetSize(m.width, m.height)
	m.launcher.SetSize(m.width, m.height)
	m.llmConfig.SetSize(m.width, m.height)

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

// View renders the multi-view TUI with modal overlays.
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

	// 2. Base View rendering based on ViewState
	var baseView string
	switch m.viewState {
	case messages.ViewStateLauncher:
		baseView = m.launcher.View()

	case messages.ViewStateLLMConfig:
		baseView = m.llmConfig.View()

	case messages.ViewStateEditor:
		sidebarView := m.sidebar.View()
		editorView := m.editor.View()
		mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, editorView)
		statusView := m.statusbar.View()
		baseView = lipgloss.JoinVertical(lipgloss.Left, mainView, statusView)
	}

	// 3. If modal is active, overlay it
	if m.modal.Active {
		return m.modal.View()
	}

	return baseView
}
