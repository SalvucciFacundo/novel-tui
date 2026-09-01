package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

// WorkspaceService manages workspace validation and bootstrap.
type WorkspaceService struct {
	chapterRepo   domain.ChapterRepository
	characterRepo domain.CharacterRepository
	baseDir       string
}

// NewWorkspaceService creates a new WorkspaceService.
func NewWorkspaceService(baseDir string, chapterRepo domain.ChapterRepository, characterRepo domain.CharacterRepository) *WorkspaceService {
	return &WorkspaceService{
		baseDir:       baseDir,
		chapterRepo:   chapterRepo,
		characterRepo: characterRepo,
	}
}

// EnsureWorkspace bootstraps folders and starter data if the project is empty.
func (s *WorkspaceService) EnsureWorkspace() error {
	chaptersDir := filepath.Join(s.baseDir, "chapters")
	if err := os.MkdirAll(chaptersDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspace chapters dir: %w", err)
	}

	// Check if chapters exist; if not, create a starter chapter
	chapters, err := s.chapterRepo.ListAll()
	if err == nil && len(chapters) == 0 {
		_, _ = s.chapterRepo.Create("Chapter 1: The Beginning")
	}

	// Check if characters exist; if not, create default starter characters
	chars, err := s.characterRepo.ListAll()
	if err == nil && len(chars) == 0 {
		starterChars := []domain.Character{
			{
				ID:          "protagonist",
				Name:        "Ren",
				Role:        "Protagonist",
				Description: "A quiet writer with a mysterious pocket watch.",
				Notes:       "Key trait: Observant, seeks the truth behind lost memories.",
			},
			{
				ID:          "mentor",
				Name:        "Professor Thorne",
				Role:        "Mentor",
				Description: "Archivist at the Grand Library.",
				Notes:       "Holds the cipher to the ancient manuscripts.",
			},
		}
		_ = s.characterRepo.SaveAll(starterChars)
	}

	return nil
}
