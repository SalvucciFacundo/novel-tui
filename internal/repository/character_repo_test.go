package repository_test

import (
	"os"
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/repository"
)

func TestFileCharacterRepository(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-character-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := repository.NewFileCharacterRepository(tempDir)

	// 1. Initial list empty
	chars, err := repo.ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(chars) != 0 {
		t.Errorf("expected 0 characters, got %d", len(chars))
	}

	// 2. Save All
	sampleChars := []domain.Character{
		{
			ID:          "char-1",
			Name:        "Lyra",
			Role:        "Mage",
			Description: "A wandering elementalist.",
			Notes:       "Specializes in frost magic.",
		},
	}
	if err := repo.SaveAll(sampleChars); err != nil {
		t.Fatalf("SaveAll failed: %v", err)
	}

	// 3. Read back
	loadedChars, err := repo.ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(loadedChars) != 1 {
		t.Fatalf("expected 1 character, got %d", len(loadedChars))
	}
	if loadedChars[0].Name != "Lyra" || loadedChars[0].Role != "Mage" {
		t.Errorf("loaded character mismatch: %+v", loadedChars[0])
	}
}
