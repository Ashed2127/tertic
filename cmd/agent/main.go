package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ashed2127/tertic/internal/gemini"
	"github.com/Ashed2127/tertic/internal/openrouter"
	"github.com/Ashed2127/tertic/internal/tools"
)

// ANSI colors for clean mobile terminal UI
const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
)

const Banner = `
  _____         _   _      
 |_   _|__ _ __| |_(_) ___ 
   | |/ _ \ '__| __| |/ __|
   | |  __/ |  | |_| | (__ 
   |_|\___|_|   \__|_|\___|
 Autonomous AI Coding Agent for Termux (OpenRouter & Gemini)
`

const SystemPrompt = `You are Tertic, an autonomous AI software engineering agent running directly inside Termux on Android.
Your job is to write, debug, test, and deliver working full-stack applications.

Tools available:
1. execute_bash: Run shell commands, package managers (npm, pip, go, git), linters, build tools, dev servers.
2. write_file: Create or overwrite files.
3. read_file: Inspect existing code and logs.
4. edit_file: Exact search and replace to modify parts of files.
5. list_dir: Explore project directories.

Rules:
- Be autonomous: Do the work directly. Create the folders, write the code, install dependencies, run tests, and fix any errors.
- Always be fast and concise.
- When done, summarize what you built and how the user can test/run it.`

type AgentConfig struct {
	Provider string // "openrouter" or "gemini"
	APIKey   string
	Model    string
}

// resolveConfig checks environment variables, ~/.tertic/config.env, or prompts the user interactively
func resolveConfig() (*AgentConfig, error) {
	// 1. Direct env vars
	if key := os.Getenv("OPENROUTER_API_KEY"); strings.TrimSpace(key) != "" {
		model := os.Getenv("OPENROUTER_MODEL")
		if model == "" {
			model = openrouter.DefaultModel
		}
		return &AgentConfig{Provider: "openrouter", APIKey: strings.TrimSpace(key), Model: model}, nil
	}

	if key := os.Getenv("GEMINI_API_KEY"); strings.TrimSpace(key) != "" {
		model := os.Getenv("GEMINI_MODEL")
		if model == "" {
			model = gemini.DefaultModel
		}
		return &AgentConfig{Provider: "gemini", APIKey: strings.TrimSpace(key), Model: model}, nil
	}

	// 2. Saved Tertic config (~/.tertic/config.env)
	home, _ := os.UserHomeDir()
	if home != "" {
		configFile := filepath.Join(home, ".tertic", "config.env")
		if data, err := os.ReadFile(configFile); err == nil {
			var orKey, gemKey, orModel, gemModel, savedProvider string
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "export TERTIC_PROVIDER=") {
					savedProvider = strings.Trim(strings.TrimPrefix(line, "export TERTIC_PROVIDER="), `"' `)
				}
				if strings.HasPrefix(line, "export OPENROUTER_API_KEY=") {
					orKey = strings.Trim(strings.TrimPrefix(line, "export OPENROUTER_API_KEY="), `"' `)
				}
				if strings.HasPrefix(line, "export OPENROUTER_MODEL=") {
					orModel = strings.Trim(strings.TrimPrefix(line, "export OPENROUTER_MODEL="), `"' `)
				}
				if strings.HasPrefix(line, "export GEMINI_API_KEY=") {
					gemKey = strings.Trim(strings.TrimPrefix(line, "export GEMINI_API_KEY="), `"' `)
				}
				if strings.HasPrefix(line, "export GEMINI_MODEL=") {
					gemModel = strings.Trim(strings.TrimPrefix(line, "export GEMINI_MODEL="), `"' `)
				}
			}

			if savedProvider == "openrouter" && orKey != "" {
				if orModel == "" {
					orModel = openrouter.DefaultModel
				}
				return &AgentConfig{Provider: "openrouter", APIKey: orKey, Model: orModel}, nil
			}
			if savedProvider == "gemini" && gemKey != "" {
				if gemModel == "" {
					gemModel = gemini.DefaultModel
				}
				return &AgentConfig{Provider: "gemini", APIKey: gemKey, Model: gemModel}, nil
			}
			if orKey != "" {
				if orModel == "" {
					orModel = openrouter.DefaultModel
				}
				return &AgentConfig{Provider: "openrouter", APIKey: orKey, Model: orModel}, nil
			}
			if gemKey != "" {
				if gemModel == "" {
					gemModel = gemini.DefaultModel
				}
				return &AgentConfig{Provider: "gemini", APIKey: gemKey, Model: gemModel}, nil
			}
		}
	}

	// 3. Interactive Setup Menu
	fmt.Printf("%s%s[Tertic Setup] Choose an AI Provider:%s\n\n", ColorBold, ColorCyan, ColorReset)
	fmt.Printf("  %s1) OpenRouter%s (Recommended: 100%% Free Models: Llama 3.3, Qwen Coder, DeepSeek)\n", ColorGreen, ColorReset)
	fmt.Printf("  %s2) Google Gemini Studio%s (Gemini Flash-Lite)\n\n", ColorYellow, ColorReset)
	fmt.Print("Select [1 or 2, default: 1]: ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	provider := "openrouter"
	var keyPrompt, keyURL, defaultModel string

	if choice == "2" {
		provider = "gemini"
		keyURL = "https://aistudio.google.com/app/apikey"
		keyPrompt = "Google Gemini API Key"
		defaultModel = gemini.DefaultModel
	} else {
		provider = "openrouter"
		keyURL = "https://openrouter.ai/keys (Free account)"
		keyPrompt = "OpenRouter API Key (sk-or-...)"
		defaultModel = openrouter.DefaultModel
	}

	fmt.Printf("\nGet your free key at: %s%s%s\n", ColorCyan, keyURL, ColorReset)
	fmt.Printf("Paste your %s: ", keyPrompt)

	key, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %w", err)
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	// Save to config.env
	if home != "" {
		terticDir := filepath.Join(home, ".tertic")
		_ = os.MkdirAll(terticDir, 0755)
		configFile := filepath.Join(terticDir, "config.env")

		f, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			_, _ = f.WriteString(fmt.Sprintf("\nexport TERTIC_PROVIDER=\"%s\"\n", provider))
			if provider == "openrouter" {
				_, _ = f.WriteString(fmt.Sprintf("export OPENROUTER_API_KEY=\"%s\"\n", key))
				_, _ = f.WriteString(fmt.Sprintf("export OPENROUTER_MODEL=\"%s\"\n", defaultModel))
			} else {
				_, _ = f.WriteString(fmt.Sprintf("export GEMINI_API_KEY=\"%s\"\n", key))
				_, _ = f.WriteString(fmt.Sprintf("export GEMINI_MODEL=\"%s\"\n", defaultModel))
			}
			_ = f.Close()
			fmt.Printf("%s✓ Settings saved to %s%s\n\n", ColorGreen, configFile, ColorReset)
		}
	}

	return &AgentConfig{
		Provider: provider,
		APIKey:   key,
		Model:    defaultModel,
	}, nil
}

// --- OPENROUTER TURN ---
func runOpenRouterTurn(client *openrouter.Client, history *[]openrouter.Message, userPrompt string) error {
	*history = append(*history, openrouter.Message{
		Role:    "user",
		Content: userPrompt,
	})

	toolDefs := tools.GetToolDefinitions()
	openRouterTools := openrouter.ConvertToolDefinitions(toolDefs)

	maxSteps := 30
	for step := 1; step <= maxSteps; step++ {
		req := openrouter.ChatCompletionRequest{
			Messages: *history,
			Tools:    openRouterTools,
		}

		resp, err := client.Chat(req)
		if err != nil {
			return err
		}

		choice := resp.Choices[0]
		assistantMsg := choice.Message
		*history = append(*history, assistantMsg)

		if assistantMsg.Content != "" {
			fmt.Printf("%s%s%s\n", ColorReset, assistantMsg.Content, ColorReset)
		}

		if len(assistantMsg.ToolCalls) == 0 {
			fmt.Printf("\n%s%s✓ Task completed!%s\n", ColorBold, ColorGreen, ColorReset)
			break
		}

		// Execute tools
		for _, tc := range assistantMsg.ToolCalls {
			fn := tc.Function
			fmt.Printf("\n%s%s⚡ [%s]:%s %s\n", ColorBold, ColorYellow, fn.Name, ColorReset, fn.Arguments)

			var args map[string]interface{}
			_ = json.Unmarshal([]byte(fn.Arguments), &args)

			output, toolErr := tools.ExecuteTool(fn.Name, args)
			if toolErr != nil {
				output = fmt.Sprintf("Error: %v", toolErr)
				fmt.Printf("%s✗ Error: %v%s\n", ColorRed, toolErr, ColorReset)
			} else {
				lines := strings.Split(output, "\n")
				preview := output
				if len(lines) > 5 {
					preview = strings.Join(lines[:5], "\n") + fmt.Sprintf("\n... (+%d more lines)", len(lines)-5)
				}
				fmt.Printf("%s%s%s\n", ColorCyan, preview, ColorReset)
			}

			*history = append(*history, openrouter.Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    output,
			})
		}
	}
	return nil
}

// --- GEMINI TURN ---
func runGeminiTurn(client *gemini.Client, history *[]gemini.Content, userPrompt string) error {
	*history = append(*history, gemini.Content{
		Role: "user",
		Parts: []gemini.Part{
			{Text: userPrompt},
		},
	})

	toolDefs := tools.GetToolDefinitions()
	req := gemini.GenerateContentRequest{
		SystemInstruction: &gemini.SystemInstruction{
			Parts: []gemini.Part{
				{Text: SystemPrompt},
			},
		},
		Tools: []gemini.ToolDeclarationWrapper{
			{FunctionDeclarations: toolDefs},
		},
	}

	maxSteps := 30
	for step := 1; step <= maxSteps; step++ {
		req.Contents = *history

		resp, err := client.Generate(req)
		if err != nil {
			return err
		}

		candidate := resp.Candidates[0]
		*history = append(*history, candidate.Content)

		hasToolCalls := false
		var toolResponseParts []gemini.Part

		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				fmt.Printf("%s%s%s\n", ColorReset, part.Text, ColorReset)
			}

			if part.FunctionCall != nil {
				hasToolCalls = true
				call := part.FunctionCall

				fmt.Printf("\n%s%s⚡ [%s]:%s %v\n", ColorBold, ColorYellow, call.Name, ColorReset, call.Args)

				output, toolErr := tools.ExecuteTool(call.Name, call.Args)
				if toolErr != nil {
					output = fmt.Sprintf("Error: %v", toolErr)
					fmt.Printf("%s✗ Error: %v%s\n", ColorRed, toolErr, ColorReset)
				} else {
					lines := strings.Split(output, "\n")
					preview := output
					if len(lines) > 5 {
						preview = strings.Join(lines[:5], "\n") + fmt.Sprintf("\n... (+%d more lines)", len(lines)-5)
					}
					fmt.Printf("%s%s%s\n", ColorCyan, preview, ColorReset)
				}

				toolResponseParts = append(toolResponseParts, gemini.Part{
					FunctionResponse: &gemini.FunctionResponse{
						Name: call.Name,
						Response: map[string]interface{}{
							"result": output,
						},
					},
				})
			}
		}

		if hasToolCalls {
			*history = append(*history, gemini.Content{
				Role:  "user",
				Parts: toolResponseParts,
			})
		} else {
			fmt.Printf("\n%s%s✓ Task completed!%s\n", ColorBold, ColorGreen, ColorReset)
			break
		}
	}
	return nil
}

func main() {
	cfg, err := resolveConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	var orClient *openrouter.Client
	var orHistory []openrouter.Message

	var gemClient *gemini.Client
	var gemHistory []gemini.Content

	if cfg.Provider == "openrouter" {
		orClient = openrouter.NewClient(cfg.APIKey, cfg.Model)
		orHistory = append(orHistory, openrouter.Message{
			Role:    "system",
			Content: SystemPrompt,
		})
	} else {
		gemClient = gemini.NewClient(cfg.APIKey, cfg.Model)
	}

	// Direct prompt execution (e.g. trtc "build an app")
	if len(os.Args) > 1 {
		userPrompt := strings.Join(os.Args[1:], " ")
		fmt.Printf("\n%s%s🚀 Tertic Agent (%s - %s):%s %s\n\n", ColorBold, ColorCyan, cfg.Provider, cfg.Model, ColorReset, userPrompt)
		if cfg.Provider == "openrouter" {
			if err := runOpenRouterTurn(orClient, &orHistory, userPrompt); err != nil {
				fmt.Fprintf(os.Stderr, "\n%sError: %v%s\n", ColorRed, err, ColorReset)
			}
		} else {
			if err := runGeminiTurn(gemClient, &gemHistory, userPrompt); err != nil {
				fmt.Fprintf(os.Stderr, "\n%sError: %v%s\n", ColorRed, err, ColorReset)
			}
		}
		return
	}

	// Interactive Mode (e.g. trtc)
	fmt.Print(ColorCyan, Banner, ColorReset)
	fmt.Printf("%sProvider:%s %s | %sModel:%s %s\n", ColorBold, ColorReset, cfg.Provider, ColorBold, ColorReset, cfg.Model)
	fmt.Printf("%sInteractive Mode%s (type your instructions or 'exit' to quit)\n\n", ColorBold, ColorReset)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s%strtc > %s", ColorBold, ColorGreen, ColorReset)
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		prompt := strings.TrimSpace(input)
		if prompt == "" {
			continue
		}
		if prompt == "exit" || prompt == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		if cfg.Provider == "openrouter" {
			if err := runOpenRouterTurn(orClient, &orHistory, prompt); err != nil {
				fmt.Fprintf(os.Stderr, "%sError: %v%s\n\n", ColorRed, err, ColorReset)
			}
		} else {
			if err := runGeminiTurn(gemClient, &gemHistory, prompt); err != nil {
				fmt.Fprintf(os.Stderr, "%sError: %v%s\n\n", ColorRed, err, ColorReset)
			}
		}
		fmt.Println()
	}
}
