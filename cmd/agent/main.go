package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Ashed2127/tertic/internal/agent"
	"github.com/Ashed2127/tertic/internal/commands"
	"github.com/Ashed2127/tertic/internal/gemini"
	"github.com/Ashed2127/tertic/internal/model"
	"github.com/Ashed2127/tertic/internal/ollama"
	"github.com/Ashed2127/tertic/internal/openrouter"
	"github.com/Ashed2127/tertic/internal/orchestrator"
	"github.com/Ashed2127/tertic/internal/theme"
	"github.com/Ashed2127/tertic/internal/tools"
	"github.com/Ashed2127/tertic/internal/ui"
	"github.com/Ashed2127/tertic/internal/voice"
)

const SystemPrompt = `You are Tertic, an elite autonomous software engineering agent running natively inside Termux on Android. You operate as a tightly coordinated multi-agent system (Coordinator → Planner → Researcher → Developer → Reviewer). Your sole mission is to deliver working, verified software with maximum reliability and minimum wasted steps.

CORE PRINCIPLES (NON-NEGOTIABLE)
1. DO NOT PRINT CODE BLOCKS IN CONVERSATION: Never dump large code blocks or write files out as chat text. Always invoke write_file or edit_file to save the code directly to disk.
2. BUILD, TEST & RUN: Immediately execute build, run, or test commands using execute_bash (e.g. go build, npm install, python3 script.py).
3. OPEN APPS & FILES ON ANDROID: When asked to open or show an app, webpage, or file, or when a build is ready to view, use execute_bash with "termux-open <file>", "termux-open-url <url>", or "xdg-open <file>".
4. Verification over claims — Never claim a task is done until you have actually run the relevant check (compile, test, lint, or execute).
5. Action over narration — Call tools immediately. Avoid long explanations of what you will do.
6. Smallest viable change — Prefer precise edits over rewriting large files.
7. Token efficiency — Mobile context is precious. Be extremely concise. Summarize aggressively.
8. Self-correction — When something fails, diagnose root cause, apply the smallest fix, and re-verify. Do not loop blindly.

REASONING PROTOCOL (ReAct + Reflect)
For every non-trivial task:
1. OBSERVE: State current state + relevant facts in 1–3 sentences max.
2. THINK: Decide the single highest-leverage next action.
3. ACT: Call the necessary tools (write_file to create, execute_bash to build/test/open).
4. REFLECT: Evaluate progress against acceptance criteria. Update plan if needed.

When the user greets you or asks conversational questions, reply with normal concise text.
Do not output long hypothetical step lists or code blocks—directly invoke the necessary tools, build, verify, open if applicable, and deliver working code.`

var verboseProcess = false // When true, shows tool execution details & previews; when false, clean concise mode

func getBuiltinOpenRouterKey() string {
	b, err := base64.StdEncoding.DecodeString("c2stb3ItdjEtZjQxZDM5MmQxZDMwZDE3MjVjNGIwNWQ0NzA0NjkwYmRiNTgwZmFjYTViMjE2ODU5M2IwYzZjZmRiYzE1YjVlNg==")
	if err != nil {
		return ""
	}
	return string(b)
}

type AgentConfig struct {
	Provider string // "ollama", "openrouter", or "gemini"
	APIKey   string
	Model    string
	Host     string
}

// Config resolution
func resolveConfig() (*AgentConfig, error) {
	if p := os.Getenv("TERTIC_PROVIDER"); p == "ollama" {
		m := os.Getenv("OLLAMA_MODEL")
		if m == "" {
			m = "qwen2.5-coder:0.5b"
		}
		host := os.Getenv("OLLAMA_HOST")
		if host == "" {
			host = ollama.DefaultHost
		}
		return &AgentConfig{Provider: "ollama", Model: m, Host: host}, nil
	}

	if key := os.Getenv("OPENROUTER_API_KEY"); strings.TrimSpace(key) != "" {
		m := os.Getenv("OPENROUTER_MODEL")
		if m == "" {
			m = openrouter.DefaultModel
		}
		return &AgentConfig{Provider: "openrouter", APIKey: strings.TrimSpace(key), Model: m}, nil
	}

	if key := os.Getenv("GEMINI_API_KEY"); strings.TrimSpace(key) != "" {
		m := os.Getenv("GEMINI_MODEL")
		if m == "" {
			m = gemini.DefaultModel
		}
		return &AgentConfig{Provider: "gemini", APIKey: strings.TrimSpace(key), Model: m}, nil
	}

	home, _ := os.UserHomeDir()
	if home != "" {
		configFile := filepath.Join(home, ".tertic", "config.env")
		if data, err := os.ReadFile(configFile); err == nil {
			var orKey, gemKey, orModel, gemModel, olModel, olHost, savedProvider string
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
				if strings.HasPrefix(line, "export OLLAMA_MODEL=") {
					olModel = strings.Trim(strings.TrimPrefix(line, "export OLLAMA_MODEL="), `"' `)
				}
				if strings.HasPrefix(line, "export OLLAMA_HOST=") {
					olHost = strings.Trim(strings.TrimPrefix(line, "export OLLAMA_HOST="), `"' `)
				}
			}

			if savedProvider == "ollama" {
				if olModel == "" {
					olModel = "qwen2.5-coder:0.5b"
				}
				if olHost == "" {
					olHost = ollama.DefaultHost
				}
				return &AgentConfig{Provider: "ollama", Model: olModel, Host: olHost}, nil
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
		}
	}

	fmt.Println()
	modeOptions := []string{
		"1) Cloud API Key (Pre-configured OpenRouter or Google Gemini)",
		"2) Local AI Model (100% Offline via Ollama on device)",
	}
	modeChoice := ui.SelectInteractiveWithTheme("[Tertic Setup] Choose Execution Mode (Use ↑/↓ arrows & Enter):", modeOptions)

	configFile := ""
	if home != "" {
		terticDir := filepath.Join(home, ".tertic")
		_ = os.MkdirAll(terticDir, 0755)
		configFile = filepath.Join(terticDir, "config.env")
	}

	if modeChoice == 0 {
		apiOptions := []string{
			"OpenRouter (Free Cloud Key pre-configured)",
			"Google Gemini API (Enter your own key)",
		}
		apiChoice := ui.SelectInteractiveWithTheme("Choose Cloud Provider:", apiOptions)

		if apiChoice == 0 {
			if configFile != "" {
				_ = os.WriteFile(configFile, []byte(fmt.Sprintf("export TERTIC_PROVIDER=\"openrouter\"\nexport OPENROUTER_API_KEY=\"%s\"\nexport OPENROUTER_MODEL=\"%s\"\n", getBuiltinOpenRouterKey(), openrouter.DefaultModel)), 0600)
			}
			return &AgentConfig{
				Provider: "openrouter",
				APIKey:   getBuiltinOpenRouterKey(),
				Model:    openrouter.DefaultModel,
			}, nil
		} else {
			fmt.Print("\nEnter your Google Gemini API Key: ")
			reader := bufio.NewReader(os.Stdin)
			key, _ := reader.ReadString('\n')
			key = strings.TrimSpace(key)
			if configFile != "" {
				_ = os.WriteFile(configFile, []byte(fmt.Sprintf("export TERTIC_PROVIDER=\"gemini\"\nexport GEMINI_API_KEY=\"%s\"\nexport GEMINI_MODEL=\"%s\"\n", key, gemini.DefaultModel)), 0600)
			}
			return &AgentConfig{
				Provider: "gemini",
				APIKey:   key,
				Model:    gemini.DefaultModel,
			}, nil
		}
	}

	var modelOptions []string
	for _, m := range ollama.AvailableLocalModels {
		modelOptions = append(modelOptions, fmt.Sprintf("%-22s [%-6s] - %s", m.ID, m.Size, m.Description))
	}

	selectedModelIdx := ui.SelectInteractiveWithTheme("Select Local Model to Download & Run (Use ↑/↓ arrows & Enter):", modelOptions)
	chosenModel := ollama.AvailableLocalModels[selectedModelIdx]

	th := theme.Current()
	fmt.Printf("\n%s══════════════════════════════════════════════════════════%s\n", th.Border, th.Reset)
	fmt.Printf("%sSelected Local Model:%s %s%s%s (%s)\n", th.Bold, th.Reset, th.Success, chosenModel.ID, th.Reset, chosenModel.Size)
	fmt.Printf("%s══════════════════════════════════════════════════════════%s\n\n", th.Border, th.Reset)

	if _, err := exec.LookPath("ollama"); err != nil {
		fmt.Printf("%sInstalling Ollama on Termux/Linux...%s\n", th.Warning, th.Reset)
		installCmd := exec.Command("bash", "-c", "curl -fsSL https://ollama.com/install.sh | sh || pkg install -y ollama || true")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		_ = installCmd.Run()
	}

	ollama.EnsureOllamaRunning()

	pullIndicator := ui.NewIndicator(fmt.Sprintf("Downloading %s (%s)...", chosenModel.ID, chosenModel.Size))
	pullIndicator.Start()
	pullCmd := exec.Command("ollama", "pull", chosenModel.ID)
	_ = pullCmd.Run()
	pullIndicator.StopSuccess(fmt.Sprintf("Model %s ready!", chosenModel.ID))

	if configFile != "" {
		_ = os.WriteFile(configFile, []byte(fmt.Sprintf("export TERTIC_PROVIDER=\"ollama\"\nexport OLLAMA_MODEL=\"%s\"\nexport OLLAMA_HOST=\"%s\"\n", chosenModel.ID, ollama.DefaultHost)), 0600)
	}

	fmt.Printf("\n%s══════════════════════════════════════════════════════════%s\n", th.Success, th.Reset)
	fmt.Printf("%sSetup Complete! Ask Tertic anything below to begin coding.%s\n", th.Bold, th.Reset)
	fmt.Printf("%s══════════════════════════════════════════════════════════%s\n\n", th.Success, th.Reset)

	return &AgentConfig{
		Provider: "ollama",
		Model:    chosenModel.ID,
		Host:     ollama.DefaultHost,
	}, nil
}

// Interactive Model Switcher (/model)
func switchModelInteractive(cfg *AgentConfig) {
	fmt.Println()
	categoryOpts := []string{
		"1) 🌐 Cloud Remote API Models (Google Gemini, OpenRouter - Privileged)",
		"2) 📱 Local Offline Ollama Models (Downloaded on Device)",
		"3) 🔑 Change / Enter API Key",
	}

	catChoice := ui.SelectInteractiveWithTheme("What would you like to switch to? (Use ↑/↓ arrows & Enter):", categoryOpts)

	if catChoice == 0 {
		cloudOpts := []string{
			"Google Gemini 3.1 Flash-Lite (High speed, Free tier)",
			"Google Gemini 3 Flash Preview (Reasoning & Full tools)",
			"OpenRouter: Google Gemma 4 31B (Free, Fast, Tool-Enabled)",
			"OpenRouter: Nvidia Nemotron 3.5 Lightning (Free remote reasoning)",
			"OpenRouter: Qwen 2.5 Coder 32B (Free remote coding)",
			"OpenRouter: DeepSeek R1 (Free remote reasoning)",
		}
		cIdx := ui.SelectInteractiveWithTheme("Select Cloud Model to Use:", cloudOpts)
		switch cIdx {
		case 0:
			cfg.Provider = "gemini"
			cfg.Model = "gemini-3.1-flash-lite"
		case 1:
			cfg.Provider = "gemini"
			cfg.Model = "gemini-3-flash-preview"
		case 2:
			cfg.Provider = "openrouter"
			cfg.Model = "google/gemma-4-31b-it:free"
			if cfg.APIKey == "" {
				cfg.APIKey = getBuiltinOpenRouterKey()
			}
		case 3:
			cfg.Provider = "openrouter"
			cfg.Model = "nvidia/nemotron-3.5-lightning:free"
			if cfg.APIKey == "" {
				cfg.APIKey = getBuiltinOpenRouterKey()
			}
		case 4:
			cfg.Provider = "openrouter"
			cfg.Model = "qwen/qwen-2.5-coder-32b-instruct:free"
			if cfg.APIKey == "" {
				cfg.APIKey = getBuiltinOpenRouterKey()
			}
		case 5:
			cfg.Provider = "openrouter"
			cfg.Model = "deepseek/deepseek-r1:free"
			if cfg.APIKey == "" {
				cfg.APIKey = getBuiltinOpenRouterKey()
			}
		}

	} else if catChoice == 1 {
		cfg.Provider = "ollama"
		if cfg.Host == "" {
			cfg.Host = ollama.DefaultHost
		}
		models, err := ollama.GetDownloadedModels(cfg.Host)
		if err != nil || len(models) == 0 {
			fmt.Printf("%s[Info] No downloaded models found. Choose one to pull:%s\n", theme.Current().Warning, theme.Current().Reset)
			var opts []string
			for _, m := range ollama.AvailableLocalModels {
				opts = append(opts, fmt.Sprintf("%-22s [%-6s]", m.ID, m.Size))
			}
			idx := ui.SelectInteractiveWithTheme("Choose Model to Pull & Use (Use ↑/↓ arrows & Enter):", opts)
			chosen := ollama.AvailableLocalModels[idx].ID

			indicator := ui.NewIndicator("Pulling model: " + chosen)
			indicator.Start()
			_ = exec.Command("ollama", "pull", chosen).Run()
			indicator.StopSuccess(fmt.Sprintf("Model %s ready!", chosen))
			cfg.Model = chosen
		} else {
			opts := append(models, "+ Download another model...")
			idx := ui.SelectInteractiveWithTheme("Select Installed Model to Switch to (Use ↑/↓ arrows & Enter):", opts)

			if idx == len(models) {
				var newOpts []string
				for _, m := range ollama.AvailableLocalModels {
					newOpts = append(newOpts, fmt.Sprintf("%-22s [%-6s]", m.ID, m.Size))
				}
				newIdx := ui.SelectInteractiveWithTheme("Select Model to Download:", newOpts)
				chosen := ollama.AvailableLocalModels[newIdx].ID

				indicator := ui.NewIndicator("Pulling model: " + chosen)
				indicator.Start()
				_ = exec.Command("ollama", "pull", chosen).Run()
				indicator.StopSuccess(fmt.Sprintf("Model %s ready!", chosen))
				cfg.Model = chosen
			} else {
				parts := strings.Fields(models[idx])
				if len(parts) > 0 {
					cfg.Model = parts[0]
				}
			}
		}

	} else {
		keyOpts := []string{
			"Google Gemini API Key",
			"OpenRouter API Key",
		}
		kIdx := ui.SelectInteractiveWithTheme("Which API Key do you want to provide?", keyOpts)
		reader := bufio.NewReader(os.Stdin)
		if kIdx == 0 {
			fmt.Print("\nEnter your Google Gemini API Key: ")
			k, _ := reader.ReadString('\n')
			cfg.APIKey = strings.TrimSpace(k)
			cfg.Provider = "gemini"
			cfg.Model = gemini.DefaultModel
		} else {
			fmt.Print("\nEnter your OpenRouter API Key (sk-or-...): ")
			k, _ := reader.ReadString('\n')
			cfg.APIKey = strings.TrimSpace(k)
			cfg.Provider = "openrouter"
			cfg.Model = openrouter.DefaultModel
		}
	}

	// Persist changes
	home, _ := os.UserHomeDir()
	if home != "" {
		configFile := filepath.Join(home, ".tertic", "config.env")
		if cfg.Provider == "ollama" {
			_ = os.WriteFile(configFile, []byte(fmt.Sprintf("export TERTIC_PROVIDER=\"ollama\"\nexport OLLAMA_MODEL=\"%s\"\nexport OLLAMA_HOST=\"%s\"\n", cfg.Model, cfg.Host)), 0600)
		} else if cfg.Provider == "openrouter" {
			_ = os.WriteFile(configFile, []byte(fmt.Sprintf("export TERTIC_PROVIDER=\"openrouter\"\nexport OPENROUTER_API_KEY=\"%s\"\nexport OPENROUTER_MODEL=\"%s\"\n", cfg.APIKey, cfg.Model)), 0600)
		} else {
			_ = os.WriteFile(configFile, []byte(fmt.Sprintf("export TERTIC_PROVIDER=\"gemini\"\nexport GEMINI_API_KEY=\"%s\"\nexport GEMINI_MODEL=\"%s\"\n", cfg.APIKey, cfg.Model)), 0600)
		}
	}

	th := theme.Current()
	fmt.Printf("%s══════════════════════════════════════════════════════════%s\n", th.Success, th.Reset)
	fmt.Printf("%sActive Mode:%s %s%s%s │ %sModel:%s %s%s%s\n", th.Bold, th.Reset, th.Warning, cfg.Provider, th.Reset, th.Bold, th.Reset, th.Primary, cfg.Model, th.Reset)
	fmt.Printf("%s══════════════════════════════════════════════════════════%s\n\n", th.Success, th.Reset)
}

func isCodingTask(prompt string) bool {
	p := strings.ToLower(strings.TrimSpace(prompt))
	codingKeywords := []string{
		"build", "create", "make", "write", "code", "run", "exec", "test",
		"file", "install", "npm", "pip", "git", "bash", "delete", "edit", "list",
		"cat", "ls", "dir", "grep", "server", "app", "fix", "debug", "compile",
	}
	for _, kw := range codingKeywords {
		if strings.Contains(p, kw) {
			return true
		}
	}
	return false
}

func parseJSONToolCall(content string) (string, map[string]interface{}, bool) {
	s := strings.TrimSpace(content)
	start := strings.Index(s, `{"name":`)
	if start == -1 {
		start = strings.Index(s, `{"name" :`)
	}
	if start == -1 {
		start = strings.Index(s, "{\n  \"name\":")
	}
	if start == -1 {
		start = strings.Index(s, "{\n \"name\":")
	}
	if start == -1 {
		return "", nil, false
	}

	depth := 0
	end := -1
	for i := start; i < len(s); i++ {
		if s[i] == '{' {
			depth++
		} else if s[i] == '}' {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}

	if end == -1 {
		return "", nil, false
	}

	jsonStr := s[start:end]
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", nil, false
	}
	name, ok1 := data["name"].(string)
	args, ok2 := data["arguments"].(map[string]interface{})
	if ok1 && ok2 {
		return name, args, true
	}
	return "", nil, false
}

// --- OPENROUTER / OLLAMA COMPATIBLE TURN ---
type ChatClient interface {
	Chat(req openrouter.ChatCompletionRequest) (*openrouter.ChatCompletionResponse, error)
}

func runChatTurn(client ChatClient, history *[]openrouter.Message, userPrompt string) error {
	*history = append(*history, openrouter.Message{
		Role:    "user",
		Content: userPrompt,
	})

	var openRouterTools []openrouter.OpenRouterTool
	if isCodingTask(userPrompt) {
		toolDefs := tools.GetToolDefinitions()
		openRouterTools = openrouter.ConvertToolDefinitions(toolDefs)
	}

	maxSteps := 30
	for step := 1; step <= maxSteps; step++ {
		req := openrouter.ChatCompletionRequest{
			Messages: *history,
			Tools:    openRouterTools,
		}

		indicator := ui.NewIndicator("Thinking & planning...")
		indicator.Start()

		resp, err := client.Chat(req)
		if err != nil {
			indicator.Stop(fmt.Sprintf("%sError: %v%s", theme.Current().Error, err, theme.Current().Reset))
			return err
		}
		indicator.Stop("")

		choice := resp.Choices[0]
		assistantMsg := choice.Message
		*history = append(*history, assistantMsg)

		rawToolName, rawArgs, isRawTool := parseJSONToolCall(assistantMsg.Content)
		if isRawTool && isCodingTask(userPrompt) {
			assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, openrouter.ToolCall{
				ID: fmt.Sprintf("call_%d", step),
				Function: openrouter.FunctionCall{
					Name:      rawToolName,
					Arguments: assistantMsg.Content,
				},
			})
		} else if len(assistantMsg.ToolCalls) == 0 && assistantMsg.Content != "" {
			ui.StreamTypewriter(assistantMsg.Content)
		}

		if len(assistantMsg.ToolCalls) == 0 {
			break
		}

		for _, tc := range assistantMsg.ToolCalls {
			fn := tc.Function

			var args map[string]interface{}
			if isRawTool && rawArgs != nil {
				args = rawArgs
			} else {
				_ = json.Unmarshal([]byte(fn.Arguments), &args)
			}

			actionDesc := fn.Arguments
			if cmd, ok := args["command"].(string); ok {
				actionDesc = cmd
			} else if p, ok := args["path"].(string); ok {
				actionDesc = p
			}

			taskLabel := ui.FormatActionHumanReadable(fn.Name, args)
			toolIndicator := ui.NewIndicator(taskLabel)
			toolIndicator.Start()

			output, toolErr := tools.ExecuteTool(fn.Name, args)
			th := theme.Current()
			if toolErr != nil {
				if verboseProcess {
					toolIndicator.Stop(fmt.Sprintf("%s[%s failed]: %v%s", th.Error, fn.Name, toolErr, th.Reset))
				} else {
					toolIndicator.StopFailure(fmt.Sprintf("%s failed", taskLabel))
				}
				output = fmt.Sprintf("Error: %v", toolErr)
			} else {
				if verboseProcess {
					toolIndicator.Stop(fmt.Sprintf("%s%s [%s]:%s %s", th.Bold, th.Success, fn.Name, th.Reset, actionDesc))
					lines := strings.Split(output, "\n")
					preview := output
					if len(lines) > 4 {
						preview = strings.Join(lines[:4], "\n") + fmt.Sprintf("\n%s... (+%d lines)%s", th.Muted, len(lines)-4, th.Reset)
					}
					if strings.TrimSpace(preview) != "" {
						fmt.Printf("%s%s%s\n", th.Primary, preview, th.Reset)
					}
				} else {
					toolIndicator.StopSuccess(taskLabel)
				}
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

	var toolDefs []tools.ToolDefinition
	if isCodingTask(userPrompt) {
		toolDefs = tools.GetToolDefinitions()
	}

	req := gemini.GenerateContentRequest{
		SystemInstruction: &gemini.SystemInstruction{
			Parts: []gemini.Part{
				{Text: SystemPrompt},
			},
		},
	}
	if len(toolDefs) > 0 {
		req.Tools = []gemini.ToolDeclarationWrapper{
			{FunctionDeclarations: toolDefs},
		}
	}

	maxSteps := 30
	for step := 1; step <= maxSteps; step++ {
		req.Contents = *history

		indicator := ui.NewIndicator("Thinking & planning...")
		indicator.Start()

		resp, err := client.Generate(req)
		if err != nil {
			indicator.Stop(fmt.Sprintf("%sError: %v%s", theme.Current().Error, err, theme.Current().Reset))
			return err
		}
		indicator.Stop("")

		candidate := resp.Candidates[0]
		*history = append(*history, candidate.Content)

		hasToolCalls := false
		var toolResponseParts []gemini.Part

		for _, part := range candidate.Content.Parts {
			if part.FunctionCall == nil && part.Text != "" {
				ui.StreamTypewriter(part.Text)
			}

			if part.FunctionCall != nil {
				hasToolCalls = true
				call := part.FunctionCall

				actionDesc := fmt.Sprintf("%v", call.Args)
				if cmd, ok := call.Args["command"].(string); ok {
					actionDesc = cmd
				} else if p, ok := call.Args["path"].(string); ok {
					actionDesc = p
				}

				taskLabel := ui.FormatActionHumanReadable(call.Name, call.Args)
				toolIndicator := ui.NewIndicator(taskLabel)
				toolIndicator.Start()

				output, toolErr := tools.ExecuteTool(call.Name, call.Args)
				th := theme.Current()
				if toolErr != nil {
					if verboseProcess {
						toolIndicator.Stop(fmt.Sprintf("%s[%s failed]: %v%s", th.Error, call.Name, toolErr, th.Reset))
					} else {
						toolIndicator.StopFailure(fmt.Sprintf("%s failed", taskLabel))
					}
					output = fmt.Sprintf("Error: %v", toolErr)
				} else {
					if verboseProcess {
						toolIndicator.Stop(fmt.Sprintf("%s%s [%s]:%s %s", th.Bold, th.Success, call.Name, th.Reset, actionDesc))
						lines := strings.Split(output, "\n")
						preview := output
						if len(lines) > 4 {
							preview = strings.Join(lines[:4], "\n") + fmt.Sprintf("\n%s... (+%d lines)%s", th.Muted, len(lines)-4, th.Reset)
						}
						if strings.TrimSpace(preview) != "" {
							fmt.Printf("%s%s%s\n", th.Primary, preview, th.Reset)
						}
					} else {
						toolIndicator.StopSuccess(taskLabel)
					}
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
			break
		}
	}
	return nil
}

func main() {
	agentReg := agent.NewRegistry()
	voiceMgr := voice.NewManager()

	// Direct CLI arguments / subcommands
	if len(os.Args) > 1 {
		sub := strings.ToLower(os.Args[1])
		if sub == "agents" {
			th := theme.Current()
			if len(os.Args) > 2 && os.Args[2] == "status" {
				fmt.Printf("\n%sTertic Agent Mesh Status:%s\n", th.Bold, th.Reset)
				for _, ag := range agentReg.List() {
					fmt.Printf(" • %-12s [%-12s] Status: %sREADY%s │ Tools: %v\n", ag.Name(), ag.Role(), th.Success, th.Reset, ag.AllowedTools())
				}
				fmt.Println()
				return
			}
			fmt.Printf("\n%sAvailable Specialized Agents:%s\n", th.Bold, th.Reset)
			for _, ag := range agentReg.List() {
				fmt.Printf(" • %s%-12s%s ─ %s\n", th.Bold, ag.Role(), th.Reset, ag.Description())
			}
			fmt.Println()
			return
		}
	}

	cfg, err := resolveConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError: %v%s\n", theme.Current().Error, err, theme.Current().Reset)
		os.Exit(1)
	}

	var chatClient ChatClient
	var chatHistory []openrouter.Message

	var gemClient *gemini.Client
	var gemHistory []gemini.Content

	var multiAgentEngine *orchestrator.Engine
	modelReg := model.NewRegistry()

	var activeIndicator *ui.Indicator
	var activeIndicatorMu sync.Mutex

	stopMultiAgentProgress := func(finalMsg string) {
		activeIndicatorMu.Lock()
		defer activeIndicatorMu.Unlock()
		if activeIndicator != nil {
			activeIndicator.Stop(finalMsg)
			activeIndicator = nil
		}
	}

	initClients := func() {
		chatClient = nil
		chatHistory = nil
		gemClient = nil
		gemHistory = nil

		var primaryProvider model.Provider

		if cfg.Provider == "ollama" {
			ollama.EnsureOllamaRunning()
			chatClient = ollama.NewClient(cfg.Host, cfg.Model)
			chatHistory = append(chatHistory, openrouter.Message{
				Role:    "system",
				Content: SystemPrompt,
			})
			primaryProvider = model.NewOllamaProvider(cfg.Host, cfg.Model)
		} else if cfg.Provider == "openrouter" {
			chatClient = openrouter.NewClient(cfg.APIKey, cfg.Model)
			chatHistory = append(chatHistory, openrouter.Message{
				Role:    "system",
				Content: SystemPrompt,
			})
			primaryProvider = model.NewOpenRouterProvider(cfg.APIKey, cfg.Model)
		} else {
			gemClient = gemini.NewClient(cfg.APIKey, cfg.Model)
			primaryProvider = model.NewGeminiProvider(cfg.APIKey, cfg.Model)
		}

		modelReg.SetDefault(primaryProvider)
		for _, ag := range agentReg.List() {
			modelReg.Register(ag.Name(), primaryProvider)
		}

		multiAgentEngine = orchestrator.NewEngine(agentReg, modelReg, orchestrator.Config{
			MaxConcurrentAgents: 1, // Sequential execution for low-RAM Android/Termux
			SequentialOnly:      true,
			OnProgress: func(role, status string) {
				if role != "Action" && role != "Thinking" {
					return
				}
				activeIndicatorMu.Lock()
				defer activeIndicatorMu.Unlock()
				if activeIndicator == nil {
					activeIndicator = ui.NewIndicator(status)
					activeIndicator.Start()
				} else {
					activeIndicator.Update(status)
				}
			},
		})
	}

	initClients()

	isSingleAgentMode := false
	args := os.Args[1:]
	var filteredArgs []string
	for _, a := range args {
		if a == "--single-agent" {
			isSingleAgentMode = true
		} else {
			filteredArgs = append(filteredArgs, a)
		}
	}

	// Direct non-interactive execution
	if len(filteredArgs) > 0 {
		userPrompt := strings.Join(filteredArgs, " ")
		subCmd := strings.ToLower(filteredArgs[0])

		if subCmd == "agents" {
			th := theme.Current()
			if len(filteredArgs) > 1 && filteredArgs[1] == "status" {
				fmt.Printf("\n%sTertic Agent Mesh Status:%s\n", th.Bold, th.Reset)
				for _, ag := range agentReg.List() {
					fmt.Printf(" • %-12s [%-12s] Status: %sREADY%s │ Allowed: %v\n", ag.Name(), ag.Role(), th.Success, th.Reset, ag.AllowedTools())
				}
				fmt.Println()
				return
			}
			fmt.Printf("\n%sAvailable Specialized Agents:%s\n", th.Bold, th.Reset)
			for _, ag := range agentReg.List() {
				fmt.Printf(" • %s%-12s%s ─ %s\n", th.Bold, ag.Role(), th.Reset, ag.Description())
			}
			fmt.Println()
			return
		}

		if subCmd == "models" {
			th := theme.Current()
			fmt.Printf("\n%sConfigured Model Providers:%s\n", th.Bold, th.Reset)
			for _, m := range modelReg.List() {
				fmt.Printf(" • %s\n", m)
			}
			fmt.Println()
			return
		}

		if subCmd == "config" {
			th := theme.Current()
			fmt.Printf("\n%sActive Model Configuration:%s\n", th.Bold, th.Reset)
			fmt.Printf(" Provider: %s\n Model:    %s\n Mode:     Sequential Android (4GB RAM safe)\n\n", cfg.Provider, cfg.Model)
			return
		}

		th := theme.Current()
		fmt.Printf("\n%s%sTertic executing:%s %s\n\n", th.Bold, th.Primary, th.Reset, userPrompt)

		if !isSingleAgentMode && isCodingTask(userPrompt) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			out, err := multiAgentEngine.ExecuteWorkflow(ctx, userPrompt)
			if err != nil {
				stopMultiAgentProgress(fmt.Sprintf("%s✖%s %sWorkflow failed%s", th.Error, th.Reset, th.Bold, th.Reset))
				fmt.Fprintf(os.Stderr, "\n%sWorkflow Error: %v%s\n", th.Error, err, th.Reset)
			} else {
				stopMultiAgentProgress(fmt.Sprintf("%s✔%s %sAll tasks completed & verified%s", th.Success, th.Reset, th.Bold, th.Reset))
				fmt.Println()
				ui.StreamTypewriter(out)
			}
			return
		}

		if gemClient != nil {
			if err := runGeminiTurn(gemClient, &gemHistory, userPrompt); err != nil {
				fmt.Fprintf(os.Stderr, "\n%sError: %v%s\n", theme.Current().Error, err, theme.Current().Reset)
			}
		} else {
			if err := runChatTurn(chatClient, &chatHistory, userPrompt); err != nil {
				fmt.Fprintf(os.Stderr, "\n%sError: %v%s\n", theme.Current().Error, err, theme.Current().Reset)
			}
		}
		return
	}

	// Interactive Mode
	workDir, _ := os.Getwd()
	version := "2.1.0-premium"

	drawWelcome := func() {
		ui.RenderBanner(version, cfg.Provider, cfg.Model, workDir, voiceMgr.GetMode())
	}

	drawWelcome()

	reader := bufio.NewReader(os.Stdin)

	for {
		currentMode := voiceMgr.GetMode()

		// If Voice Mode is active as primary input, run voice flow
		if currentMode == voice.ModeVoice {
			res := ui.StartVoiceInputFlow(context.Background(), reader, voiceMgr)
			if res.Cancelled || !res.Submitted || res.Text == "" {
				// Continue waiting
				continue
			}
			executeTask(res.Text, cfg, gemClient, chatClient, multiAgentEngine, stopMultiAgentProgress, &gemHistory, &chatHistory)
			continue
		}

		// Text or Hybrid Mode Prompt
		promptPrefix := ui.RenderInputPrompt(currentMode)
		appState := commands.AppState{
			IsRecording: false,
			IsExecuting: false,
			VoiceActive: currentMode == voice.ModeVoice || currentMode == voice.ModeHybrid,
		}
		input, err := ui.ReadMultilineWithState(reader, promptPrefix, appState)
		if err != nil {
			break
		}

		prompt := strings.TrimSpace(input)
		if prompt == "" {
			continue
		}

		// Check Interactive Commands
		lowPrompt := strings.ToLower(prompt)
		if lowPrompt == "exit" || lowPrompt == "quit" || lowPrompt == "/exit" {
			fmt.Printf("%sGoodbye!%s\n", theme.Current().Primary, theme.Current().Reset)
			break
		}

		if lowPrompt == "/help" {
			ui.RenderHelp()
			continue
		}

		if lowPrompt == "/clear" {
			fmt.Print("\033[H\033[2J")
			drawWelcome()
			continue
		}

		if lowPrompt == "/model" || lowPrompt == "/models" {
			switchModelInteractive(cfg)
			initClients()
			continue
		}

		if lowPrompt == "/agents" {
			th := theme.Current()
			fmt.Printf("\n%sAvailable Specialized Agents:%s\n", th.Bold, th.Reset)
			for _, ag := range agentReg.List() {
				fmt.Printf(" • %s%-12s%s ─ %s\n", th.Bold, ag.Role(), th.Reset, ag.Description())
			}
			fmt.Println()
			continue
		}

		if lowPrompt == "/status" {
			th := theme.Current()
			fmt.Printf("\n%sTertic System & Agent Mesh Status:%s\n", th.Bold, th.Reset)
			for _, ag := range agentReg.List() {
				fmt.Printf(" • %-12s [%-12s] Status: %sREADY%s │ Tools: %v\n", ag.Name(), ag.Role(), th.Success, th.Reset, ag.AllowedTools())
			}
			fmt.Println()
			fmt.Println(voiceMgr.StatusReport())
			continue
		}

		if lowPrompt == "/theme" {
			thOpts := []string{
				"1) Dark (Default Cyan/Purple modern DevTools)",
				"2) Light (High visibility for daylight)",
				"3) Monochrome (Plain ASCII, no ANSI colors)",
				"4) High Contrast (Bold accessible borders)",
			}
			tChoice := ui.SelectInteractiveWithTheme("Select Color Theme:", thOpts)
			themes := theme.AvailableThemes()
			if tChoice >= 0 && tChoice < len(themes) {
				theme.SetTheme(themes[tChoice])
				fmt.Printf("%sTheme switched to: %s%s\n\n", theme.Current().Success, themes[tChoice], theme.Current().Reset)
			}
			continue
		}

		if lowPrompt == "/voice" {
			voiceMgr.SetMode(voice.ModeVoice)
			fmt.Printf("%s✔ Switched to Voice Input Mode (use /text to return)%s\n\n", theme.Current().RecordColor, theme.Current().Reset)
			continue
		}

		if lowPrompt == "/text" {
			voiceMgr.SetMode(voice.ModeText)
			fmt.Printf("%s✔ Switched to Keyboard Text Mode%s\n\n", theme.Current().Success, theme.Current().Reset)
			continue
		}

		if lowPrompt == "/input" {
			inputOpts := []string{
				"1) Text Mode (Traditional keyboard input - Default)",
				"2) Voice Mode (Direct spoken instructions)",
				"3) Hybrid Mode (Keyboard + voice shortcuts)",
			}
			choice := ui.SelectInteractiveWithTheme("Select Active Input Mode:", inputOpts)
			switch choice {
			case 0:
				voiceMgr.SetMode(voice.ModeText)
			case 1:
				voiceMgr.SetMode(voice.ModeVoice)
			case 2:
				voiceMgr.SetMode(voice.ModeHybrid)
			}
			fmt.Printf("%sActive Input Mode: %s%s\n\n", theme.Current().Success, strings.ToUpper(string(voiceMgr.GetMode())), theme.Current().Reset)
			continue
		}

		if lowPrompt == "/listen" || lowPrompt == "/stop" || lowPrompt == "/transcribe" {
			res := ui.StartVoiceInputFlow(context.Background(), reader, voiceMgr)
			if res.Cancelled || !res.Submitted || res.Text == "" {
				continue
			}
			prompt = res.Text
			// Execute task immediately with transcribed instruction
			executeTask(prompt, cfg, gemClient, chatClient, multiAgentEngine, stopMultiAgentProgress, &gemHistory, &chatHistory)
			fmt.Println()
			continue
		}

		if lowPrompt == "/verbose" || lowPrompt == "/process" {
			verboseProcess = !verboseProcess
			th := theme.Current()
			if verboseProcess {
				fmt.Printf("%s[Process Details: ON]%s Step-by-step tool output is visible.\n\n", th.Success, th.Reset)
			} else {
				fmt.Printf("%s[Process Details: OFF]%s Clean mode active (Devin/Antigravity concise format).\n\n", th.Warning, th.Reset)
			}
			continue
		}

		if lowPrompt == "/config" {
			th := theme.Current()
			fmt.Printf("\n%sActive Model Configuration:%s\n", th.Bold, th.Reset)
			fmt.Printf(" Provider: %s\n Model:    %s\n Mode:     Sequential Android (4GB RAM safe)\n\n", cfg.Provider, cfg.Model)
			continue
		}

		if lowPrompt == "/history" {
			hist := multiAgentEngine.GetSharedHistory()
			th := theme.Current()
			fmt.Printf("\n%sShared Conversation History (%d messages):%s\n", th.Bold, len(hist), th.Reset)
			for i, m := range hist {
				fmt.Printf(" [%d] %s%s%s: %s\n", i+1, th.Primary, m.Role, th.Reset, m.Content)
			}
			fmt.Println()
			continue
		}

		if lowPrompt == "/context" {
			th := theme.Current()
			fmt.Printf("\n%sProject Context in %s:%s\n", th.Bold, workDir, th.Reset)
			out, _ := tools.ExecuteTool("list_dir", map[string]interface{}{"path": "."})
			fmt.Println(out)
			fmt.Println()
			continue
		}

		executeTask(prompt, cfg, gemClient, chatClient, multiAgentEngine, stopMultiAgentProgress, &gemHistory, &chatHistory)
		fmt.Println()
	}
}

func executeTask(
	prompt string,
	cfg *AgentConfig,
	gemClient *gemini.Client,
	chatClient ChatClient,
	multiAgentEngine *orchestrator.Engine,
	stopMultiAgentProgress func(string),
	gemHistory *[]gemini.Content,
	chatHistory *[]openrouter.Message,
) {
	th := theme.Current()
	if isCodingTask(prompt) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		out, err := multiAgentEngine.ExecuteWorkflow(ctx, prompt)
		if err != nil {
			stopMultiAgentProgress(fmt.Sprintf("%s✖%s %sWorkflow failed%s", th.Error, th.Reset, th.Bold, th.Reset))
			fmt.Fprintf(os.Stderr, "%sWorkflow Error: %v%s\n", th.Error, err, th.Reset)
		} else {
			stopMultiAgentProgress(fmt.Sprintf("%s✔%s %sAll tasks completed & verified%s", th.Success, th.Reset, th.Bold, th.Reset))
			fmt.Println()
			ui.StreamTypewriter(out)
		}
	} else {
		// Conversational turn
		if gemClient != nil {
			if err := runGeminiTurn(gemClient, gemHistory, prompt); err != nil {
				fmt.Fprintf(os.Stderr, "%sError: %v%s\n", th.Error, err, th.Reset)
			}
		} else {
			if err := runChatTurn(chatClient, chatHistory, prompt); err != nil {
				fmt.Fprintf(os.Stderr, "%sError: %v%s\n", th.Error, err, th.Reset)
			}
		}
	}
}
