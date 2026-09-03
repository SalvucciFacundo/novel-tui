package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

func TestCommandItem_Properties(t *testing.T) {
	cmd := domain.CommandItem{
		ID:          "save_chapter",
		Title:       "Guardar Capítulo",
		Category:    "Editor",
		Shortcut:    "Ctrl+S",
		Description: "Guardar el capítulo actual",
	}

	if cmd.ID != "save_chapter" {
		t.Errorf("expected ID 'save_chapter', got %s", cmd.ID)
	}
	if cmd.Title != "Guardar Capítulo" {
		t.Errorf("expected Title 'Guardar Capítulo', got %s", cmd.Title)
	}
	if cmd.Category != "Editor" {
		t.Errorf("expected Category 'Editor', got %s", cmd.Category)
	}
	if cmd.Shortcut != "Ctrl+S" {
		t.Errorf("expected Shortcut 'Ctrl+S', got %s", cmd.Shortcut)
	}
	if cmd.Description != "Guardar el capítulo actual" {
		t.Errorf("expected Description 'Guardar el capítulo actual', got %s", cmd.Description)
	}
}

func TestDefaultCommands(t *testing.T) {
	commands := domain.DefaultCommands()
	if len(commands) == 0 {
		t.Fatalf("expected DefaultCommands to return a non-empty slice")
	}

	requiredIDs := []string{
		"global_search",
		"save_chapter",
		"new_chapter",
		"toggle_ai",
		"go_launcher",
		"tab_chapters",
		"tab_characters",
		"tab_notes",
		"tab_brain",
		"toggle_timeline",
		"llm_config",
	}

	foundMap := make(map[string]domain.CommandItem)
	for _, cmd := range commands {
		if cmd.ID == "" {
			t.Errorf("command has empty ID: %+v", cmd)
		}
		if cmd.Title == "" {
			t.Errorf("command %s has empty Title", cmd.ID)
		}
		if cmd.Category == "" {
			t.Errorf("command %s has empty Category", cmd.ID)
		}
		if cmd.Shortcut == "" {
			t.Errorf("command %s has empty Shortcut", cmd.ID)
		}
		if cmd.Description == "" {
			t.Errorf("command %s has empty Description", cmd.ID)
		}
		foundMap[cmd.ID] = cmd
	}

	for _, reqID := range requiredIDs {
		if _, ok := foundMap[reqID]; !ok {
			t.Errorf("expected command ID %q in DefaultCommands()", reqID)
		}
	}
}

func TestCommandItem_JSONSerialization(t *testing.T) {
	cmd := domain.CommandItem{
		ID:          "test_cmd",
		Title:       "Comando de Prueba",
		Category:    "Test",
		Shortcut:    "Ctrl+T",
		Description: "Descripción de prueba",
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("failed to marshal CommandItem: %v", err)
	}

	var decoded domain.CommandItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CommandItem: %v", err)
	}

	if decoded.ID != cmd.ID || decoded.Title != cmd.Title || decoded.Category != cmd.Category ||
		decoded.Shortcut != cmd.Shortcut || decoded.Description != cmd.Description {
		t.Errorf("decoded struct %+v does not match original %+v", decoded, cmd)
	}
}

