package domain

import (
	"time"
)

// ChatSession represents a persistent conversation thread associated with a novel.
type ChatSession struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	NovelPath   string         `json:"novel_path"`
	EffortLevel LLMEffortLevel `json:"effort_level"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Messages    []ChatMessage  `json:"messages"`
}

// ChatSessionRepository defines persistence contracts for chat sessions.
type ChatSessionRepository interface {
	Save(session ChatSession) error
	List(novelDir string) ([]ChatSession, error)
	Get(novelDir, id string) (ChatSession, error)
	Delete(novelDir, id string) error
}
