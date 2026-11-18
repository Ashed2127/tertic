package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Ashed2127/tertic/internal/tools"
)

const (
	DefaultModel = "gemini-3.8-flash"
	APIBaseURL   = "https://generativelanguage.googleapis.com/v1beta/models"
)

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

// Request and Response Structs for Google Gemini API
type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type Part struct {
	Text             string            `json:"text,omitempty"`
	FunctionCall     *FunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
}

type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type FunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type SystemInstruction struct {
	Parts []Part `json:"parts"`
}

type ToolDeclarationWrapper struct {
	FunctionDeclarations []tools.ToolDefinition `json:"functionDeclarations"`
}

type GenerateContentRequest struct {
	SystemInstruction *SystemInstruction       `json:"systemInstruction,omitempty"`
	Contents          []Content                `json:"contents"`
	Tools             []ToolDeclarationWrapper `json:"tools,omitempty"`
}

type GenerateContentResponse struct {
	Candidates []struct {
		Content Content `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func (c *Client) Generate(req GenerateContentRequest) (*GenerateContentResponse, error) {
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", APIBaseURL, c.Model, c.APIKey)

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to build http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var genResp GenerateContentResponse
	if err := json.Unmarshal(bodyBytes, &genResp); err != nil {
		return nil, fmt.Errorf("failed to parse json response: %w, raw: %s", err, string(bodyBytes))
	}

	if genResp.Error != nil {
		return nil, fmt.Errorf("Gemini API error (%d - %s): %s", genResp.Error.Code, genResp.Error.Status, genResp.Error.Message)
	}

	if len(genResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned from Gemini API")
	}

	return &genResp, nil
}
