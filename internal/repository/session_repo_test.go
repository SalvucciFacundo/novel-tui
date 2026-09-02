package repository_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/repository"
)

func TestFileChatSessionRepository_SaveAndGet(t *testing.T) {
	tempDir := t.TempDir()
	repo := repository.NewFileChatSessionRepository()

	session := domain.ChatSession{
		ID:          "test_session_1",
		Title:       "Mi Sesión de Prueba",
		NovelPath:   tempDir,
		EffortLevel: domain.EffortMedium,
		Messages: []domain.ChatMessage{
			{Role: "user", Content: "Hola asistente", Timestamp: time.Now().UTC()},
			{Role: "assistant", Content: "Hola, ¿en qué te ayudo?", Timestamp: time.Now().UTC()},
		},
	}

	err := repo.Save(session)
	if err != nil {
		t.Fatalf("expected no error saving session, got %v", err)
	}

	retrieved, err := repo.Get(tempDir, "test_session_1")
	if err != nil {
		t.Fatalf("expected no error getting session, got %v", err)
	}

	if retrieved.ID != session.ID {
		t.Errorf("expected ID %s, got %s", session.ID, retrieved.ID)
	}
	if retrieved.Title != session.Title {
		t.Errorf("expected Title %s, got %s", session.Title, retrieved.Title)
	}
	if retrieved.EffortLevel != domain.EffortMedium {
		t.Errorf("expected EffortLevel %s, got %s", domain.EffortMedium, retrieved.EffortLevel)
	}
	if len(retrieved.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(retrieved.Messages))
	}
	if retrieved.Messages[0].Content != "Hola asistente" {
		t.Errorf("expected message 0 content 'Hola asistente', got '%s'", retrieved.Messages[0].Content)
	}
}

func TestFileChatSessionRepository_AutoTitle(t *testing.T) {
	tempDir := t.TempDir()
	repo := repository.NewFileChatSessionRepository()

	session := domain.ChatSession{
		NovelPath:   tempDir,
		EffortLevel: domain.EffortLow,
		Messages: []domain.ChatMessage{
			{Role: "user", Content: "Revisar diálogo entre Elena y Kuno cuando descubren el artefacto antiguo"},
		},
	}

	err := repo.Save(session)
	if err != nil {
		t.Fatalf("Save error: %v", err)
	}

	list, err := repo.List(tempDir)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 session, got %d", len(list))
	}

	saved := list[0]
	if len([]rune(saved.Title)) > 40 {
		t.Errorf("title exceeded 40 runes: %s (len %d)", saved.Title, len([]rune(saved.Title)))
	}
	expectedPrefix := "Revisar diálogo entre Elena y Kuno"
	if len(saved.Title) < len(expectedPrefix) || saved.Title[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("expected title to start with %s, got %s", expectedPrefix, saved.Title)
	}
}

func TestFileChatSessionRepository_ListOrder(t *testing.T) {
	tempDir := t.TempDir()
	repo := repository.NewFileChatSessionRepository()

	s1 := domain.ChatSession{
		ID:        "session_old",
		NovelPath: tempDir,
		Title:     "Old Session",
		Messages:  []domain.ChatMessage{{Role: "user", Content: "Old"}},
	}
	if err := repo.Save(s1); err != nil {
		t.Fatalf("Save s1 error: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	s2 := domain.ChatSession{
		ID:        "session_new",
		NovelPath: tempDir,
		Title:     "New Session",
		Messages:  []domain.ChatMessage{{Role: "user", Content: "New"}},
	}
	if err := repo.Save(s2); err != nil {
		t.Fatalf("Save s2 error: %v", err)
	}

	list, err := repo.List(tempDir)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(list))
	}

	if list[0].ID != "session_new" {
		t.Errorf("expected first session to be session_new, got %s", list[0].ID)
	}
	if list[1].ID != "session_old" {
		t.Errorf("expected second session to be session_old, got %s", list[1].ID)
	}
}

func TestFileChatSessionRepository_Delete(t *testing.T) {
	tempDir := t.TempDir()
	repo := repository.NewFileChatSessionRepository()

	s := domain.ChatSession{
		ID:        "session_to_delete",
		NovelPath: tempDir,
		Title:     "To Delete",
		Messages:  []domain.ChatMessage{{Role: "user", Content: "Delete me"}},
	}
	if err := repo.Save(s); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Confirm file exists
	filePath := filepath.Join(tempDir, "chats", "session_to_delete.json")
	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("expected file to exist at %s", filePath)
	}

	if err := repo.Delete(tempDir, "session_to_delete"); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	// Confirm file deleted
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("expected file to be deleted")
	}

	// Delete non-existent should not fail
	if err := repo.Delete(tempDir, "non_existent"); err != nil {
		t.Errorf("expected no error deleting non-existent session, got %v", err)
	}

	// Get non-existent should fail
	if _, err := repo.Get(tempDir, "session_to_delete"); err == nil {
		t.Errorf("expected error getting deleted session")
	}
}
