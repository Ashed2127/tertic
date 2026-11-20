package mistral

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
	DefaultModel = "mistral-small-latest"
	APIBaseURL   = "https://api.mistral.ai/v1/chat/completions"
)

// Client handles requests to Mistral AI (Free tier / La Plateforme)
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
		return nil, fmt.Errorf("mistral http request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read mistral response body: %w", err)
	}

	var chatResp openrouter.ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse mistral response: %w, raw: %s", err, string(bodyBytes))
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("Mistral API error: %s (code %d)", chatResp.Error.Message, chatResp.Error.Code)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from Mistral API")
	}

	return &chatResp, nil
}
