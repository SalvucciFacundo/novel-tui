package llm

import (
	"os"
	"strings"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

// NewProvider creates an instance of domain.LLMProvider based on the given LLMConfig.
func NewProvider(cfg domain.LLMConfig) (domain.LLMProvider, error) {
	providerType := strings.ToLower(strings.TrimSpace(cfg.Provider))

	switch providerType {
	case "openai", "openai-compatible", "openrouter", "vllm":
		apiKey := os.Getenv("OPENAI_API_KEY")
		return NewOpenAICompatibleProvider(cfg.BaseURL, apiKey, nil), nil
	case "ollama":
		fallthrough
	default:
		return NewOllamaProvider(cfg.BaseURL, nil), nil
	}
}
