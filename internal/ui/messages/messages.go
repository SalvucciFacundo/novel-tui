package messages

import (
	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

// ViewState represents the active top-level screen view.
type ViewState int

const (
	ViewStateLauncher ViewState = iota
	ViewStateEditor
	ViewStateLLMConfig
)

// ModalPurpose identifies the intent of an open modal dialog.
type ModalPurpose string

const (
	ModalPurposeNewNovel        ModalPurpose = "new_novel"
	ModalPurposeNewChapter      ModalPurpose = "new_chapter"
	ModalPurposeSetRootDir      ModalPurpose = "set_root_dir"
	ModalPurposeOpenFolder      ModalPurpose = "open_folder"
	ModalPurposeConfigureGenres ModalPurpose = "configure_genres"
)

// FocusState represents which panel currently has keyboard focus.
type FocusState int

const (
	FocusSidebar FocusState = iota
	FocusEditor
	FocusChat
)

// FocusMsg notifies components of focus changes.
type FocusMsg struct {
	Target FocusState
}

// ChangeViewMsg requests a transition to another top-level view state.
type ChangeViewMsg struct {
	View ViewState
}

// ShowModalMsg requests opening a centered modal dialog.
type ShowModalMsg struct {
	Purpose      ModalPurpose
	Title        string
	Prompt       string
	InitialValue string
	ErrorMsg     string
}

// HideModalMsg closes the active modal dialog without committing.
type HideModalMsg struct{}

// SubmitModalMsg is dispatched when a modal dialog is confirmed with input.
type SubmitModalMsg struct {
	Purpose ModalPurpose
	Value   string
}

// CreateNovelMsg requests scaffolding a new novel project.
type CreateNovelMsg struct {
	Title string
}

// CreateChapterMsg requests creating a new sequential chapter in the active novel.
type CreateChapterMsg struct {
	Title string
}

// OpenNovelMsg requests opening a specific novel folder path in the editor.
type OpenNovelMsg struct {
	Path string
}

// SetRootDirMsg requests updating the global root directory for novels.
type SetRootDirMsg struct {
	Path string
}

// SaveLLMConfigMsg requests persisting modified LLM configuration settings.
type SaveLLMConfigMsg struct {
	Config domain.LLMConfig
}

// ConfigLoadedMsg informs components of the loaded global configuration.
type ConfigLoadedMsg struct {
	Config *domain.AppConfig
}

// NovelListRefreshedMsg is sent when recent novels in root dir are re-scanned.
type NovelListRefreshedMsg struct {
	Novels []domain.NovelMetadata
}

// NotificationMsg presents a transient status message to the user.
type NotificationMsg struct {
	Message string
	IsError bool
}

// ChapterSelectedMsg is emitted when a chapter is selected in the sidebar.
type ChapterSelectedMsg struct {
	Chapter domain.Chapter
}

// ChapterCreatedMsg is emitted when a new chapter is created.
type ChapterCreatedMsg struct {
	Chapter domain.Chapter
}

// TextChangedMsg is emitted by the editor when text is modified.
type TextChangedMsg struct {
	ChapterID string
	Content   string
	Metrics   domain.EditorMetrics
}

// SaveRequestedMsg is emitted when the user presses Ctrl+S.
type SaveRequestedMsg struct {
	ChapterID string
	Content   string
}

// SaveCompletedMsg is emitted after a file write operation finishes.
type SaveCompletedMsg struct {
	ChapterID string
	Success   bool
	Error     error
}

// ReloadChaptersMsg triggers a reload of the chapter list from disk.
type ReloadChaptersMsg struct{}

// SelectSidebarTabMsg requests switching the active sidebar tab.
type SelectSidebarTabMsg struct {
	Tab int
}

// ToggleChatDrawerMsg requests opening or closing the LLM chat drawer.
type ToggleChatDrawerMsg struct{}

// SendChatMessageMsg requests sending a prompt to the LLM assistant.
type SendChatMessageMsg struct {
	Content     string
	EffortLevel domain.LLMEffortLevel
}

// TokenReceivedMsg is emitted when a new token delta is received from the LLM stream.
type TokenReceivedMsg struct {
	Content string
}

// StreamTokenMsg is an alias for TokenReceivedMsg for compatibility.
type StreamTokenMsg = TokenReceivedMsg

// StreamFinishedMsg is emitted when LLM streaming concludes successfully.
type StreamFinishedMsg struct{}

// StreamErrorMsg is emitted when an error occurs during LLM streaming.
type StreamErrorMsg struct {
	Err error
}

// SelectSessionMsg requests switching to an existing chat session.
type SelectSessionMsg struct {
	SessionID string
}

// CreateSessionMsg requests creating a new blank chat session.
type CreateSessionMsg struct{}

// SetEffortLevelMsg requests updating the active LLM effort level.
type SetEffortLevelMsg struct {
	EffortLevel domain.LLMEffortLevel
}

// BrainActivityMsg is emitted when Brain extracts, indexes, or updates facts/summaries.
type BrainActivityMsg struct {
	Event domain.BrainActivityEvent
}

// OpenGlobalSearchMsg requests opening the global search and replace modal.
type OpenGlobalSearchMsg struct{}

// CloseGlobalSearchMsg requests closing the global search modal.
type CloseGlobalSearchMsg struct{}

// JumpToMatchMsg requests navigating the editor to a specific search match.
type JumpToMatchMsg struct {
	Match domain.SearchMatch
}

// GlobalReplaceMsg requests replacing occurrences of Query with Replacement novel-wide.
type GlobalReplaceMsg struct {
	Query         string
	Replacement   string
	CaseSensitive bool
}

// GlobalReplaceCompletedMsg is emitted after a global replace operation finishes.
type GlobalReplaceCompletedMsg struct {
	Result domain.SearchReplaceResult
}

// OpenCommandPaletteMsg requests opening the command palette modal.
type OpenCommandPaletteMsg struct{}

// CloseCommandPaletteMsg requests closing the command palette modal.
type CloseCommandPaletteMsg struct{}

// ExecuteCommandMsg is dispatched when a command item is selected and executed.
type ExecuteCommandMsg struct {
	Command domain.CommandItem
}

// StartColabServerMsg triggers the provisioning and startup of the Colab GPU LLM server.
type StartColabServerMsg struct{}

// ColabServerStartedMsg is emitted when the Colab GPU LLM server is successfully reachable.
type ColabServerStartedMsg struct {
	BaseURL string
}

// ColabServerErrorMsg is emitted when Colab provisioning or execution fails.
type ColabServerErrorMsg struct {
	Err error
}



