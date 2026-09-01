package repository_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/repository"
)

func TestFileChapterRepository(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-chapter-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo, err := repository.NewFileChapterRepository(tempDir)
	if err != nil {
		t.Fatalf("NewFileChapterRepository failed: %v", err)
	}

	// 1. Initial list should be empty
	chapters, err := repo.ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(chapters) != 0 {
		t.Errorf("expected 0 chapters, got %d", len(chapters))
	}

	// 2. Create Chapter
	newChap, err := repo.Create("The Gathering")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if !strings.Contains(newChap.ID, "the_gathering") {
		t.Errorf("unexpected ID: %s", newChap.ID)
	}

	// 3. List All should return 1 chapter
	chapters, err = repo.ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(chapters) != 1 {
		t.Fatalf("expected 1 chapter, got %d", len(chapters))
	}
	if chapters[0].Title != "The Gathering" {
		t.Errorf("unexpected title: %s", chapters[0].Title)
	}

	// 4. Save Content and Load Content
	newContent := "# The Gathering\n\nIt was a dark and stormy night in the realm of Eldoria."
	err = repo.SaveContent(newChap.ID, newContent)
	if err != nil {
		t.Fatalf("SaveContent failed: %v", err)
	}

	loadedContent, err := repo.LoadContent(newChap.ID)
	if err != nil {
		t.Fatalf("LoadContent failed: %v", err)
	}
	if loadedContent != newContent {
		t.Errorf("loadedContent mismatch.\nGot: %s\nWant: %s", loadedContent, newContent)
	}

	// 5. Check atomic file creation on disk in capitulos
	expectedFile := filepath.Join(tempDir, "capitulos", newChap.ID+".txt")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("expected chapter file at %s, but does not exist", expectedFile)
	}
}

func TestFileChapterRepository_LegacyChapters(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-legacy-chapter-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	legacyDir := filepath.Join(tempDir, "chapters")
	_ = os.MkdirAll(legacyDir, 0755)

	repo, err := repository.NewFileChapterRepository(tempDir)
	if err != nil {
		t.Fatalf("NewFileChapterRepository failed: %v", err)
	}

	newChap, err := repo.Create("Chapter 1: The Gathering")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if newChap.ID != "chapter-1-the-gathering" {
		t.Errorf("unexpected ID: %s", newChap.ID)
	}

	expectedFile := filepath.Join(legacyDir, newChap.ID+".md")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("expected chapter file at %s, but does not exist", expectedFile)
	}
}
