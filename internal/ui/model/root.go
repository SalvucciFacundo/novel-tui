package model

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/repository"
	"github.com/SalvucciFacundo/novel-tui/internal/service"
	"github.com/SalvucciFacundo/novel-tui/internal/service/llm"
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
	Quit           key.Binding
	NextTab        key.Binding
	PrevTab        key.Binding
	Save           key.Binding
	NewChapter     key.Binding
	Launcher       key.Binding
	ToggleChat     key.Binding
	Search         key.Binding
	CommandPalette key.Binding
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
		ToggleChat: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "toggle chat drawer"),
		),
		Search: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "búsqueda global"),
		),
		CommandPalette: key.NewBinding(
			key.WithKeys("ctrl+p", "f1"),
			key.WithHelp("ctrl+p", "paleta de comandos"),
		),
	}
}

// RootModel coordinates the top-level application state and active view.
type RootModel struct {
	viewState messages.ViewState

	configRepo   repository.ConfigRepository
	workspaceMgr service.WorkspaceManager
	config       *domain.AppConfig

	activeNovelPath     string
	activeNovelSettings domain.NovelSettings
	chapterRepo         domain.ChapterRepository
	characterRepo   domain.CharacterRepository
	sessionRepo     domain.ChatSessionRepository
	brainRepo       domain.BrainRepository
	brainService    *service.BrainService
	searchService   *service.SearchService

	launcher       components.LauncherModel
	llmConfig      components.LLMConfigModel
	modal          components.ModalModel
	searchModal    components.SearchModalModel
	commandPalette components.CommandPaletteModel
	navbar         components.NavbarModel
	sidebar        components.SidebarModel
	editor         components.EditorModel
	statusbar      components.StatusBarModel
	chatDrawer     components.ChatDrawerModel

	showChatDrawer bool
	streamCancel   context.CancelFunc

	activeFocus messages.FocusState
	width       int
	height      int
	ready       bool

	styles theme.Styles
	keys   RootKeyMap
}

type streamChunkMsg struct {
	content string
	ch      <-chan domain.StreamChunk
}

type streamChunkWithDoneMsg struct {
	content string
}

func readStreamChunkCmd(ch <-chan domain.StreamChunk) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-ch
		if !ok {
			return messages.StreamFinishedMsg{}
		}
		if chunk.Err != nil {
			return messages.StreamErrorMsg{Err: chunk.Err}
		}
		if chunk.Done {
			if chunk.Content != "" {
				return streamChunkWithDoneMsg{content: chunk.Content}
			}
			return messages.StreamFinishedMsg{}
		}
		return streamChunkMsg{content: chunk.Content, ch: ch}
	}
}

func defaultRootCommands() []domain.CommandItem {
	return append(domain.DefaultCommands(), domain.CommandItem{
		ID:          "configure_genres",
		Title:       "Configurar Géneros y Clasificación de la Novela",
		Category:    "Configuración",
		Shortcut:    "g",
		Description: "Configurar géneros, tropos R-18 y clasificación de la novela",
	})
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
	sessionRepo := repository.NewFileChatSessionRepository()

	m := RootModel{
		viewState:     messages.ViewStateEditor,
		configRepo:    cfgRepo,
		workspaceMgr:  wsMgr,
		config:        cfg,
		chapterRepo:   chapterRepo,
		characterRepo: characterRepo,
		sessionRepo:   sessionRepo,
		launcher:       components.NewLauncherModel(styles),
		llmConfig:      components.NewLLMConfigModel(styles),
		modal:          components.NewModalModel(styles),
		searchModal:    components.NewSearchModalModel(styles),
		commandPalette: components.NewCommandPaletteModel(styles),
		navbar:         components.NewNavbarModel(styles),
		sidebar:        components.NewSidebarModel(chapterRepo, characterRepo, styles),
		editor:         components.NewEditorModel(styles),
		statusbar:      components.NewStatusBarModel(styles),
		chatDrawer:     components.NewChatDrawerModel(sessionRepo, styles),
		activeNovelSettings: domain.DefaultNovelSettings(),
		activeFocus:   messages.FocusSidebar,
		styles:        styles,
		keys:          DefaultRootKeyMap(),
	}
	m.commandPalette.SetCommands(defaultRootCommands())
	if chapterRepo != nil {
		m.searchService = service.NewSearchService(chapterRepo)
		m.searchModal.SearchFunc = func(query string, caseSensitive bool) ([]domain.SearchMatch, error) {
			if m.searchService != nil {
				return m.searchService.Search(query, caseSensitive)
			}
			return nil, nil
		}
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
	sessionRepo := repository.NewFileChatSessionRepository()

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
		sessionRepo:     sessionRepo,
		launcher:        components.NewLauncherModel(styles),
		llmConfig:       components.NewLLMConfigModel(styles),
		modal:           components.NewModalModel(styles),
		searchModal:     components.NewSearchModalModel(styles),
		commandPalette:  components.NewCommandPaletteModel(styles),
		navbar:          components.NewNavbarModel(styles),
		sidebar:         components.NewSidebarModel(chapRepo, charRepo, styles),
		editor:          components.NewEditorModel(styles),
		statusbar:       components.NewStatusBarModel(styles),
		chatDrawer:      components.NewChatDrawerModel(sessionRepo, styles),
		activeNovelSettings: domain.DefaultNovelSettings(),
		activeFocus:     messages.FocusSidebar,
		styles:          styles,
		keys:            DefaultRootKeyMap(),
	}
	m.commandPalette.SetCommands(defaultRootCommands())

	if initialDir != "" && wsMgr != nil {
		if settings, err := wsMgr.LoadNovelSettings(initialDir); err == nil {
			m.activeNovelSettings = settings
		}
	}

	if chapRepo != nil {
		m.searchService = service.NewSearchService(chapRepo)
		m.searchModal.SearchFunc = func(query string, caseSensitive bool) ([]domain.SearchMatch, error) {
			if m.searchService != nil {
				return m.searchService.Search(query, caseSensitive)
			}
			return nil, nil
		}
	}

	m.launcher.SetRootDir(cfg.RootDir)
	m.llmConfig.SetConfig(cfg.LLM)
	if initialDir != "" {
		m.chatDrawer.SetNovelDir(initialDir)
		m.sidebar.SetNovelPath(initialDir)
		m.navbar.SetNovelTitle(filepath.Base(initialDir))
	}
	return m
}

// Init initializes the active components.
func (m RootModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds,
		m.modal.Init(),
		m.searchModal.Init(),
		m.commandPalette.Init(),
		m.launcher.Init(),
		m.llmConfig.Init(),
		m.navbar.Init(),
		m.sidebar.Init(),
		m.editor.Init(),
		m.statusbar.Init(),
		m.chatDrawer.Init(),
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
			m.modal.Show(messages.ModalPurposeNewNovel, "Nueva Novela", "Título de la novela:", msg.Title)
			m.modal.SetError(err.Error())
			return m, nil
		}

		m.navbar.SetNovelTitle(meta.Title)
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

		if m.workspaceMgr != nil {
			if settings, err := m.workspaceMgr.LoadNovelSettings(targetPath); err == nil {
				m.activeNovelSettings = settings
			} else {
				m.activeNovelSettings = domain.DefaultNovelSettings()
			}
		} else {
			m.activeNovelSettings = domain.DefaultNovelSettings()
		}

		chapRepo, err := repository.NewFileChapterRepository(targetPath)
		if err != nil {
			m.launcher.Notification = fmt.Sprintf("Error al abrir capítulos: %v", err)
			m.launcher.NotificationErr = true
			return m, nil
		}
		charRepo := repository.NewFileCharacterRepository(targetPath)

		m.chapterRepo = chapRepo
		m.characterRepo = charRepo
		m.searchService = service.NewSearchService(chapRepo)
		m.searchModal.SearchFunc = func(query string, caseSensitive bool) ([]domain.SearchMatch, error) {
			if m.searchService != nil {
				return m.searchService.Search(query, caseSensitive)
			}
			return nil, nil
		}

		brainPath := filepath.Join(targetPath, ".novel", "brain.db")
		brainRepo, err := repository.NewSQLiteBrainRepository(brainPath)
		if err == nil {
			if m.brainRepo != nil {
				_ = m.brainRepo.Close()
			}
			m.brainRepo = brainRepo
			m.brainService = service.NewBrainService(brainRepo)
		}

		sCmd := m.sidebar.SetRepositories(chapRepo, charRepo)
		cmds = append(cmds, sCmd)
		if m.brainRepo != nil {
			cmds = append(cmds, m.sidebar.SetBrainRepository(m.brainRepo))
		}

		m.sidebar.SetNovelPath(targetPath)
		if m.navbar.NovelTitle == "" {
			m.navbar.SetNovelTitle(filepath.Base(targetPath))
		}
		m.chatDrawer.SetNovelDir(targetPath)

		m.viewState = messages.ViewStateEditor
		m.activeFocus = messages.FocusSidebar

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

		cmds = append(cmds,
			func() tea.Msg { return messages.ChapterCreatedMsg{Chapter: newChap} },
			func() tea.Msg { return messages.ChapterSelectedMsg{Chapter: newChap} },
		)
		return m, tea.Batch(cmds...)

	case messages.ShowModalMsg:
		var mCmd tea.Cmd
		m.modal, mCmd = m.modal.Update(msg)
		return m, mCmd

	case messages.OpenCommandPaletteMsg:
		var cpCmd tea.Cmd
		m.commandPalette, cpCmd = m.commandPalette.Update(msg)
		return m, cpCmd

	case messages.CloseCommandPaletteMsg:
		var cpCmd tea.Cmd
		m.commandPalette, cpCmd = m.commandPalette.Update(msg)
		return m, cpCmd

	case messages.ExecuteCommandMsg:
		switch msg.Command.ID {
		case "global_search":
			var smCmd tea.Cmd
			m.searchModal, smCmd = m.searchModal.Update(messages.OpenGlobalSearchMsg{})
			return m, smCmd
		case "save_chapter":
			return m, func() tea.Msg {
				return messages.SaveRequestedMsg{
					ChapterID: m.editor.ActiveChapter.ID,
					Content:   m.editor.ActiveChapter.Content,
				}
			}
		case "new_chapter":
			m.modal.Show(messages.ModalPurposeNewChapter, "Nuevo Capítulo", "Título del nuevo capítulo:", "")
			return m, nil
		case "toggle_ai":
			m.showChatDrawer = !m.showChatDrawer
			m.recalculateLayout()
			if m.showChatDrawer {
				m.activeFocus = messages.FocusChat
				m.sidebar.Focused = false
				m.editor.Focused = false
				m.chatDrawer.Focused = true
			} else {
				if m.activeFocus == messages.FocusChat {
					m.activeFocus = messages.FocusEditor
					m.editor.Focused = true
				}
				m.chatDrawer.Focused = false
			}
			focusMsg := messages.FocusMsg{Target: m.activeFocus}
			var sCmd, eCmd, cCmd tea.Cmd
			m.sidebar, sCmd = m.sidebar.Update(focusMsg)
			m.editor, eCmd = m.editor.Update(focusMsg)
			m.chatDrawer, cCmd = m.chatDrawer.Update(focusMsg)
			return m, tea.Batch(sCmd, eCmd, cCmd)
		case "go_launcher":
			return m, func() tea.Msg {
				return messages.ChangeViewMsg{View: messages.ViewStateLauncher}
			}
		case "tab_chapters":
			var sCmd tea.Cmd
			m.sidebar, sCmd = m.sidebar.Update(messages.SelectSidebarTabMsg{Tab: 0})
			return m, sCmd
		case "tab_characters":
			var sCmd tea.Cmd
			m.sidebar, sCmd = m.sidebar.Update(messages.SelectSidebarTabMsg{Tab: 1})
			return m, sCmd
		case "tab_notes":
			var sCmd tea.Cmd
			m.sidebar, sCmd = m.sidebar.Update(messages.SelectSidebarTabMsg{Tab: 2})
			return m, sCmd
		case "tab_brain":
			var sCmd tea.Cmd
			m.sidebar, sCmd = m.sidebar.Update(messages.SelectSidebarTabMsg{Tab: 3})
			return m, sCmd
		case "toggle_timeline":
			var sCmd tea.Cmd
			m.sidebar, sCmd = m.sidebar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
			return m, sCmd
		case "configure_genres":
			if m.activeNovelPath != "" {
				initialVal := ""
				if len(m.activeNovelSettings.Genres) > 0 {
					initialVal = strings.Join(m.activeNovelSettings.Genres, ", ")
				}
				m.modal.Show(messages.ModalPurposeConfigureGenres, "Configurar Géneros y Clasificación", "Géneros separados por coma (ej: isekai_harem_r18, yandere_obsession):", initialVal)
			}
			return m, nil
		case "llm_config":
			return m, func() tea.Msg {
				return messages.ChangeViewMsg{View: messages.ViewStateLLMConfig}
			}
		}
		return m, nil

	case messages.SubmitModalMsg:
		if msg.Purpose == messages.ModalPurposeConfigureGenres {
			if m.activeNovelPath != "" {
				raw := strings.TrimSpace(msg.Value)
				var genres []string
				if raw != "" {
					parts := strings.Split(raw, ",")
					for _, p := range parts {
						p = strings.TrimSpace(p)
						if p != "" {
							genres = append(genres, p)
						}
					}
				}
				m.activeNovelSettings.Genres = genres
				if m.workspaceMgr != nil {
					_ = m.workspaceMgr.SaveNovelSettings(m.activeNovelPath, m.activeNovelSettings)
				}
				return m, func() tea.Msg {
					return messages.NotificationMsg{
						Message: "Géneros actualizados para la novela",
						IsError: false,
					}
				}
			}
			return m, nil
		}
		return m, nil

	case messages.OpenGlobalSearchMsg:
		var smCmd tea.Cmd
		m.searchModal, smCmd = m.searchModal.Update(msg)
		return m, smCmd

	case messages.CloseGlobalSearchMsg:
		var smCmd tea.Cmd
		m.searchModal, smCmd = m.searchModal.Update(msg)
		return m, smCmd

	case messages.JumpToMatchMsg:
		targetMatch := msg.Match
		m.searchModal.Hide()

		var chap domain.Chapter
		if m.chapterRepo != nil {
			content, err := m.chapterRepo.LoadContent(targetMatch.ChapterID)
			if err == nil {
				chap = domain.Chapter{
					ID:        targetMatch.ChapterID,
					Title:     targetMatch.ChapterTitle,
					FilePath:  targetMatch.FilePath,
					Content:   content,
					WordCount: service.CalculateMetrics(content, false).WordCount,
				}
			}
		}
		if chap.ID == "" {
			chap = domain.Chapter{
				ID:       targetMatch.ChapterID,
				Title:    targetMatch.ChapterTitle,
				FilePath: targetMatch.FilePath,
			}
		}

		m.navbar.SetChapterTitle(chap.Title)
		for i, ch := range m.sidebar.Chapters {
			if ch.ID == chap.ID {
				m.sidebar.SelectedChapter = i
				break
			}
		}
		m.editor, _ = m.editor.Update(messages.ChapterSelectedMsg{Chapter: chap})
		m.statusbar, _ = m.statusbar.Update(messages.ChapterSelectedMsg{Chapter: chap})
		m.editor.SetCursorPosition(targetMatch.LineNumber, targetMatch.Column)

		m.activeFocus = messages.FocusEditor
		m.sidebar.Focused = false
		m.editor.Focused = true
		cmds = append(cmds, func() tea.Msg { return messages.FocusMsg{Target: messages.FocusEditor} })
		return m, tea.Batch(cmds...)

	case messages.GlobalReplaceMsg:
		if m.searchService != nil {
			res, err := m.searchService.ReplaceAll(msg.Query, msg.Replacement, msg.CaseSensitive)
			if err != nil {
				cmds = append(cmds, func() tea.Msg {
					return messages.NotificationMsg{
						Message: fmt.Sprintf("Error al reemplazar: %v", err),
						IsError: true,
					}
				})
			} else {
				if m.chapterRepo != nil && m.editor.ActiveChapter.ID != "" {
					content, err := m.chapterRepo.LoadContent(m.editor.ActiveChapter.ID)
					if err == nil {
						m.editor.ActiveChapter.Content = content
						m.editor, _ = m.editor.Update(messages.ChapterSelectedMsg{Chapter: m.editor.ActiveChapter})
					}
				}
				cmds = append(cmds,
					func() tea.Msg {
						return messages.NotificationMsg{
							Message: fmt.Sprintf("✓ Reemplazadas %d ocurrencias en %d capítulos", res.TotalReplaced, res.ChaptersAffected),
							IsError: false,
						}
					},
					func() tea.Msg { return messages.ReloadChaptersMsg{} },
					func() tea.Msg { return messages.GlobalReplaceCompletedMsg{Result: res} },
				)
			}
		}
		return m, tea.Batch(cmds...)

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
		var navCmd tea.Cmd
		m.navbar, navCmd = m.navbar.Update(msg)
		cmds = append(cmds, navCmd)

		var editorCmd tea.Cmd
		m.editor, editorCmd = m.editor.Update(msg)
		cmds = append(cmds, editorCmd)

		var statusCmd tea.Cmd
		m.statusbar, statusCmd = m.statusbar.Update(msg)
		cmds = append(cmds, statusCmd)

		return m, tea.Batch(cmds...)

	case messages.SelectSidebarTabMsg:
		var sCmd tea.Cmd
		m.sidebar, sCmd = m.sidebar.Update(msg)
		cmds = append(cmds, sCmd)
		return m, tea.Batch(cmds...)

	case messages.ChapterCreatedMsg:
		var sCmd tea.Cmd
		m.sidebar, sCmd = m.sidebar.Update(msg)
		cmds = append(cmds, sCmd)

		m.activeFocus = messages.FocusEditor
		cmds = append(cmds, func() tea.Msg { return messages.FocusMsg{Target: messages.FocusEditor} })
		return m, tea.Batch(cmds...)

	case messages.ToggleChatDrawerMsg:
		m.showChatDrawer = !m.showChatDrawer
		if m.showChatDrawer {
			m.activeFocus = messages.FocusChat
			if m.activeNovelPath != "" {
				m.chatDrawer.SetNovelDir(m.activeNovelPath)
			}
		} else {
			if m.activeFocus == messages.FocusChat {
				m.activeFocus = messages.FocusEditor
			}
		}
		m.recalculateLayout()
		cmds = append(cmds, func() tea.Msg { return messages.FocusMsg{Target: m.activeFocus} })
		return m, tea.Batch(cmds...)

	case messages.SendChatMessageMsg:
		if m.streamCancel != nil {
			m.streamCancel()
		}
		ctx, cancel := context.WithCancel(context.Background())
		m.streamCancel = cancel

		builder := llm.NewContextBuilder()
		genrePrompt := ""
		if m.config != nil && len(m.config.LLM.GenrePrompts) > 0 {
			for _, p := range m.config.LLM.GenrePrompts {
				genrePrompt = p
				break
			}
		}

		var brainFacts []domain.BrainFact
		if m.brainService != nil {
			queryText := msg.Content + " " + m.editor.ActiveChapter.Title
			facts, err := m.brainService.SearchRelevantFacts(context.Background(), queryText, 5)
			if err == nil {
				brainFacts = facts
			}
		}

		sysPrompt := builder.BuildContext(llm.ContextParams{
			NovelDir:           m.activeNovelPath,
			Genres:             m.activeNovelSettings.Genres,
			Rating:             m.activeNovelSettings.Rating,
			CustomPrompt:       m.activeNovelSettings.CustomPrompt,
			GenrePrompt:        genrePrompt,
			ActiveChapterTitle: m.editor.ActiveChapter.Title,
			ActiveChapterText:  m.editor.Value(),
			EffortLevel:        msg.EffortLevel,
			BrainFacts:         brainFacts,
		})

		if m.brainService != nil && m.brainRepo != nil {
			timelineEvents, err := m.brainRepo.ListTimelineEvents(context.Background())
			if err == nil && len(timelineEvents) > 0 {
				timelineBlock := m.brainService.FormatTimelineForPrompt(timelineEvents)
				if timelineBlock != "" {
					sysPrompt = strings.TrimSpace(sysPrompt) + "\n\n" + timelineBlock
				}
			}
		}

		activeSession := m.chatDrawer.ActiveSession()
		var chatMsgs []domain.ChatMessage
		chatMsgs = append(chatMsgs, domain.ChatMessage{
			Role:    "system",
			Content: sysPrompt,
		})

		for _, msgItem := range activeSession.Messages {
			if strings.TrimSpace(msgItem.Content) != "" {
				chatMsgs = append(chatMsgs, domain.ChatMessage{
					Role:    msgItem.Role,
					Content: msgItem.Content,
				})
			}
		}

		llmCfg := domain.DefaultLLMConfig()
		if m.config != nil {
			llmCfg = m.config.LLM
		}

		temp := llmCfg.Temperature
		switch msg.EffortLevel {
		case domain.EffortLow:
			temp = 0.4
		case domain.EffortHigh:
			temp = 0.5
		case domain.EffortMedium:
			temp = 0.7
		}

		chatReq := domain.ChatRequest{
			Model:       llmCfg.Model,
			Messages:    chatMsgs,
			Temperature: temp,
		}

		provider, err := llm.NewProvider(llmCfg)
		if err != nil {
			return m, func() tea.Msg {
				return messages.StreamErrorMsg{Err: err}
			}
		}

		return m, func() tea.Msg {
			ch, err := provider.StreamChat(ctx, chatReq)
			if err != nil {
				return messages.StreamErrorMsg{Err: err}
			}
			return streamChunkMsg{ch: ch}
		}

	case streamChunkMsg:
		var cCmd tea.Cmd
		if msg.content != "" {
			m.chatDrawer, cCmd = m.chatDrawer.Update(messages.TokenReceivedMsg{Content: msg.content})
			cmds = append(cmds, cCmd)
		}
		if msg.ch != nil {
			cmds = append(cmds, readStreamChunkCmd(msg.ch))
		}
		return m, tea.Batch(cmds...)

	case streamChunkWithDoneMsg:
		m.streamCancel = nil
		var cCmd1, cCmd2 tea.Cmd
		m.chatDrawer, cCmd1 = m.chatDrawer.Update(messages.TokenReceivedMsg{Content: msg.content})
		m.chatDrawer, cCmd2 = m.chatDrawer.Update(messages.StreamFinishedMsg{})
		return m, tea.Batch(cCmd1, cCmd2)

	case messages.StreamFinishedMsg:
		m.streamCancel = nil
		var cCmd tea.Cmd
		m.chatDrawer, cCmd = m.chatDrawer.Update(msg)
		cmds = append(cmds, cCmd)

		if m.brainService != nil && m.config != nil {
			lastMsg := m.chatDrawer.LastAssistantMessage()
			if strings.TrimSpace(lastMsg) != "" {
				chapTitle := m.editor.ActiveChapter.Title
				chapID := m.editor.ActiveChapter.ID
				brainSvc := m.brainService
				llmCfg := m.config.LLM

				cmds = append(cmds, func() tea.Msg {
					extractPrompt := brainSvc.BuildExtractionPrompt(lastMsg, chapTitle)
					provider, err := llm.NewProvider(llmCfg)
					if err != nil {
						return nil
					}
					ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
					defer cancel()

					req := domain.ChatRequest{
						Model:       llmCfg.Model,
						Temperature: 0.1,
						Messages: []domain.ChatMessage{
							{Role: "user", Content: extractPrompt},
						},
					}
					ch, err := provider.StreamChat(ctx, req)
					if err != nil {
						return nil
					}

					var fullRes strings.Builder
					for chunk := range ch {
						if chunk.Err != nil {
							break
						}
						fullRes.WriteString(chunk.Content)
					}

					facts, err := brainSvc.RecordAutoLearnedFacts(context.Background(), fullRes.String(), chapID)
					if err == nil && len(facts) > 0 {
						var concepts []string
						for _, f := range facts {
							concepts = append(concepts, f.Concept)
						}
						desc := fmt.Sprintf("🧠 [Brain] Memorizado: %d hecho(s) (%s)", len(facts), strings.Join(concepts, ", "))
						return messages.BrainActivityMsg{
							Event: domain.BrainActivityEvent{
								Type:        "saved",
								FactCount:   len(facts),
								Description: desc,
								Timestamp:   time.Now(),
							},
						}
					}
					return nil
				})
			}
		}
		return m, tea.Batch(cmds...)

	case messages.BrainActivityMsg:
		var sCmd, cCmd tea.Cmd
		m.sidebar, sCmd = m.sidebar.Update(msg)
		m.chatDrawer, cCmd = m.chatDrawer.Update(msg)
		return m, tea.Batch(sCmd, cCmd)

	case messages.StreamErrorMsg:
		m.streamCancel = nil
		var cCmd tea.Cmd
		m.chatDrawer, cCmd = m.chatDrawer.Update(msg)
		return m, cCmd

	case messages.FocusMsg:
		m.activeFocus = msg.Target
		var sCmd, eCmd, cCmd tea.Cmd
		m.sidebar, sCmd = m.sidebar.Update(msg)
		m.editor, eCmd = m.editor.Update(msg)
		m.chatDrawer, cCmd = m.chatDrawer.Update(msg)
		return m, tea.Batch(sCmd, eCmd, cCmd)

	case tea.MouseMsg:
		if m.commandPalette.Active {
			var cpCmd tea.Cmd
			m.commandPalette, cpCmd = m.commandPalette.Update(msg)
			return m, cpCmd
		}
		if m.searchModal.Active {
			var smCmd tea.Cmd
			m.searchModal, smCmd = m.searchModal.Update(msg)
			return m, smCmd
		}
		if m.modal.Active {
			var mCmd tea.Cmd
			m.modal, mCmd = m.modal.Update(msg)
			return m, mCmd
		}

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
			if msg.Y == 0 {
				var navCmd tea.Cmd
				m.navbar, navCmd = m.navbar.Update(msg)
				return m, navCmd
			}

			if m.height > 0 && msg.Y >= m.height-1 {
				var stCmd tea.Cmd
				m.statusbar, stCmd = m.statusbar.Update(msg)
				return m, stCmd
			}

			sidebarWidth := m.sidebar.Width
			if sidebarWidth <= 0 {
				sidebarWidth = SidebarDefaultWidth
			}
			editorWidth := m.editor.Width

			localMsg := msg
			localMsg.Y -= 1 // Account for top navbar height (1)

			if msg.X < sidebarWidth {
				if msg.Type == tea.MouseLeft {
					m.activeFocus = messages.FocusSidebar
					m.editor.Focused = false
					m.chatDrawer.Focused = false
				}
				var sCmd tea.Cmd
				m.sidebar, sCmd = m.sidebar.Update(localMsg)
				return m, sCmd
			} else if !m.showChatDrawer || msg.X < sidebarWidth+editorWidth {
				if msg.Type == tea.MouseLeft {
					m.activeFocus = messages.FocusEditor
					m.sidebar.Focused = false
					m.chatDrawer.Focused = false
				}
				localMsg.X -= sidebarWidth
				var eCmd tea.Cmd
				m.editor, eCmd = m.editor.Update(localMsg)
				return m, eCmd
			} else {
				if msg.Type == tea.MouseLeft {
					m.activeFocus = messages.FocusChat
					m.sidebar.Focused = false
					m.editor.Focused = false
					m.chatDrawer.Focused = true
				}
				localMsg.X -= (sidebarWidth + editorWidth)
				var cCmd tea.Cmd
				m.chatDrawer, cCmd = m.chatDrawer.Update(localMsg)
				return m, cCmd
			}
		}

	case tea.KeyMsg:
		if m.commandPalette.Active {
			var cpCmd tea.Cmd
			m.commandPalette, cpCmd = m.commandPalette.Update(msg)
			return m, cpCmd
		}
		if m.searchModal.Active {
			var smCmd tea.Cmd
			m.searchModal, smCmd = m.searchModal.Update(msg)
			return m, smCmd
		}
		if m.modal.Active {
			var mCmd tea.Cmd
			m.modal, mCmd = m.modal.Update(msg)
			return m, mCmd
		}

		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

		if key.Matches(msg, m.keys.CommandPalette) {
			var cpCmd tea.Cmd
			m.commandPalette, cpCmd = m.commandPalette.Update(messages.OpenCommandPaletteMsg{})
			return m, cpCmd
		}

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

			case key.Matches(msg, m.keys.ToggleChat):
				return m, func() tea.Msg {
					return messages.ToggleChatDrawerMsg{}
				}

			case key.Matches(msg, m.keys.Search):
				return m, func() tea.Msg {
					return messages.OpenGlobalSearchMsg{}
				}

			case key.Matches(msg, m.keys.NewChapter):
				if m.activeFocus != messages.FocusChat {
					return m, func() tea.Msg {
						return messages.ShowModalMsg{
							Purpose: messages.ModalPurposeNewChapter,
							Title:   "Nuevo Capítulo",
							Prompt:  "Título del nuevo capítulo:",
						}
					}
				}

			case key.Matches(msg, m.keys.NextTab):
				if m.showChatDrawer {
					switch m.activeFocus {
					case messages.FocusSidebar:
						m.activeFocus = messages.FocusEditor
					case messages.FocusEditor:
						m.activeFocus = messages.FocusChat
					case messages.FocusChat:
						fallthrough
					default:
						m.activeFocus = messages.FocusSidebar
					}
				} else {
					if m.activeFocus == messages.FocusSidebar {
						m.activeFocus = messages.FocusEditor
					} else {
						m.activeFocus = messages.FocusSidebar
					}
				}
				focusMsg := messages.FocusMsg{Target: m.activeFocus}
				var sCmd, eCmd, cCmd tea.Cmd
				m.sidebar, sCmd = m.sidebar.Update(focusMsg)
				m.editor, eCmd = m.editor.Update(focusMsg)
				m.chatDrawer, cCmd = m.chatDrawer.Update(focusMsg)
				return m, tea.Batch(sCmd, eCmd, cCmd)

			case key.Matches(msg, m.keys.PrevTab):
				if m.showChatDrawer {
					switch m.activeFocus {
					case messages.FocusSidebar:
						m.activeFocus = messages.FocusChat
					case messages.FocusChat:
						m.activeFocus = messages.FocusEditor
					case messages.FocusEditor:
						fallthrough
					default:
						m.activeFocus = messages.FocusSidebar
					}
				} else {
					if m.activeFocus == messages.FocusSidebar {
						m.activeFocus = messages.FocusEditor
					} else {
						m.activeFocus = messages.FocusSidebar
					}
				}
				focusMsg := messages.FocusMsg{Target: m.activeFocus}
				var sCmd, eCmd, cCmd tea.Cmd
				m.sidebar, sCmd = m.sidebar.Update(focusMsg)
				m.editor, eCmd = m.editor.Update(focusMsg)
				m.chatDrawer, cCmd = m.chatDrawer.Update(focusMsg)
				return m, tea.Batch(sCmd, eCmd, cCmd)

			case msg.String() == "esc" && m.activeFocus == messages.FocusChat && m.chatDrawer.IsGenerating():
				if m.streamCancel != nil {
					m.streamCancel()
					m.streamCancel = nil
				}
				var cCmd tea.Cmd
				m.chatDrawer, cCmd = m.chatDrawer.Update(messages.StreamFinishedMsg{})
				return m, cCmd
			}

			// Forward keys to active focused sub-panel in editor
			var sCmd, eCmd, cCmd, statusCmd tea.Cmd
			m.sidebar, sCmd = m.sidebar.Update(msg)
			m.editor, eCmd = m.editor.Update(msg)
			if m.showChatDrawer {
				m.chatDrawer, cCmd = m.chatDrawer.Update(msg)
			}
			m.statusbar, statusCmd = m.statusbar.Update(msg)
			return m, tea.Batch(sCmd, eCmd, cCmd, statusCmd)
		}
	}

	// Forward non-key messages to child components
	var lCmd, cfgCmd, mCmd, smCmd, cpCmd, navCmd, sCmd, eCmd, cCmd, stCmd tea.Cmd
	m.launcher, lCmd = m.launcher.Update(msg)
	m.llmConfig, cfgCmd = m.llmConfig.Update(msg)
	m.modal, mCmd = m.modal.Update(msg)
	m.searchModal, smCmd = m.searchModal.Update(msg)
	m.commandPalette, cpCmd = m.commandPalette.Update(msg)
	m.navbar, navCmd = m.navbar.Update(msg)
	m.sidebar, sCmd = m.sidebar.Update(msg)
	m.editor, eCmd = m.editor.Update(msg)
	m.chatDrawer, cCmd = m.chatDrawer.Update(msg)
	m.statusbar, stCmd = m.statusbar.Update(msg)

	cmds = append(cmds, lCmd, cfgCmd, mCmd, smCmd, cpCmd, navCmd, sCmd, eCmd, cCmd, stCmd)
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
	m.searchModal.SetSize(m.width, m.height)
	m.commandPalette.SetSize(m.width, m.height)
	m.launcher.SetSize(m.width, m.height)
	m.llmConfig.SetSize(m.width, m.height)

	navbarHeight := 1
	statusHeight := 1
	mainHeight := m.height - navbarHeight - statusHeight
	if mainHeight < 5 {
		mainHeight = 5
	}

	m.navbar.SetWidth(m.width)
	m.statusbar.SetWidth(m.width)

	if m.showChatDrawer {
		var sidebarWidth, drawerWidth, editorWidth int

		if m.width >= 90 {
			sidebarWidth = SidebarDefaultWidth
			if sidebarWidth > m.width/4 {
				sidebarWidth = m.width / 4
			}
			drawerWidth = int(float64(m.width) * 0.35)
			if drawerWidth < 32 {
				drawerWidth = 32
			}
			if drawerWidth > m.width/2 {
				drawerWidth = m.width / 2
			}
			editorWidth = m.width - sidebarWidth - drawerWidth
			if editorWidth < 25 {
				editorWidth = 25
				drawerWidth = m.width - sidebarWidth - editorWidth
			}
		} else {
			sidebarWidth = 20
			if sidebarWidth > m.width/3 {
				sidebarWidth = m.width / 3
			}
			drawerWidth = 30
			if drawerWidth > m.width/2 {
				drawerWidth = m.width / 2
			}
			editorWidth = m.width - sidebarWidth - drawerWidth
			if editorWidth < 15 {
				editorWidth = 15
			}
		}

		m.sidebar.SetSize(sidebarWidth, mainHeight)
		m.editor.SetSize(editorWidth, mainHeight)
		m.chatDrawer.SetDimensions(drawerWidth, mainHeight)
	} else {
		sidebarWidth := SidebarDefaultWidth
		if sidebarWidth > m.width/2 {
			sidebarWidth = m.width / 3
		}
		editorWidth := m.width - sidebarWidth

		m.sidebar.SetSize(sidebarWidth, mainHeight)
		m.editor.SetSize(editorWidth, mainHeight)
	}
}

// View renders the multi-view TUI with modal overlays.
func (m RootModel) View() string {
	if !m.ready {
		return "Initializing Novel TUI..."
	}

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

	var baseView string
	switch m.viewState {
	case messages.ViewStateLauncher:
		baseView = m.launcher.View()

	case messages.ViewStateLLMConfig:
		baseView = m.llmConfig.View()

	case messages.ViewStateEditor:
		navView := m.navbar.View()
		sidebarView := m.sidebar.View()
		editorView := m.editor.View()
		var mainView string
		if m.showChatDrawer {
			chatView := m.chatDrawer.View()
			mainView = lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, editorView, chatView)
		} else {
			mainView = lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, editorView)
		}
		statusView := m.statusbar.View()
		baseView = lipgloss.JoinVertical(lipgloss.Left, navView, mainView, statusView)
	}

	if m.commandPalette.Active {
		return m.commandPalette.View()
	}

	if m.searchModal.Active {
		return m.searchModal.View()
	}

	if m.modal.Active {
		return m.modal.View()
	}

	return baseView
}
