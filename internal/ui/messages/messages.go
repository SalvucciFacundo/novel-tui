package messages

import (
	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

// FocusState represents which panel currently has keyboard focus.
type FocusState int

const (
	FocusSidebar FocusState = iota
	FocusEditor
)

// FocusMsg notifies components of focus changes.
type FocusMsg struct {
	Target FocusState
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
