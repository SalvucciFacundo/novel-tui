package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

// FileCharacterRepository implements domain.CharacterRepository using JSON persistence.
type FileCharacterRepository struct {
	baseDir string
}

// NewFileCharacterRepository creates a new FileCharacterRepository.
func NewFileCharacterRepository(baseDir string) *FileCharacterRepository {
	return &FileCharacterRepository{baseDir: baseDir}
}

func (r *FileCharacterRepository) filePath() string {
	return filepath.Join(r.baseDir, "characters.json")
}

// ListAll loads all characters from characters.json.
func (r *FileCharacterRepository) ListAll() ([]domain.Character, error) {
	path := r.filePath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []domain.Character{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read characters file: %w", err)
	}

	if len(data) == 0 {
		return []domain.Character{}, nil
	}

	var characters []domain.Character
	if err := json.Unmarshal(data, &characters); err != nil {
		return nil, fmt.Errorf("failed to parse characters JSON: %w", err)
	}

	return characters, nil
}

// SaveAll serializes all characters to characters.json atomically.
func (r *FileCharacterRepository) SaveAll(chars []domain.Character) error {
	path := r.filePath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for characters: %w", err)
	}

	data, err := json.MarshalIndent(chars, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal characters: %w", err)
	}

	tmpPath := filepath.Join(dir, ".characters.json.tmp")
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp character file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp character file: %w", err)
	}

	return nil
}
