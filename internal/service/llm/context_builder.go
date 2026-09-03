package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/repository"
)

const (
	DefaultMaxChapterChars = 12000
	TruncationNotice       = "[...contenido anterior truncado por límites de contexto...]"
)

// ContextBuilder compiles lore, notes, genre guidelines, and active chapter text into a system prompt.
type ContextBuilder struct {
	MaxChapterChars int
}

// NewContextBuilder creates a new ContextBuilder with default truncation limits.
func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{
		MaxChapterChars: DefaultMaxChapterChars,
	}
}

// ContextParams encapsulates all available project data for context generation.
type ContextParams struct {
	NovelDir           string
	Genre              string
	GenrePrompt        string
	ActiveChapterTitle string
	ActiveChapterText  string
	EffortLevel        domain.LLMEffortLevel
	BrainFacts         []domain.BrainFact
}

// BuildContext compiles the full system prompt from workspace data.
func (cb *ContextBuilder) BuildContext(params ContextParams) string {
	var sb strings.Builder

	// 1. Base Persona and Genre
	if strings.TrimSpace(params.GenrePrompt) != "" {
		sb.WriteString(strings.TrimSpace(params.GenrePrompt))
		sb.WriteString("\n\n")
	} else if strings.TrimSpace(params.Genre) != "" {
		sb.WriteString(fmt.Sprintf("Eres un asistente de escritura para novelas del género %s.", params.Genre))
		sb.WriteString("\n\n")
	} else {
		sb.WriteString("Eres un asistente de escritura creativa para autores de novelas.")
		sb.WriteString("\n\n")
	}

	// 2. Effort Level Directive
	sb.WriteString(cb.getEffortInstruction(params.EffortLevel))
	sb.WriteString("\n\n")

	// 3. Brain Memory & Continuity Block (FTS5 / Indexed Facts)
	if len(params.BrainFacts) > 0 {
		sb.WriteString("--- MEMORIA DE CONTINUIDAD Y HECHOS (BRAIN) ---\n")
		for _, f := range params.BrainFacts {
			sb.WriteString(fmt.Sprintf("• [%s] %s: %s\n", f.Topic, f.Concept, f.Fact))
		}
		sb.WriteString("\n\n")
	}

	// 4. Characters / Lore Block
	if params.NovelDir != "" {
		charsBlock := cb.loadCharacterLore(params.NovelDir)
		if charsBlock != "" {
			sb.WriteString("--- PERSONAJES Y LORE ---\n")
			sb.WriteString(charsBlock)
			sb.WriteString("\n\n")
		}
	}

	// 4. Author Notes Block
	if params.NovelDir != "" {
		notesBlock := cb.loadAuthorNotes(params.NovelDir)
		if notesBlock != "" {
			sb.WriteString("--- NOTAS DEL AUTOR ---\n")
			sb.WriteString(notesBlock)
			sb.WriteString("\n\n")
		}
	}

	// 5. Active Chapter Context
	if strings.TrimSpace(params.ActiveChapterText) != "" {
		chapterTitle := params.ActiveChapterTitle
		if chapterTitle == "" {
			chapterTitle = "Capítulo Actual"
		}

		chapterText := params.ActiveChapterText
		maxChars := cb.MaxChapterChars
		if maxChars <= 0 {
			maxChars = DefaultMaxChapterChars
		}

		runes := []rune(chapterText)
		if len(runes) > maxChars {
			truncatedRunes := runes[len(runes)-maxChars:]
			chapterText = TruncationNotice + "\n" + string(truncatedRunes)
		}

		sb.WriteString(fmt.Sprintf("--- INICIO CAPÍTULO ACTUAL (%s) ---\n", chapterTitle))
		sb.WriteString(strings.TrimSpace(chapterText))
		sb.WriteString("\n--- FIN CAPÍTULO ACTUAL ---")
	}

	return strings.TrimSpace(sb.String())
}

func (cb *ContextBuilder) getEffortInstruction(effort domain.LLMEffortLevel) string {
	switch effort {
	case domain.EffortLow:
		return "Sé conciso, directo y breve. Responde en 1-3 frases máximo."
	case domain.EffortHigh:
		return "Actúa como editor literario senior. Analiza estructura narrativa, arcos de personajes, subtramas y coherencia causal paso a paso antes de dar recomendaciones detalladas."
	case domain.EffortMedium:
		fallthrough
	default:
		return "Actúa como co-escritor creativo. Sugiere mejoras estilísticas y extensiones de prosa manteniendo la voz del autor."
	}
}

func (cb *ContextBuilder) loadCharacterLore(novelDir string) string {
	dir := repository.ExpandHome(novelDir)
	path := filepath.Join(dir, "personajes.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var characters []domain.Character
	if err := json.Unmarshal(data, &characters); err != nil || len(characters) == 0 {
		return ""
	}

	var lines []string
	for _, c := range characters {
		if strings.TrimSpace(c.Name) == "" {
			continue
		}
		desc := ""
		if c.Role != "" {
			desc += fmt.Sprintf(" (Rol: %s)", c.Role)
		}
		if c.Description != "" {
			desc += fmt.Sprintf(" - %s", c.Description)
		}
		if c.Notes != "" {
			desc += fmt.Sprintf(" [Notas: %s]", c.Notes)
		}
		lines = append(lines, fmt.Sprintf("• %s%s", c.Name, desc))
	}

	return strings.Join(lines, "\n")
}

func (cb *ContextBuilder) loadAuthorNotes(novelDir string) string {
	dir := repository.ExpandHome(novelDir)
	for _, name := range []string{"notas.txt", "notes.txt"} {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err == nil && len(strings.TrimSpace(string(data))) > 0 {
			return strings.TrimSpace(string(data))
		}
	}
	return ""
}
