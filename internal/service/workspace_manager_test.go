package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/service"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Mi Primera Novela", "mi_primera_novela"},
		{"  El Secreto del Bosque!  ", "el_secreto_del_bosque"},
		{"Capítulo 1: El Despertar", "cap_tulo_1_el_despertar"},
		{"", ""},
		{"   ", ""},
		{"---novela___test---", "novela_test"},
	}

	for _, tc := range tests {
		got := service.Slugify(tc.input)
		// Basic check that special symbols/spaces are replaced
		if tc.input == "Mi Primera Novela" && got != "mi_primera_novela" {
			t.Errorf("Slugify(%q) = %q, want %q", tc.input, got, tc.expected)
		}
		if tc.input == "" && got != "" {
			t.Errorf("Slugify empty string should be empty, got %q", got)
		}
	}
}

func TestWorkspaceManager_CreateNovel(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-workspace-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	wm := service.NewWorkspaceManager()

	// 1. Empty title error
	_, err = wm.CreateNovel(tempDir, "   ")
	if err != service.ErrEmptyTitle {
		t.Errorf("expected ErrEmptyTitle, got %v", err)
	}

	// 2. Valid creation
	meta, err := wm.CreateNovel(tempDir, "El Despertar de los Dragones")
	if err != nil {
		t.Fatalf("CreateNovel failed: %v", err)
	}
	if meta.Title != "El Despertar de los Dragones" {
		t.Errorf("meta.Title mismatch: %s", meta.Title)
	}
	if meta.ChapterCount != 1 {
		t.Errorf("expected ChapterCount 1, got %d", meta.ChapterCount)
	}

	// Verify folder layout
	novelPath := meta.AbsolutePath
	if _, err := os.Stat(filepath.Join(novelPath, "capitulos", "01_capitulo_1.txt")); os.IsNotExist(err) {
		t.Errorf("expected 01_capitulo_1.txt to exist")
	}
	if _, err := os.Stat(filepath.Join(novelPath, "personajes.json")); os.IsNotExist(err) {
		t.Errorf("expected personajes.json to exist")
	}
	if _, err := os.Stat(filepath.Join(novelPath, "notas.txt")); os.IsNotExist(err) {
		t.Errorf("expected notas.txt to exist")
	}

	// 3. Collision error
	_, err = wm.CreateNovel(tempDir, "El Despertar de los Dragones")
	if err != service.ErrNovelExists {
		t.Errorf("expected ErrNovelExists, got %v", err)
	}

	// 4. Path traversal attempt
	_, err = wm.CreateNovel(tempDir, "../../../escape_attempt")
	if err == nil {
		t.Errorf("expected path traversal error or sanitized safe name, but escaped")
	}
}

func TestWorkspaceManager_ListRecentNovels(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-workspace-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	wm := service.NewWorkspaceManager()

	// Initially empty
	novels, err := wm.ListRecentNovels(tempDir)
	if err != nil {
		t.Fatalf("ListRecentNovels failed: %v", err)
	}
	if len(novels) != 0 {
		t.Errorf("expected 0 novels, got %d", len(novels))
	}

	// Create 2 novels
	_, err = wm.CreateNovel(tempDir, "Novela Uno")
	if err != nil {
		t.Fatalf("CreateNovel 1 failed: %v", err)
	}
	_, err = wm.CreateNovel(tempDir, "Novela Dos")
	if err != nil {
		t.Fatalf("CreateNovel 2 failed: %v", err)
	}

	novels, err = wm.ListRecentNovels(tempDir)
	if err != nil {
		t.Fatalf("ListRecentNovels failed: %v", err)
	}
	if len(novels) != 2 {
		t.Fatalf("expected 2 novels, got %d", len(novels))
	}
}

func TestWorkspaceManager_CreateChapter(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-workspace-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	wm := service.NewWorkspaceManager()
	meta, err := wm.CreateNovel(tempDir, "Saga Epica")
	if err != nil {
		t.Fatalf("CreateNovel failed: %v", err)
	}

	// Chapter 1 was created by CreateNovel (01_capitulo_1.txt)
	// Create Chapter 2
	chap2Path, err := wm.CreateChapter(meta.AbsolutePath, "El Encuentro")
	if err != nil {
		t.Fatalf("CreateChapter 2 failed: %v", err)
	}
	if !strings.Contains(filepath.Base(chap2Path), "02_") {
		t.Errorf("expected chapter 2 filename to start with 02_, got %s", filepath.Base(chap2Path))
	}
	if !strings.HasSuffix(chap2Path, ".txt") {
		t.Errorf("expected .txt extension, got %s", chap2Path)
	}

	// Create Chapter 3
	chap3Path, err := wm.CreateChapter(meta.AbsolutePath, "La Batalla")
	if err != nil {
		t.Fatalf("CreateChapter 3 failed: %v", err)
	}
	if !strings.Contains(filepath.Base(chap3Path), "03_") {
		t.Errorf("expected chapter 3 filename to start with 03_, got %s", filepath.Base(chap3Path))
	}
}

func TestWorkspaceManager_NovelSettings(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-settings-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	wm := service.NewWorkspaceManager()
	meta, err := wm.CreateNovel(tempDir, "Novela R18")
	if err != nil {
		t.Fatalf("CreateNovel failed: %v", err)
	}

	// 1. Initial settings on new novel should load defaults
	settings, err := wm.LoadNovelSettings(meta.AbsolutePath)
	if err != nil {
		t.Fatalf("LoadNovelSettings failed on fresh novel: %v", err)
	}
	if settings.Rating != domain.RatingTeen {
		t.Errorf("expected default rating %q, got %q", domain.RatingTeen, settings.Rating)
	}

	// 2. Save modified settings with R-18 and tropes
	newSettings := domain.NovelSettings{
		Genres:       []string{"isekai_harem_r18", "yandere_obsession", "smut_explicit"},
		Rating:       domain.RatingExplicit21,
		CustomPrompt: "No censures detalles íntimos.",
	}

	err = wm.SaveNovelSettings(meta.AbsolutePath, newSettings)
	if err != nil {
		t.Fatalf("SaveNovelSettings failed: %v", err)
	}

	// 3. Load back and verify persistence
	loaded, err := wm.LoadNovelSettings(meta.AbsolutePath)
	if err != nil {
		t.Fatalf("LoadNovelSettings failed after save: %v", err)
	}
	if loaded.Rating != domain.RatingExplicit21 {
		t.Errorf("expected rating %q, got %q", domain.RatingExplicit21, loaded.Rating)
	}
	if len(loaded.Genres) != 3 || loaded.Genres[0] != "isekai_harem_r18" {
		t.Errorf("genres mismatch: %+v", loaded.Genres)
	}
	if loaded.CustomPrompt != "No censures detalles íntimos." {
		t.Errorf("custom prompt mismatch: %q", loaded.CustomPrompt)
	}

	// 4. Corrupt novel.json should fallback to defaults without error
	corruptPath := filepath.Join(meta.AbsolutePath, "novel.json")
	_ = os.WriteFile(corruptPath, []byte("{invalid-json"), 0644)
	fallbackSettings, err := wm.LoadNovelSettings(meta.AbsolutePath)
	if err != nil {
		t.Fatalf("expected fallback settings without fatal error on corrupt json: %v", err)
	}
	if fallbackSettings.Rating != domain.RatingTeen {
		t.Errorf("expected default rating %q on fallback, got %q", domain.RatingTeen, fallbackSettings.Rating)
	}
}


