package service_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
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

func TestBrainService_TimelineExtractionAndRecording(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "brain.db")

	repo, err := repository.NewSQLiteBrainRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}
	defer repo.Close()

	brainSvc := service.NewBrainService(repo)
	ctx := context.Background()

	// 1. BuildTimelineExtractionPrompt
	prompt := brainSvc.BuildTimelineExtractionPrompt("Siglos atrás, los antiguos forjaron la espada. Hoy Kuno parte hacia el valle.", "Capítulo 1")
	if prompt == "" || !strings.Contains(prompt, "Cronología") && !strings.Contains(prompt, "cronológica") {
		t.Errorf("expected timeline extraction prompt, got:\n%s", prompt)
	}

	// 2. ParseExtractedTimelineEvents with codeblock
	rawResponse := "Aquí está la cronología de eventos identificados:\n```json\n[\n  {\n    \"chronological_order\": 1,\n    \"period\": \"Era Antigua\",\n    \"title\": \"Forja de la Espada\",\n    \"description\": \"Los antiguos forjan la espada sagrada\",\n    \"characters\": [\"Antiguos\"]\n  },\n  {\n    \"chronological_order\": 2,\n    \"period\": \"Presente\",\n    \"title\": \"Partida al Valle\",\n    \"description\": \"Kuno emprende su viaje hacia el valle\",\n    \"characters\": [\"Kuno\"]\n  }\n]\n```\n"

	events, err := brainSvc.ParseExtractedTimelineEvents(rawResponse, "cap-1")
	if err != nil {
		t.Fatalf("ParseExtractedTimelineEvents failed: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 timeline events, got %d", len(events))
	}
	if events[0].Title != "Forja de la Espada" || events[0].ChronologicalOrder != 1 {
		t.Errorf("unexpected event 0: %+v", events[0])
	}
	if events[1].Title != "Partida al Valle" || events[1].ChronologicalOrder != 2 {
		t.Errorf("unexpected event 1: %+v", events[1])
	}

	// 3. RecordAutoLearnedTimelineEvents
	recorded, err := brainSvc.RecordAutoLearnedTimelineEvents(ctx, rawResponse, "cap-1")
	if err != nil {
		t.Fatalf("RecordAutoLearnedTimelineEvents failed: %v", err)
	}
	if len(recorded) != 2 {
		t.Fatalf("expected 2 recorded events, got %d", len(recorded))
	}

	// Verify persistence in SQLite repository
	inDB, err := repo.ListTimelineEvents(ctx)
	if err != nil {
		t.Fatalf("ListTimelineEvents failed: %v", err)
	}
	if len(inDB) != 2 {
		t.Fatalf("expected 2 events in DB, got %d", len(inDB))
	}

	// 4. FormatTimelineForPrompt
	formatted := brainSvc.FormatTimelineForPrompt(inDB)
	if !strings.Contains(formatted, "CRONOLOGÍA") && !strings.Contains(formatted, "TIMELINE") {
		t.Errorf("expected timeline header in formatted prompt, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "Forja de la Espada") || !strings.Contains(formatted, "Partida al Valle") {
		t.Errorf("expected events in formatted prompt, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "Era Antigua") || !strings.Contains(formatted, "Kuno") {
		t.Errorf("expected period and characters in formatted prompt, got:\n%s", formatted)
	}

	// 5. Empty events handling
	emptyEvents, err := brainSvc.ParseExtractedTimelineEvents("[]", "cap-1")
	if err != nil || len(emptyEvents) != 0 {
		t.Errorf("expected empty slice for empty json array, got %v (err: %v)", emptyEvents, err)
	}
	emptyFormatted := brainSvc.FormatTimelineForPrompt([]domain.TimelineEvent{})
	if emptyFormatted != "" {
		t.Errorf("expected empty string for empty timeline events list, got: %s", emptyFormatted)
	}
	// 6. Triangulation: Missing order and period defaults
	rawWithoutDefaults := "```json\n[{\"title\": \"Suceso Inesperado\", \"description\": \"Ocurre algo sin orden explícito\"}]\n```"
	parsedDefaults, err := brainSvc.ParseExtractedTimelineEvents(rawWithoutDefaults, "cap-2")
	if err != nil {
		t.Fatalf("unexpected error parsing events without defaults: %v", err)
	}
	if len(parsedDefaults) != 1 || parsedDefaults[0].ChronologicalOrder != 1 || parsedDefaults[0].Period != "Presente" {
		t.Errorf("expected defaults applied, got %+v", parsedDefaults)
	}

	// Triangulation: Invalid JSON returns error
	_, err = brainSvc.ParseExtractedTimelineEvents("```json\n[invalid\n```", "cap-2")
	if err == nil {
		t.Errorf("expected error for invalid JSON")
	}

	// Triangulation: Nil repository in RecordAutoLearnedTimelineEvents
	nilBrainSvc := service.NewBrainService(nil)
	nilRes, err := nilBrainSvc.RecordAutoLearnedTimelineEvents(ctx, rawResponse, "cap-1")
	if err != nil || len(nilRes) != 2 {
		t.Errorf("expected nil repo to return parsed events without error, got %d events (err: %v)", len(nilRes), err)
	}
}

