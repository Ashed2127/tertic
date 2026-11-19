package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/Ashed2127/tertic/internal/gemini"
	"github.com/Ashed2127/tertic/internal/ollama"
	"github.com/Ashed2127/tertic/internal/openrouter"
	"github.com/Ashed2127/tertic/internal/tools"
)

// ANSI colors & styling
const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
)

const Banner = `
  _____         _   _      
 |_   _|__ _ __| |_(_) ___ 
   | |/ _ \ '__| __| |/ __|
   | |  __/ |  | |_| | (__ 
   |_|\___|_|   \__|_|\___|
 Autonomous AI Coding Agent for Termux
`

const SystemPrompt = `You are Tertic, an autonomous AI software engineering agent running directly inside Termux on Android.
Your job is to act like Devin or Antigravity: directly execute tasks without narrating or dumping an exhaustive list of every step upfront. Take initiative, invoke tool actions immediately, and provide concise, high-level summaries upon completion.

When the user greets you or asks conversational questions (e.g. "hey", "hello", "what can you do?"), reply with normal conversational text.
Do not output long numbered lists of hypothetical steps—directly run the necessary bash commands, create/edit files, test, and deliver the result.`

// Motion Spinner
type Spinner struct {
	frames  []string
	msg     string
	stop    chan struct{}
	wg      sync.WaitGroup
	enabled bool
}

var verboseProcess = false // When true, shows tool execution details & previews; when false, clean concise mode

func NewSpinner(msg string) *Spinner {
	// Clean frame when running inside Termux non-interactive / wrapped terminals
	cleanMsg := msg
	if len(cleanMsg) > 60 {
		cleanMsg = cleanMsg[:57] + "..."
	}
	return &Spinner{
		// Bold "T" that glides and smoothly fades in and out from left to right
		frames: []string{
			"\033[38;5;238m\033[1mT\033[0m   ",
			"\033[38;5;30m\033[1mT\033[0m   ",
			" \033[38;5;37m\033[1mT\033[0m  ",
			" \033[38;5;44m\033[1mT\033[0m  ",
			"  \033[1m\033[97mT\033[0m ",
			"  \033[38;5;51m\033[1mT\033[0m ",
			"   \033[38;5;37m\033[1mT\033[0m",
			"   \033[38;5;240m\033[1mT\033[0m",
			"    ",
		},
		msg:     cleanMsg,
		stop:    make(chan struct{}),
		enabled: true,
	}
}

func (s *Spinner) Start() {
	if !s.enabled {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Print("\r\033[2K")
				return
			default:
				frame := s.frames[i%len(s.frames)]
				fmt.Printf("\r\033[2K%s %s%s%s", frame, ColorDim, s.msg, ColorReset)
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) Stop(finalMsg string) {
	if !s.enabled {
		return
	}
	close(s.stop)
	s.wg.Wait()
	if finalMsg != "" {
		fmt.Printf("\r\033[2K%s\n", finalMsg)
	} else {
		fmt.Print("\r\033[2K")
	}
}

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

// Arrow Key Interactive Selector
func selectInteractive(title string, options []string) int {
	fd := int(os.Stdin.Fd())
	var oldState syscall.Termios
	_, _, _ = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafePointer(&oldState)), 0, 0, 0)
	raw := oldState
	raw.Lflag &^= (syscall.ICANON | syscall.ECHO)
	_, _, _ = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafePointer(&raw)), 0, 0, 0)

	defer func() {
		_, _, _ = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafePointer(&oldState)), 0, 0, 0)
		fmt.Print("\033[?25h") // Show cursor
	}()

	fmt.Print("\033[?25l") // Hide cursor
	selected := 0

	render := func() {
		fmt.Printf("\r\033[%dA", len(options)+1) // move up
		fmt.Printf("%s%s%s\n", ColorBold, title, ColorReset)
		for i, opt := range options {
			if i == selected {
				fmt.Printf("\r\033[K %s%s❯ %s%s\n", ColorBold, ColorGreen, opt, ColorReset)
			} else {
				fmt.Printf("\r\033[K   %s%s%s\n", ColorDim, opt, ColorReset)
			}
		}
	}

	fmt.Printf("%s%s%s\n", ColorBold, title, ColorReset)
	for i, opt := range options {
		if i == selected {
			fmt.Printf(" %s%s❯ %s%s\n", ColorBold, ColorGreen, opt, ColorReset)
		} else {
			fmt.Printf("   %s%s%s\n", ColorDim, opt, ColorReset)
		}
	}

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}

		if n == 1 {
			if buf[0] == '\n' || buf[0] == '\r' {
				break
			}
			if buf[0] == 3 { // Ctrl+C
				os.Exit(0)
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
			switch buf[2] {
			case 'A': // UP arrow
				if selected > 0 {
					selected--
					render()
				}
			case 'B': // DOWN arrow
				if selected < len(options)-1 {
					selected++
					render()
				}
			}
		}
	}
	fmt.Println()
	return selected
}

func unsafePointer(t *syscall.Termios) uintptr {
	return uintptr(unsafe.Pointer(t))
}

// Config resolution
func resolveConfig() (*AgentConfig, error) {
	if p := os.Getenv("TERTIC_PROVIDER"); p == "ollama" {
		model := os.Getenv("OLLAMA_MODEL")
		if model == "" {
			model = "qwen2.5-coder:0.5b"
		}
		host := os.Getenv("OLLAMA_HOST")
		if host == "" {
			host = ollama.DefaultHost
		}
		return &AgentConfig{Provider: "ollama", Model: model, Host: host}, nil
	}

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
	modeChoice := selectInteractive("[Tertic Setup] Choose Execution Mode (Use ↑/↓ arrows & Enter):", modeOptions)

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
		apiChoice := selectInteractive("Choose Cloud Provider:", apiOptions)

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

	selectedModelIdx := selectInteractive("Select Local Model to Download & Run (Use ↑/↓ arrows & Enter):", modelOptions)
	chosenModel := ollama.AvailableLocalModels[selectedModelIdx]

	fmt.Printf("\n%s══════════════════════════════════════════════════════════%s\n", ColorCyan, ColorReset)
	fmt.Printf("%sSelected Local Model:%s %s%s%s (%s)\n", ColorBold, ColorReset, ColorGreen, chosenModel.ID, ColorReset, chosenModel.Size)
	fmt.Printf("%s══════════════════════════════════════════════════════════%s\n\n", ColorCyan, ColorReset)

	if _, err := exec.LookPath("ollama"); err != nil {
		fmt.Printf("%sInstalling Ollama on Termux/Linux...%s\n", ColorYellow, ColorReset)
		installCmd := exec.Command("bash", "-c", "curl -fsSL https://ollama.com/install.sh | sh || pkg install -y ollama || true")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		_ = installCmd.Run()
	}

	ollama.EnsureOllamaRunning()

	pullSpinner := NewSpinner(fmt.Sprintf("Downloading %s (%s)...", chosenModel.ID, chosenModel.Size))
	pullSpinner.Start()
	pullCmd := exec.Command("ollama", "pull", chosenModel.ID)
	_ = pullCmd.Run()
	pullSpinner.Stop(fmt.Sprintf("%sModel %s ready!%s", ColorGreen, chosenModel.ID, ColorReset))

	if configFile != "" {
		_ = os.WriteFile(configFile, []byte(fmt.Sprintf("export TERTIC_PROVIDER=\"ollama\"\nexport OLLAMA_MODEL=\"%s\"\nexport OLLAMA_HOST=\"%s\"\n", chosenModel.ID, ollama.DefaultHost)), 0600)
	}

	fmt.Printf("\n%s══════════════════════════════════════════════════════════%s\n", ColorGreen, ColorReset)
	fmt.Printf("%sSetup Complete! Ask Tertic anything below to begin coding.%s\n", ColorBold, ColorReset)
	fmt.Printf("%s══════════════════════════════════════════════════════════%s\n\n", ColorGreen, ColorReset)

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

	catChoice := selectInteractive("What would you like to switch to? (Use ↑/↓ arrows & Enter):", categoryOpts)

	if catChoice == 0 {
		// Switch to Cloud Models
		cloudOpts := []string{
			"Google Gemini 3.1 Flash-Lite (High speed, Free tier)",
			"Google Gemini 3 Flash Preview (Reasoning & Full tools)",
			"OpenRouter: Cohere North Mini Code (Free, Fast, Tool-Enabled)",
			"OpenRouter: Llama 3.3 70B (Free remote full-stack)",
			"OpenRouter: Qwen 2.5 Coder 32B (Free remote coding)",
			"OpenRouter: DeepSeek R1 (Free remote reasoning)",
		}
		cIdx := selectInteractive("Select Cloud Model to Use:", cloudOpts)
		switch cIdx {
		case 0:
			cfg.Provider = "gemini"
			cfg.Model = "gemini-3.1-flash-lite"
		case 1:
			cfg.Provider = "gemini"
			cfg.Model = "gemini-3-flash-preview"
		case 2:
			cfg.Provider = "openrouter"
			cfg.Model = "cohere/north-mini-code:free"
			if cfg.APIKey == "" {
				cfg.APIKey = getBuiltinOpenRouterKey()
			}
		case 3:
			cfg.Provider = "openrouter"
			cfg.Model = "meta-llama/llama-3.3-70b-instruct:free"
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
		// Switch to Local Ollama Models
		cfg.Provider = "ollama"
		if cfg.Host == "" {
			cfg.Host = ollama.DefaultHost
		}
		models, err := ollama.GetDownloadedModels(cfg.Host)
		if err != nil || len(models) == 0 {
			fmt.Printf("%s[Info] No downloaded models found. Choose one to pull:%s\n", ColorYellow, ColorReset)
			var opts []string
			for _, m := range ollama.AvailableLocalModels {
				opts = append(opts, fmt.Sprintf("%-22s [%-6s]", m.ID, m.Size))
			}
			idx := selectInteractive("Choose Model to Pull & Use (Use ↑/↓ arrows & Enter):", opts)
			chosen := ollama.AvailableLocalModels[idx].ID

			spinner := NewSpinner("Pulling model: " + chosen)
			spinner.Start()
			_ = exec.Command("ollama", "pull", chosen).Run()
			spinner.Stop(fmt.Sprintf("%sModel %s ready!%s", ColorGreen, chosen, ColorReset))
			cfg.Model = chosen
		} else {
			opts := append(models, "+ Download another model...")
			idx := selectInteractive("Select Installed Model to Switch to (Use ↑/↓ arrows & Enter):", opts)

			if idx == len(models) {
				var newOpts []string
				for _, m := range ollama.AvailableLocalModels {
					newOpts = append(newOpts, fmt.Sprintf("%-22s [%-6s]", m.ID, m.Size))
				}
				newIdx := selectInteractive("Select Model to Download:", newOpts)
				chosen := ollama.AvailableLocalModels[newIdx].ID

				spinner := NewSpinner("Pulling model: " + chosen)
				spinner.Start()
				_ = exec.Command("ollama", "pull", chosen).Run()
				spinner.Stop(fmt.Sprintf("%sModel %s ready!%s", ColorGreen, chosen, ColorReset))
				cfg.Model = chosen
			} else {
				parts := strings.Fields(models[idx])
				if len(parts) > 0 {
					cfg.Model = parts[0]
				}
			}
		}

	} else {
		// Change API Key option
		keyOpts := []string{
			"Google Gemini API Key",
			"OpenRouter API Key",
		}
		kIdx := selectInteractive("Which API Key do you want to provide?", keyOpts)
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

	fmt.Printf("%s══════════════════════════════════════════════════════════%s\n", ColorGreen, ColorReset)
	fmt.Printf("%sActive Mode:%s %s%s%s | %sModel:%s %s%s%s\n", ColorBold, ColorReset, ColorYellow, cfg.Provider, ColorReset, ColorBold, ColorReset, ColorCyan, cfg.Model, ColorReset)
	fmt.Printf("%s══════════════════════════════════════════════════════════%s\n\n", ColorGreen, ColorReset)
}

func printStreamed(text string) {
	if text == "" {
		return
	}
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '\n'
	})
	_ = words

	// Stream rune by rune / token with micro delays for live typewriter output
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		fmt.Print(string(runes[i]))
		if runes[i] == '\n' {
			time.Sleep(15 * time.Millisecond)
		} else if runes[i] == '.' || runes[i] == '!' || runes[i] == '?' {
			time.Sleep(12 * time.Millisecond)
		} else {
			time.Sleep(5 * time.Millisecond)
		}
	}
	fmt.Println()
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

		spinner := NewSpinner("Thinking & planning...")
		spinner.Start()

		resp, err := client.Chat(req)
		if err != nil {
			spinner.Stop(fmt.Sprintf("%sError: %v%s", ColorRed, err, ColorReset))
			return err
		}
		spinner.Stop("")

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
		} else if assistantMsg.Content != "" {
			printStreamed(assistantMsg.Content)
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

			toolSpinner := NewSpinner(fmt.Sprintf("Running %s: %s", fn.Name, actionDesc))
			if verboseProcess {
				toolSpinner.Start()
			}

			output, toolErr := tools.ExecuteTool(fn.Name, args)
			if toolErr != nil {
				if verboseProcess {
					toolSpinner.Stop(fmt.Sprintf("%s[%s failed]: %v%s", ColorRed, fn.Name, toolErr, ColorReset))
				}
				output = fmt.Sprintf("Error: %v", toolErr)
			} else {
				if verboseProcess {
					toolSpinner.Stop(fmt.Sprintf("%s%s [%s]:%s %s", ColorBold, ColorGreen, fn.Name, ColorReset, actionDesc))
					lines := strings.Split(output, "\n")
					preview := output
					if len(lines) > 4 {
						preview = strings.Join(lines[:4], "\n") + fmt.Sprintf("\n%s... (+%d lines)%s", ColorDim, len(lines)-4, ColorReset)
					}
					if strings.TrimSpace(preview) != "" {
						fmt.Printf("%s%s%s\n", ColorCyan, preview, ColorReset)
					}
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

		spinner := NewSpinner("Thinking & planning...")
		spinner.Start()

		resp, err := client.Generate(req)
		if err != nil {
			spinner.Stop(fmt.Sprintf("%sError: %v%s", ColorRed, err, ColorReset))
			return err
		}
		spinner.Stop("")

		candidate := resp.Candidates[0]
		*history = append(*history, candidate.Content)

		hasToolCalls := false
		var toolResponseParts []gemini.Part

		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				printStreamed(part.Text)
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

				toolSpinner := NewSpinner(fmt.Sprintf("Running %s: %s", call.Name, actionDesc))
				if verboseProcess {
					toolSpinner.Start()
				}

				output, toolErr := tools.ExecuteTool(call.Name, call.Args)
				if toolErr != nil {
					if verboseProcess {
						toolSpinner.Stop(fmt.Sprintf("%s[%s failed]: %v%s", ColorRed, call.Name, toolErr, ColorReset))
					}
					output = fmt.Sprintf("Error: %v", toolErr)
				} else {
					if verboseProcess {
						toolSpinner.Stop(fmt.Sprintf("%s%s [%s]:%s %s", ColorBold, ColorGreen, call.Name, ColorReset, actionDesc))
						lines := strings.Split(output, "\n")
						preview := output
						if len(lines) > 4 {
							preview = strings.Join(lines[:4], "\n") + fmt.Sprintf("\n%s... (+%d lines)%s", ColorDim, len(lines)-4, ColorReset)
						}
						if strings.TrimSpace(preview) != "" {
							fmt.Printf("%s%s%s\n", ColorCyan, preview, ColorReset)
						}
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
	cfg, err := resolveConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	var chatClient ChatClient
	var chatHistory []openrouter.Message

	var gemClient *gemini.Client
	var gemHistory []gemini.Content

	initClients := func() {
		chatClient = nil
		chatHistory = nil
		gemClient = nil
		gemHistory = nil

		if cfg.Provider == "ollama" {
			ollama.EnsureOllamaRunning()
			chatClient = ollama.NewClient(cfg.Host, cfg.Model)
			chatHistory = append(chatHistory, openrouter.Message{
				Role:    "system",
				Content: SystemPrompt,
			})
		} else if cfg.Provider == "openrouter" {
			chatClient = openrouter.NewClient(cfg.APIKey, cfg.Model)
			chatHistory = append(chatHistory, openrouter.Message{
				Role:    "system",
				Content: SystemPrompt,
			})
		} else {
			gemClient = gemini.NewClient(cfg.APIKey, cfg.Model)
		}
	}

	initClients()

	if len(os.Args) > 1 {
		userPrompt := strings.Join(os.Args[1:], " ")
		fmt.Printf("\n%s%sTertic Agent executing:%s %s\n\n", ColorBold, ColorCyan, ColorReset, userPrompt)
		if gemClient != nil {
			if err := runGeminiTurn(gemClient, &gemHistory, userPrompt); err != nil {
				fmt.Fprintf(os.Stderr, "\n%sError: %v%s\n", ColorRed, err, ColorReset)
			}
		} else {
			if err := runChatTurn(chatClient, &chatHistory, userPrompt); err != nil {
				fmt.Fprintf(os.Stderr, "\n%sError: %v%s\n", ColorRed, err, ColorReset)
			}
		}
		return
	}

	fmt.Print(ColorCyan, Banner, ColorReset)
	fmt.Printf("%sMode:%s %s | %sModel:%s %s  %s(/model to switch | /process to toggle)%s\n", ColorBold, ColorReset, cfg.Provider, ColorBold, ColorReset, cfg.Model, ColorDim, ColorReset)
	fmt.Printf("%sReady for instructions%s (type your task or 'exit' to quit)\n\n", ColorDim, ColorReset)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s%strtc ❯%s ", ColorBold, ColorGreen, ColorReset)
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		prompt := strings.TrimSpace(input)
		if prompt == "" {
			continue
		}
		if prompt == "exit" || prompt == "quit" {
			fmt.Printf("%sGoodbye!%s\n", ColorCyan, ColorReset)
			break
		}

		if prompt == "/model" || prompt == "/models" {
			switchModelInteractive(cfg)
			initClients()
			continue
		}

		if prompt == "/process" || prompt == "/verbose" {
			verboseProcess = !verboseProcess
			if verboseProcess {
				fmt.Printf("%s[Process Details: ON]%s Raw step-by-step tool output is now visible.\n\n", ColorGreen, ColorReset)
			} else {
				fmt.Printf("%s[Process Details: OFF]%s Clean mode active (hidden tool noise, like Devin/Antigravity).\n\n", ColorYellow, ColorReset)
			}
			continue
		}

		if gemClient != nil {
			if err := runGeminiTurn(gemClient, &gemHistory, prompt); err != nil {
				fmt.Fprintf(os.Stderr, "%sError: %v%s\n\n", ColorRed, err, ColorReset)
			}
		} else {
			if err := runChatTurn(chatClient, &chatHistory, prompt); err != nil {
				fmt.Fprintf(os.Stderr, "%sError: %v%s\n\n", ColorRed, err, ColorReset)
			}
		}
		fmt.Println()
	}
}
