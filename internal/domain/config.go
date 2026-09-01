package domain

import (
	"os"
	"path/filepath"
	"time"
)

// LLMConfig holds settings for local and remote LLM integrations.
type LLMConfig struct {
	Provider     string            `json:"provider"`
	BaseURL      string            `json:"base_url"`
	Model        string            `json:"model"`
	Temperature  float64           `json:"temperature"`
	GenrePrompts map[string]string `json:"genre_prompts"`
}

// DefaultLLMConfig returns standard defaults for LLM integration.
func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Provider:    "ollama",
		BaseURL:     "http://localhost:11434",
		Model:       "qwen2.5:7b",
		Temperature: 0.7,
		GenrePrompts: map[string]string{
			"Fantasy": "Eres un asistente de escritura para novelas de fantasía épica. Enfócate en la construcción de mundo, descripciones mágicas y arcos míticos.",
			"Sci-Fi":  "Eres un asistente de escritura para novelas de ciencia ficción. Enfócate en consistencia tecnológica, ambientación futurista y dilemas filosóficos.",
			"Mystery": "Eres un asistente de escritura para novelas de misterio y suspense. Enfócate en pistas sutiles, ritmo tenso y giros argumentales.",
			"Romance": "Eres un asistente de escritura para novelas de romance. Enfócate en la química de los personajes, tensión emocional y desarrollo de relaciones.",
		},
	}
}

// AppConfig represents global user and workspace settings.
type AppConfig struct {
	RootDir      string    `json:"root_dir"`
	RecentNovels []string  `json:"recent_novels"`
	LLM          LLMConfig `json:"llm"`
}

// DefaultAppConfig creates an AppConfig with default root directory (~/Novelas) and default LLM configuration.
func DefaultAppConfig() *AppConfig {
	homeDir, err := os.UserHomeDir()
	rootDir := "~/Novelas"
	if err == nil && homeDir != "" {
		rootDir = filepath.Join(homeDir, "Novelas")
	}
	return &AppConfig{
		RootDir:      rootDir,
		RecentNovels: []string{},
		LLM:          DefaultLLMConfig(),
	}
}

// NovelMetadata encapsulates metadata about a novel workspace project.
type NovelMetadata struct {
	Title        string    `json:"title"`
	AbsolutePath string    `json:"absolute_path"`
	ChapterCount int       `json:"chapter_count"`
	LastModified time.Time `json:"last_modified"`
}
