package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

// FileChatSessionRepository implements domain.ChatSessionRepository using JSON files on disk.
type FileChatSessionRepository struct{}

// NewFileChatSessionRepository creates a new FileChatSessionRepository.
func NewFileChatSessionRepository() *FileChatSessionRepository {
	return &FileChatSessionRepository{}
}

func (r *FileChatSessionRepository) chatsDir(novelDir string) string {
	return filepath.Join(ExpandHome(novelDir), "chats")
}

// Save persists a chat session into <novel_dir>/chats/<session_id>.json.
func (r *FileChatSessionRepository) Save(session domain.ChatSession) error {
	if session.NovelPath == "" {
		return fmt.Errorf("novel path cannot be empty")
	}

	session.NovelPath = ExpandHome(session.NovelPath)
	dir := r.chatsDir(session.NovelPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create chats directory: %w", err)
	}

	if session.ID == "" {
		session.ID = fmt.Sprintf("session_%s", time.Now().UTC().Format("20060102_150405"))
	}

	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now().UTC()
	}
	session.UpdatedAt = time.Now().UTC()

	if session.Title == "" {
		session.Title = deriveSessionTitle(session.Messages)
	}

	filePath := filepath.Join(dir, fmt.Sprintf("%s.json", session.ID))
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal chat session: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write chat session file: %w", err)
	}

	return nil
}

// List retrieves all chat sessions for a given novel directory, ordered by UpdatedAt descending.
func (r *FileChatSessionRepository) List(novelDir string) ([]domain.ChatSession, error) {
	dir := r.chatsDir(novelDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to ensure chats directory: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read chats directory: %w", err)
	}

	var sessions []domain.ChatSession
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var session domain.ChatSession
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		sessions = append(sessions, session)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

// Get retrieves a specific chat session by its ID.
func (r *FileChatSessionRepository) Get(novelDir, id string) (domain.ChatSession, error) {
	dir := r.chatsDir(novelDir)
	filePath := filepath.Join(dir, fmt.Sprintf("%s.json", id))

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.ChatSession{}, fmt.Errorf("chat session not found: %s", id)
		}
		return domain.ChatSession{}, fmt.Errorf("failed to read chat session: %w", err)
	}

	var session domain.ChatSession
	if err := json.Unmarshal(data, &session); err != nil {
		return domain.ChatSession{}, fmt.Errorf("failed to parse chat session: %w", err)
	}

	return session, nil
}

// Delete removes a chat session JSON file from disk.
func (r *FileChatSessionRepository) Delete(novelDir, id string) error {
	dir := r.chatsDir(novelDir)
	filePath := filepath.Join(dir, fmt.Sprintf("%s.json", id))

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete chat session: %w", err)
	}

	return nil
}

// deriveSessionTitle extracts a short title from the first user message or fallback.
func deriveSessionTitle(messages []domain.ChatMessage) string {
	for _, msg := range messages {
		if msg.Role == "user" && strings.TrimSpace(msg.Content) != "" {
			trimmed := strings.TrimSpace(msg.Content)
			// Remove newlines
			trimmed = strings.ReplaceAll(trimmed, "\n", " ")
			runes := []rune(trimmed)
			if len(runes) > 40 {
				return string(runes[:40])
			}
			return string(runes)
		}
	}
	return "Nueva Conversación"
}
