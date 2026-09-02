package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

// OllamaProvider implements domain.LLMProvider for local Ollama instances.
type OllamaProvider struct {
	BaseURL string
	Client  *http.Client
}

// NewOllamaProvider creates a new OllamaProvider.
func NewOllamaProvider(baseURL string, client *http.Client) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if client == nil {
		client = &http.Client{}
	}
	return &OllamaProvider{
		BaseURL: baseURL,
		Client:  client,
	}
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  *ollamaOptions  `json:"options,omitempty"`
}

type ollamaChatResponse struct {
	Model   string        `json:"model"`
	Message ollamaMessage `json:"message"`
	Done    bool          `json:"done"`
	Error   string        `json:"error,omitempty"`
}

// StreamChat sends a chat request to Ollama and streams response chunks.
func (p *OllamaProvider) StreamChat(ctx context.Context, req domain.ChatRequest) (<-chan domain.StreamChunk, error) {
	endpoint := fmt.Sprintf("%s/api/chat", strings.TrimRight(p.BaseURL, "/"))

	messages := make([]ollamaMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = ollamaMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	bodyData := ollamaChatRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   true,
	}
	if req.Temperature > 0 {
		bodyData.Options = &ollamaOptions{
			Temperature: req.Temperature,
		}
	}

	payload, err := json.Marshal(bodyData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama connection failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		respBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned error status %d: %s", resp.StatusCode, string(respBytes))
	}

	outChan := make(chan domain.StreamChunk, 16)

	go func() {
		defer resp.Body.Close()
		defer close(outChan)

		scanner := bufio.NewScanner(resp.Body)
		// Support larger buffer for long lines
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Bytes()
			if len(bytes.TrimSpace(line)) == 0 {
				continue
			}

			var respObj ollamaChatResponse
			if err := json.Unmarshal(line, &respObj); err != nil {
				select {
				case outChan <- domain.StreamChunk{Err: fmt.Errorf("failed to decode json chunk: %w", err), Done: true}:
				case <-ctx.Done():
				}
				return
			}

			if respObj.Error != "" {
				select {
				case outChan <- domain.StreamChunk{Err: fmt.Errorf("ollama error: %s", respObj.Error), Done: true}:
				case <-ctx.Done():
				}
				return
			}

			chunk := domain.StreamChunk{
				Content: respObj.Message.Content,
				Done:    respObj.Done,
			}

			select {
			case outChan <- chunk:
			case <-ctx.Done():
				return
			}

			if respObj.Done {
				return
			}
		}

		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			select {
			case outChan <- domain.StreamChunk{Err: fmt.Errorf("stream read error: %w", err), Done: true}:
			case <-ctx.Done():
			}
		}
	}()

	return outChan, nil
}
