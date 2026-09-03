package service_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/repository"
	"github.com/SalvucciFacundo/novel-tui/internal/service"
)

func TestBrainService_FactExtractionAndRecording(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "brain.db")

	repo, err := repository.NewSQLiteBrainRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	defer repo.Close()

	brainSvc := service.NewBrainService(repo)
	ctx := context.Background()

	// 1. Test Prompt Generation
	prompt := brainSvc.BuildExtractionPrompt("Kuno empuñó la Espada del Alba mientras miraba a Elena.", "Capítulo 1")
	if prompt == "" || !testing.Short() && len(prompt) < 50 {
		t.Errorf("expected non-empty prompt, got %s", prompt)
	}

	// 2. Test Parsing Valid JSON response with codeblock
	rawResponse := "Aquí están los hechos identificados:\n```json\n[\n  {\n    \"topic\": \"Personajes\",\n    \"concept\": \"Elena\",\n    \"fact\": \"Es la arquera del grupo y compañera de Kuno\",\n    \"type\": \"character\",\n    \"tags\": [\"aliada\", \"arquera\"]\n  },\n  {\n    \"topic\": \"Objetos\",\n    \"concept\": \"Espada del Alba\",\n    \"fact\": \"Arma legendaria forjada en el primer solsticio\",\n    \"type\": \"lore\",\n    \"tags\": [\"arma\", \"solsticio\"]\n  }\n]\n```\nEspero te sirva."

	recorded, err := brainSvc.RecordAutoLearnedFacts(ctx, rawResponse, "cap-1")
	if err != nil {
		t.Fatalf("RecordAutoLearnedFacts failed: %v", err)
	}

	if len(recorded) != 2 {
		t.Fatalf("expected 2 recorded facts, got %d", len(recorded))
	}

	// Verify persistence in SQLite repository
	facts, err := repo.SearchFacts(ctx, "Elena", 5)
	if err != nil {
		t.Fatalf("SearchFacts failed: %v", err)
	}
	if len(facts) != 1 || facts[0].Concept != "Elena" {
		t.Errorf("expected Elena fact in DB, got %+v", facts)
	}

	// 3. Test Search Relevant Facts
	relevant, err := brainSvc.SearchRelevantFacts(ctx, "Elena disparó una flecha hacia el horizonte con la Espada", 5)
	if err != nil {
		t.Fatalf("SearchRelevantFacts failed: %v", err)
	}
	if len(relevant) == 0 {
		t.Errorf("expected relevant facts, got 0")
	}

	// 4. Test Format Facts For Prompt
	formatted := brainSvc.FormatFactsForPrompt(relevant)
	if formatted == "" {
		t.Errorf("expected non-empty formatted prompt string")
	}

	// 5. Test Empty / No facts response
	emptyRes, err := brainSvc.RecordAutoLearnedFacts(ctx, "[]", "cap-1")
	if err != nil {
		t.Fatalf("unexpected error on empty array: %v", err)
	}
	if len(emptyRes) != 0 {
		t.Errorf("expected 0 facts for empty array, got %d", len(emptyRes))
	}
}

func TestBrainService_SessionSummary(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "brain.db")

	repo, err := repository.NewSQLiteBrainRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	defer repo.Close()

	brainSvc := service.NewBrainService(repo)

	rawSummary := "```json\n{\n  \"summary\": \"Se avanzó en el enfrentamiento en las colinas.\",\n  \"highlights\": [\"Elena salva a Kuno\", \"El monstruo escapa\"]\n}\n```"

	summary, err := brainSvc.ParseSessionSummary(rawSummary)
	if err != nil {
		t.Fatalf("ParseSessionSummary failed: %v", err)
	}

	if summary.Summary != "Se avanzó en el enfrentamiento en las colinas." {
		t.Errorf("unexpected summary text: %s", summary.Summary)
	}
	if len(summary.Highlights) != 2 {
		t.Errorf("expected 2 highlights, got %d", len(summary.Highlights))
	}
}
