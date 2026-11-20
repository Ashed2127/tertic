package groq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Ashed2127/tertic/internal/openrouter"
)

const (
	DefaultModel = "llama-3.3-70b-versatile"
	APIBaseURL   = "https://api.groq.com/openai/v1/chat/completions"
)

// Client handles requests to Groq Cloud (Free tier)
type Client struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

func NewClient(apiKey string, model string) *Client {
	if model == "" {
		model = DefaultModel
	}
	return &Client{
		APIKey: apiKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) Chat(req openrouter.ChatCompletionRequest) (*openrouter.ChatCompletionResponse, error) {
	req.Model = c.Model

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", APIBaseURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to build http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("groq http request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read groq response body: %w", err)
	}

	var chatResp openrouter.ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse groq response: %w, raw: %s", err, string(bodyBytes))
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("Groq API error: %s (code %d)", chatResp.Error.Message, chatResp.Error.Code)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from Groq API")
	}

	return &chatResp, nil
}
