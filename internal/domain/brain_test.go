package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

func TestBrainFactJSONSerialization(t *testing.T) {
	fact := domain.BrainFact{
		ID:        "fact-1",
		Topic:     "Personajes",
		Concept:   "Kuno",
		Fact:      "Perdió el brazo en la batalla del norte",
		Type:      domain.FactTypeCharacter,
		ChapterID: "01",
		Tags:      []string{"herida", "batalla"},
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(fact)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var parsed domain.BrainFact
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if parsed.ID != fact.ID || parsed.Concept != fact.Concept || parsed.Type != fact.Type {
		t.Errorf("expected fact match, got %+v", parsed)
	}
}

func TestSessionSummaryJSONSerialization(t *testing.T) {
	summary := domain.SessionSummary{
		ID:         "sum-1",
		Summary:    "Avance en el capítulo 2",
		Highlights: []string{"Aparición del mentor", "Revelación de la profecía"},
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var parsed domain.SessionSummary
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if parsed.ID != summary.ID || len(parsed.Highlights) != 2 {
		t.Errorf("expected summary match, got %+v", parsed)
	}
}
