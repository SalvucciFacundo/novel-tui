package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

var jsonBlockRegex = regexp.MustCompile(`(?s)` + "```" + `(?:json)?\s*(.*?)\s*` + "```")

// BrainService orchestrates memory indexing, semantic context retrieval, and passive fact extraction.
type BrainService struct {
	repo domain.BrainRepository
}

// NewBrainService creates a new instance of BrainService.
func NewBrainService(repo domain.BrainRepository) *BrainService {
	return &BrainService{repo: repo}
}

// Repository returns the underlying BrainRepository.
func (s *BrainService) Repository() domain.BrainRepository {
	return s.repo
}

// SearchRelevantFacts finds facts matching the query or text context.
func (s *BrainService) SearchRelevantFacts(ctx context.Context, text string, limit int) ([]domain.BrainFact, error) {
	if s.repo == nil {
		return nil, nil
	}
	clean := strings.TrimSpace(text)
	if clean == "" {
		return s.repo.ListRecentFacts(ctx, limit)
	}

	// Extract meaningful keywords from prompt or chapter snippet
	keywords := extractKeywords(clean, 5)
	if len(keywords) > 0 {
		query := strings.Join(keywords, " ")
		results, err := s.repo.SearchFacts(ctx, query, limit)
		if err == nil && len(results) > 0 {
			return results, nil
		}
	}

	return s.repo.SearchFacts(ctx, clean, limit)
}

// BuildExtractionPrompt constructs the system/user prompt for autonomous fact discovery.
func (s *BrainService) BuildExtractionPrompt(content string, chapterTitle string) string {
	var sb strings.Builder
	sb.WriteString("Eres el módulo de memoria y continuidad narrativa 'Brain' de Novel-TUI.\n")
	sb.WriteString("Analiza el siguiente texto de la novela o conversación de autoría y extrae ÚNICAMENTE hechos nuevos, cambios en personajes, elementos de lore, lugares o decisiones narrativas clave.\n\n")
	if chapterTitle != "" {
		sb.WriteString(fmt.Sprintf("Capítulo de referencia: %s\n\n", chapterTitle))
	}
	sb.WriteString("TEXTO A ANALIZAR:\n")
	sb.WriteString(content)
	sb.WriteString("\n\n")
	sb.WriteString("INSTRUCCIONES DE RESPUESTA:\n")
	sb.WriteString("1. Si NO hay hechos nuevos relevantes o cambios de lore, responde con una lista JSON vacía: []\n")
	sb.WriteString("2. Si encuentras hechos, responde EXCLUSIVAMENTE con un arreglo JSON en este formato:\n")
	sb.WriteString("```json\n")
	sb.WriteString("[\n")
	sb.WriteString("  {\n")
	sb.WriteString("    \"topic\": \"Personajes|Lore|Geografía|Trama|Decisiones\",\n")
	sb.WriteString("    \"concept\": \"Nombre del personaje, lugar u objeto\",\n")
	sb.WriteString("    \"fact\": \"Descripción concisa del hecho o cambio\",\n")
	sb.WriteString("    \"type\": \"character|lore|setting|plot|decision\",\n")
	sb.WriteString("    \"tags\": [\"etiqueta1\", \"etiqueta2\"]\n")
	sb.WriteString("  }\n")
	sb.WriteString("]\n")
	sb.WriteString("```\n")
	sb.WriteString("No agregues texto explicativo fuera del bloque JSON.")

	return sb.String()
}

// ParseExtractedFacts parses the LLM output containing a JSON array of facts.
func (s *BrainService) ParseExtractedFacts(rawResponse string, chapterID string) ([]domain.BrainFact, error) {
	cleaned := strings.TrimSpace(rawResponse)
	if cleaned == "" {
		return []domain.BrainFact{}, nil
	}

	// Try extracting from ```json codeblock
	matches := jsonBlockRegex.FindStringSubmatch(cleaned)
	if len(matches) > 1 {
		cleaned = strings.TrimSpace(matches[1])
	} else {
		// Try finding first '[' and last ']'
		start := strings.Index(cleaned, "[")
		end := strings.LastIndex(cleaned, "]")
		if start != -1 && end != -1 && end > start {
			cleaned = cleaned[start : end+1]
		}
	}

	if cleaned == "" || cleaned == "[]" {
		return []domain.BrainFact{}, nil
	}

	type rawFact struct {
		Topic   string   `json:"topic"`
		Concept string   `json:"concept"`
		Fact    string   `json:"fact"`
		Type    string   `json:"type"`
		Tags    []string `json:"tags"`
	}

	var rawList []rawFact
	if err := json.Unmarshal([]byte(cleaned), &rawList); err != nil {
		return nil, fmt.Errorf("failed to parse extracted facts JSON: %w", err)
	}

	var facts []domain.BrainFact
	now := time.Now()

	for _, item := range rawList {
		if strings.TrimSpace(item.Fact) == "" || strings.TrimSpace(item.Concept) == "" {
			continue
		}

		fType := domain.BrainFactType(strings.ToLower(strings.TrimSpace(item.Type)))
		if fType == "" {
			fType = domain.FactTypeGeneral
		}

		topic := strings.TrimSpace(item.Topic)
		if topic == "" {
			topic = "General"
		}

		facts = append(facts, domain.BrainFact{
			ID:        fmt.Sprintf("fact_%d_%s", now.UnixNano(), sanitizeConcept(item.Concept)),
			Topic:     topic,
			Concept:   strings.TrimSpace(item.Concept),
			Fact:      strings.TrimSpace(item.Fact),
			Type:      fType,
			ChapterID: chapterID,
			Tags:      item.Tags,
			CreatedAt: now,
		})
	}

	return facts, nil
}

// RecordAutoLearnedFacts parses raw LLM output and persists new facts into the repository.
func (s *BrainService) RecordAutoLearnedFacts(ctx context.Context, rawResponse string, chapterID string) ([]domain.BrainFact, error) {
	facts, err := s.ParseExtractedFacts(rawResponse, chapterID)
	if err != nil {
		return nil, err
	}
	if len(facts) == 0 {
		return nil, nil
	}

	if s.repo != nil {
		if err := s.repo.SaveFacts(ctx, facts); err != nil {
			return nil, fmt.Errorf("failed to persist auto-learned facts: %w", err)
		}
	}

	return facts, nil
}

// BuildSessionSummaryPrompt constructs prompt for session summarization.
func (s *BrainService) BuildSessionSummaryPrompt(messages []domain.ChatMessage, chapterTitle string) string {
	var sb strings.Builder
	sb.WriteString("Eres el módulo de memoria 'Brain'. Genera un resumen conciso de la sesión de escritura.\n\n")
	if chapterTitle != "" {
		sb.WriteString(fmt.Sprintf("Capítulo: %s\n\n", chapterTitle))
	}
	sb.WriteString("HISTORIAL DE INTERACCIÓN:\n")
	for _, m := range messages {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", m.Role, m.Content))
	}
	sb.WriteString("\nResponde ÚNICAMENTE con un JSON en este formato:\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"summary\": \"Resumen en 1 o 2 oraciones del trabajo realizado.\",\n")
	sb.WriteString("  \"highlights\": [\"Punto clave 1\", \"Punto clave 2\"]\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n")

	return sb.String()
}

// ParseSessionSummary parses LLM JSON response for session summaries.
func (s *BrainService) ParseSessionSummary(rawResponse string) (*domain.SessionSummary, error) {
	cleaned := strings.TrimSpace(rawResponse)
	matches := jsonBlockRegex.FindStringSubmatch(cleaned)
	if len(matches) > 1 {
		cleaned = strings.TrimSpace(matches[1])
	} else {
		start := strings.Index(cleaned, "{")
		end := strings.LastIndex(cleaned, "}")
		if start != -1 && end != -1 && end > start {
			cleaned = cleaned[start : end+1]
		}
	}

	var parsed struct {
		Summary    string   `json:"summary"`
		Highlights []string `json:"highlights"`
	}

	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse session summary JSON: %w", err)
	}

	return &domain.SessionSummary{
		ID:         fmt.Sprintf("session_%d", time.Now().UnixNano()),
		Summary:    parsed.Summary,
		Highlights: parsed.Highlights,
		CreatedAt:  time.Now(),
	}, nil
}

// FormatFactsForPrompt renders Brain facts into a Markdown block for system prompt injection.
func (s *BrainService) FormatFactsForPrompt(facts []domain.BrainFact) string {
	if len(facts) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("--- MEMORIA DE CONTINUIDAD Y HECHOS (BRAIN) ---\n")
	for _, f := range facts {
		sb.WriteString(fmt.Sprintf("• [%s] %s: %s\n", f.Topic, f.Concept, f.Fact))
	}
	return sb.String()
}

// BuildTimelineExtractionPrompt constructs prompt for extracting chronological plot events.
func (s *BrainService) BuildTimelineExtractionPrompt(content string, chapterTitle string) string {
	var sb strings.Builder
	sb.WriteString("Eres el módulo de memoria y cronología narrativa 'Brain' de Novel-TUI.\n")
	sb.WriteString("Analiza el siguiente texto de la novela y extrae los eventos clave ordenados en una línea temporal (Cronología de la Trama).\n\n")
	if chapterTitle != "" {
		sb.WriteString(fmt.Sprintf("Capítulo de referencia: %s\n\n", chapterTitle))
	}
	sb.WriteString("TEXTO A ANALIZAR:\n")
	sb.WriteString(content)
	sb.WriteString("\n\n")
	sb.WriteString("INSTRUCCIONES DE RESPUESTA:\n")
	sb.WriteString("1. Si NO hay eventos cronológicos identificables, responde con una lista JSON vacía: []\n")
	sb.WriteString("2. Si encuentras eventos, responde EXCLUSIVAMENTE con un arreglo JSON en este formato:\n")
	sb.WriteString("```json\n")
	sb.WriteString("[\n")
	sb.WriteString("  {\n")
	sb.WriteString("    \"chronological_order\": 1,\n")
	sb.WriteString("    \"period\": \"Era Antigua|Año 452|Capítulo 1|Flashback|Presente\",\n")
	sb.WriteString("    \"title\": \"Título corto del evento\",\n")
	sb.WriteString("    \"description\": \"Resumen conciso del evento narrativo\",\n")
	sb.WriteString("    \"characters\": [\"Personaje 1\", \"Personaje 2\"]\n")
	sb.WriteString("  }\n")
	sb.WriteString("]\n")
	sb.WriteString("```\n")
	sb.WriteString("No agregues texto explicativo fuera del bloque JSON.")

	return sb.String()
}

// ParseExtractedTimelineEvents parses the LLM output containing a JSON array of timeline events.
func (s *BrainService) ParseExtractedTimelineEvents(rawResponse string, chapterID string) ([]domain.TimelineEvent, error) {
	cleaned := strings.TrimSpace(rawResponse)
	if cleaned == "" {
		return []domain.TimelineEvent{}, nil
	}

	matches := jsonBlockRegex.FindStringSubmatch(cleaned)
	if len(matches) > 1 {
		cleaned = strings.TrimSpace(matches[1])
	} else {
		start := strings.Index(cleaned, "[")
		end := strings.LastIndex(cleaned, "]")
		if start != -1 && end != -1 && end > start {
			cleaned = cleaned[start : end+1]
		}
	}

	if cleaned == "" || cleaned == "[]" {
		return []domain.TimelineEvent{}, nil
	}

	type rawTimelineEvent struct {
		ChronologicalOrder int      `json:"chronological_order"`
		Period             string   `json:"period"`
		Title              string   `json:"title"`
		Description        string   `json:"description"`
		Characters         []string `json:"characters"`
	}

	var rawList []rawTimelineEvent
	if err := json.Unmarshal([]byte(cleaned), &rawList); err != nil {
		return nil, fmt.Errorf("failed to parse extracted timeline events JSON: %w", err)
	}

	var events []domain.TimelineEvent
	now := time.Now()

	for i, item := range rawList {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			continue
		}

		order := item.ChronologicalOrder
		if order <= 0 {
			order = i + 1
		}

		period := strings.TrimSpace(item.Period)
		if period == "" {
			period = "Presente"
		}

		events = append(events, domain.TimelineEvent{
			ID:                 fmt.Sprintf("tl_%d_%s", now.UnixNano()+int64(i), sanitizeConcept(title)),
			ChronologicalOrder: order,
			Period:             period,
			Title:              title,
			Description:        strings.TrimSpace(item.Description),
			Characters:         item.Characters,
			ChapterID:          chapterID,
			CreatedAt:          now,
		})
	}

	return events, nil
}

// RecordAutoLearnedTimelineEvents parses raw LLM output and persists timeline events into the repository.
func (s *BrainService) RecordAutoLearnedTimelineEvents(ctx context.Context, rawResponse string, chapterID string) ([]domain.TimelineEvent, error) {
	events, err := s.ParseExtractedTimelineEvents(rawResponse, chapterID)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}

	if s.repo != nil {
		if err := s.repo.SaveTimelineEvents(ctx, events); err != nil {
			return nil, fmt.Errorf("failed to persist auto-learned timeline events: %w", err)
		}
	}

	return events, nil
}

// FormatTimelineForPrompt renders timeline events into a Markdown block for system prompt injection.
func (s *BrainService) FormatTimelineForPrompt(events []domain.TimelineEvent) string {
	if len(events) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("--- CRONOLOGÍA DE LA TRAMA (TIMELINE) ---\n")
	for _, e := range events {
		charStr := ""
		if len(e.Characters) > 0 {
			charStr = fmt.Sprintf(" (Personajes: %s)", strings.Join(e.Characters, ", "))
		}
		periodStr := ""
		if e.Period != "" {
			periodStr = fmt.Sprintf("[%s] ", e.Period)
		}
		sb.WriteString(fmt.Sprintf("%d. %s%s: %s%s\n", e.ChronologicalOrder, periodStr, e.Title, e.Description, charStr))
	}
	return sb.String()
}

func sanitizeConcept(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "entity"
	}
	return b.String()
}

func extractKeywords(text string, maxWords int) []string {
	stopWords := map[string]bool{
		"el": true, "la": true, "los": true, "las": true, "un": true, "una": true, "unos": true, "unas": true,
		"de": true, "del": true, "a": true, "al": true, "en": true, "con": true, "por": true, "para": true,
		"y": true, "o": true, "u": true, "que": true, "se": true, "su": true, "sus": true, "es": true,
		"son": true, "como": true, "este": true, "esta": true, "estos": true, "estas": true, "no": true,
		"si": true, "lo": true, "pero": true, "más": true, "fue": true, "era": true, "había": true,
	}

	words := strings.Fields(text)
	var keywords []string
	seen := make(map[string]bool)

	for _, w := range words {
		clean := strings.ToLower(strings.Trim(w, ".,;:!?'\"()[]{}—–-"))
		if len(clean) > 3 && !stopWords[clean] && !seen[clean] {
			seen[clean] = true
			keywords = append(keywords, clean)
			if len(keywords) >= maxWords {
				break
			}
		}
	}

	return keywords
}
