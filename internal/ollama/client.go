package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	"github.com/Ashed2127/tertic/internal/openrouter"
)

const DefaultHost = "http://localhost:11434"

type LocalModelInfo struct {
	ID          string
	Name        string
	Size        string
	Description string
}

var AvailableLocalModels = []LocalModelInfo{
	{ID: "qwen2.5-coder:0.5b", Name: "Qwen 2.5 Coder 0.5B", Size: "397MB", Description: "Ultra-lightweight, fastest on any phone"},
	{ID: "deepseek-coder:1.3b", Name: "DeepSeek Coder 1.3B", Size: "776MB", Description: "Fast code synthesis, low RAM"},
	{ID: "qwen2.5-coder:1.5b", Name: "Qwen 2.5 Coder 1.5B", Size: "986MB", Description: "Balanced speed & accuracy"},
	{ID: "deepseek-r1:1.5b", Name: "DeepSeek R1 1.5B", Size: "1.1GB", Description: "Local reasoning & step-by-step logic"},
	{ID: "gemma:2b", Name: "Google Gemma 2B", Size: "1.7GB", Description: "Google lightweight instruction model"},
	{ID: "qwen2.5-coder:3b", Name: "Qwen 2.5 Coder 3B", Size: "1.9GB", Description: "High accuracy, best for 6GB+ RAM"},
}

type Client struct {
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

func NewClient(baseURL, model string) *Client {
	if baseURL == "" {
		baseURL = DefaultHost
	}
	return &Client{
		BaseURL: baseURL,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

// GetDownloadedModels queries the local Ollama daemon for currently installed models
func GetDownloadedModels(baseURL string) ([]string, error) {
	if baseURL == "" {
		baseURL = DefaultHost
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(baseURL + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Models []struct {
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var names []string
	for _, m := range data.Models {
		sizeMB := float64(m.Size) / (1024 * 1024)
		if sizeMB >= 1000 {
			names = append(names, fmt.Sprintf("%-30s (%.1f GB)", m.Name, sizeMB/1024))
		} else {
			names = append(names, fmt.Sprintf("%-30s (%.0f MB)", m.Name, sizeMB))
		}
	}
	return names, nil
}

// IsOllamaRunning checks if the Ollama daemon is reachable
func (c *Client) IsOllamaRunning() bool {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// EnsureOllamaRunning tries to start ollama serve in background if not running
func EnsureOllamaRunning() {
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(DefaultHost + "/api/tags")
	if err == nil && resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		return
	}

	if _, err := exec.LookPath("ollama"); err == nil {
		cmd := exec.Command("ollama", "serve")
		_ = cmd.Start()
		time.Sleep(2 * time.Second)
	}
}

// PullModel executes `ollama pull <model>`
func PullModel(model string) error {
	cmd := exec.Command("ollama", "pull", model)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// Chat uses the OpenAI-compatible endpoint of Ollama (/v1/chat/completions)
func (c *Client) Chat(req openrouter.ChatCompletionRequest) (*openrouter.ChatCompletionResponse, error) {
	req.Model = c.Model

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.BaseURL + "/v1/chat/completions"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to build http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Ollama connection error (make sure 'ollama serve' is running): %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp openrouter.ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse json response: %w, raw: %s", err, string(bodyBytes))
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("Ollama error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response returned from local model")
	}

	return &chatResp, nil
}
