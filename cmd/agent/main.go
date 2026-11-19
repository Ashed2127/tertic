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
	"syscall"
	"time"
	"unsafe"

	"github.com/Ashed2127/tertic/internal/agent"
	"github.com/Ashed2127/tertic/internal/gemini"
	"github.com/Ashed2127/tertic/internal/model"
	"github.com/Ashed2127/tertic/internal/ollama"
	"github.com/Ashed2127/tertic/internal/openrouter"
	"github.com/Ashed2127/tertic/internal/orchestrator"
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

const SystemPrompt = `You are Tertic, an elite autonomous software engineering agent running natively inside Termux on Android. You operate as a tightly coordinated multi-agent system (Coordinator → Planner → Researcher → Developer → Reviewer). Your sole mission is to deliver working, verified software with maximum reliability and minimum wasted steps.

CORE PRINCIPLES (NON-NEGOTIABLE)
1. Verification over claims — Never claim a task is done until you have actually run the relevant check (compile, test, lint, or execute).
2. Action over narration — Call tools immediately. Avoid long explanations of what you will do.
3. Smallest viable change — Prefer precise edits over rewriting large files.
4. Token efficiency — Mobile context is precious. Be extremely concise. Summarize aggressively.
5. Self-correction — When something fails, diagnose root cause, apply the smallest fix, and re-verify. Do not loop blindly.

REASONING PROTOCOL (ReAct + Reflect)
For every non-trivial task:
1. OBSERVE: State current state + relevant facts in 1–3 sentences max.
2. THINK: Decide the single highest-leverage next action.
3. ACT: Call the necessary tools.
4. REFLECT: Evaluate progress against acceptance criteria. Update plan if needed.

When the user greets you or asks conversational questions, reply with normal concise text.
Do not output long hypothetical step lists upfront—directly invoke the necessary tools, verify, and deliver working code.`

// Motion Spinner
type Spinner struct {
	frames  []string
	msg     string
	stop    chan struct{}
	mu      sync.Mutex
	wg      sync.WaitGroup
	enabled bool
}

var verboseProcess = false // When true, shows tool execution details & previews; when false, clean concise mode

func NewSpinner(msg string) *Spinner {
	cleanMsg := msg
	if len(cleanMsg) > 60 {
		cleanMsg = cleanMsg[:57] + "..."
	}
	return &Spinner{
		// Fluid sine-wave Bold "T" with high-definition RGB easing from left to right
		frames: []string{
			"\033[38;2;15;45;60m\033[1mT\033[0m   ",
			"\033[38;2;54;85;98m\033[1mT\033[0m   ",
			"\033[38;2;93;125;134m\033[1mT\033[0m   ",
			"\033[38;2;128;161;168m\033[1mT\033[0m   ",
			"\033[38;2;159;193;197m\033[1mT\033[0m   ",
			" \033[38;2;185;219;222m\033[1mT\033[0m  ",
			" \033[38;2;204;239;240m\033[1mT\033[0m  ",
			" \033[38;2;216;250;251m\033[1mT\033[0m  ",
			"  \033[38;2;220;255;255m\033[1mT\033[0m ",
			"  \033[38;2;216;250;251m\033[1mT\033[0m ",
			"  \033[38;2;204;239;240m\033[1mT\033[0m ",
			"  \033[38;2;185;219;222m\033[1mT\033[0m ",
			"   \033[38;2;159;193;197m\033[1mT\033[0m",
			"   \033[38;2;128;161;168m\033[1mT\033[0m",
			"   \033[38;2;93;125;134m\033[1mT\033[0m",
			"   \033[38;2;54;85;98m\033[1mT\033[0m",
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
				s.mu.Lock()
				m := s.msg
				s.mu.Unlock()
				frame := s.frames[i%len(s.frames)]
				fmt.Printf("\r\033[2K%s \033[38;2;140;185;200m%s\033[0m", frame, m)
				i++
				time.Sleep(50 * time.Millisecond)
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
		// Clean fade-out motion on complete
		s.mu.Lock()
		oldMsg := s.msg
		s.mu.Unlock()
		fadeSteps := []int{160, 100, 40}
		for _, v := range fadeSteps {
			fmt.Printf("\r\033[2K\033[38;2;%d;%d;%dm\033[1mT\033[0m   \033[38;2;%d;%d;%dm%s%s", v/2, v/2, v/2, v, v, v, oldMsg, ColorReset)
			time.Sleep(14 * time.Millisecond)
		}
		fmt.Print("\r\033[2K")
	}
}

func (s *Spinner) StopSuccess(msg string) {
	s.Stop(fmt.Sprintf("%s✔%s %s%s%s", ColorGreen, ColorReset, ColorBold, msg, ColorReset))
}

func (s *Spinner) StopFailure(msg string) {
	s.Stop(fmt.Sprintf("%s✖%s %s%s%s", ColorRed, ColorReset, ColorBold, msg, ColorReset))
}

func (s *Spinner) UpdateMessage(newMsg string) {
	if !s.enabled {
		return
	}
	cleanMsg := newMsg
	if len(cleanMsg) > 60 {
		cleanMsg = cleanMsg[:57] + "..."
	}
	s.mu.Lock()
	s.msg = cleanMsg
	s.mu.Unlock()
}

// TransitionTo smoothly transitions the spinner message with a fluid fade-out/fade-in wave
func (s *Spinner) TransitionTo(newMsg string) {
	if !s.enabled {
		return
	}
	cleanMsg := newMsg
	if len(cleanMsg) > 60 {
		cleanMsg = cleanMsg[:57] + "..."
	}

	s.mu.Lock()
	oldMsg := s.msg
	s.mu.Unlock()

	// Smoothly fade-out previous step text
	fadeSteps := []int{180, 120, 60, 10}
	for _, v := range fadeSteps {
		fmt.Printf("\r\033[2K\033[38;2;%d;%d;%dm\033[1mT\033[0m   \033[38;2;%d;%d;%dm%s%s", v/2, v/2, v/2, v, v, v, oldMsg, ColorReset)
		time.Sleep(16 * time.Millisecond)
	}

	s.mu.Lock()
	s.msg = cleanMsg
	s.mu.Unlock()

	// Smoothly fade-in new step text
	fadeUpSteps := []int{20, 80, 140, 210}
	for _, v := range fadeUpSteps {
		fmt.Printf("\r\033[2K\033[38;2;%d;%d;%dm\033[1mT\033[0m   \033[38;2;%d;%d;%dm%s%s", v/2, v, v, v, v, v, cleanMsg, ColorReset)
		time.Sleep(16 * time.Millisecond)
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

func formatActionMessage(name string, args map[string]interface{}) string {
	cmd := ""
	if c, ok := args["command"].(string); ok {
		cmd = strings.ToLower(c)
	}
	path := ""
	if p, ok := args["path"].(string); ok {
		path = filepath.Base(p)
	}

	switch name {
	case "execute_bash":
		if strings.Contains(cmd, "install") || strings.Contains(cmd, "npm i") || strings.Contains(cmd, "pip i") || strings.Contains(cmd, "pkg in") || strings.Contains(cmd, "apt in") || strings.Contains(cmd, "go get") {
			return "Installing packages..."
		}
		if strings.Contains(cmd, "am start") || strings.Contains(cmd, "termux-open") || strings.Contains(cmd, "xdg-open") || strings.Contains(cmd, "open ") {
			return "Opening application..."
		}
		if strings.Contains(cmd, "mkdir") || strings.Contains(cmd, " mv ") || strings.Contains(cmd, " cp ") || strings.Contains(cmd, " rm ") {
			return "Organizing files..."
		}
		if strings.Contains(cmd, "chmod") || strings.Contains(cmd, "config") || strings.Contains(cmd, "setup") || strings.Contains(cmd, "export") {
			return "Configuring environment..."
		}
		if strings.Contains(cmd, "build") || strings.Contains(cmd, "make") || strings.Contains(cmd, "compile") {
			return "Building project..."
		}
		if strings.Contains(cmd, "test") || strings.Contains(cmd, "check") {
			return "Testing & verifying..."
		}
		return "Executing command..."
	case "write_file":
		if path != "" {
			return fmt.Sprintf("Writing %s...", path)
		}
		return "Writing code..."
	case "edit_file":
		if path != "" {
			return fmt.Sprintf("Editing %s...", path)
		}
		return "Refactoring code..."
	case "read_file":
		if path != "" {
			return fmt.Sprintf("Reading %s...", path)
		}
		return "Reading code..."
	case "list_dir":
		return "Inspecting files..."
	default:
		return "Processing..."
	}
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

			taskLabel := formatActionMessage(fn.Name, args)
			toolSpinner := NewSpinner(taskLabel)
			toolSpinner.Start()

			output, toolErr := tools.ExecuteTool(fn.Name, args)
			if toolErr != nil {
				if verboseProcess {
					toolSpinner.Stop(fmt.Sprintf("%s[%s failed]: %v%s", ColorRed, fn.Name, toolErr, ColorReset))
				} else {
					toolSpinner.StopFailure(fmt.Sprintf("%s failed", taskLabel))
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
				} else {
					toolSpinner.StopSuccess(taskLabel)
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

				taskLabel := formatActionMessage(call.Name, call.Args)
				toolSpinner := NewSpinner(taskLabel)
				toolSpinner.Start()

				output, toolErr := tools.ExecuteTool(call.Name, call.Args)
				if toolErr != nil {
					if verboseProcess {
						toolSpinner.Stop(fmt.Sprintf("%s[%s failed]: %v%s", ColorRed, call.Name, toolErr, ColorReset))
					} else {
						toolSpinner.StopFailure(fmt.Sprintf("%s failed", taskLabel))
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
					} else {
						toolSpinner.StopSuccess(taskLabel)
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
	// First check informational subcommands before triggering interactive setup
	agentReg := agent.NewRegistry()
	if len(os.Args) > 1 {
		sub := strings.ToLower(os.Args[1])
		if sub == "agents" {
			if len(os.Args) > 2 && os.Args[2] == "status" {
				fmt.Printf("\n%sTertic Agent Mesh Status:%s\n", ColorBold, ColorReset)
				for _, ag := range agentReg.List() {
					fmt.Printf(" • %-12s [%-12s] Status: %sREADY%s | Tools: %v\n", ag.Name(), ag.Role(), ColorGreen, ColorReset, ag.AllowedTools())
				}
				fmt.Println()
				return
			}
			fmt.Printf("\n%sAvailable Specialized Agents:%s\n", ColorBold, ColorReset)
			for _, ag := range agentReg.List() {
				fmt.Printf(" • %s%-12s%s - %s\n", ColorBold, ag.Role(), ColorReset, ag.Description())
			}
			fmt.Println()
			return
		}
	}

	cfg, err := resolveConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	var chatClient ChatClient
	var chatHistory []openrouter.Message

	var gemClient *gemini.Client
	var gemHistory []gemini.Content

	var multiAgentEngine *orchestrator.Engine
	modelReg := model.NewRegistry()

	var stopMultiAgentProgress func(finalMsg string)

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

		var activeProgressSpinner *Spinner
		var activeProgressMu sync.Mutex

		multiAgentEngine = orchestrator.NewEngine(agentReg, modelReg, orchestrator.Config{
			MaxConcurrentAgents: 1, // Sequential execution for low-RAM Android/Termux
			SequentialOnly:      true,
			OnProgress: func(role, status string) {
				activeProgressMu.Lock()
				defer activeProgressMu.Unlock()
				msg := fmt.Sprintf("[%s] %s", role, status)
				if activeProgressSpinner == nil {
					activeProgressSpinner = NewSpinner(msg)
					activeProgressSpinner.Start()
				} else {
					activeProgressSpinner.TransitionTo(msg)
				}
			},
		})

		stopMultiAgentProgress = func(finalMsg string) {
			activeProgressMu.Lock()
			defer activeProgressMu.Unlock()
			if activeProgressSpinner != nil {
				activeProgressSpinner.Stop(finalMsg)
				activeProgressSpinner = nil
			}
		}
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

	if len(filteredArgs) > 0 {
		userPrompt := strings.Join(filteredArgs, " ")
		subCmd := strings.ToLower(filteredArgs[0])

		if subCmd == "agents" {
			if len(filteredArgs) > 1 && filteredArgs[1] == "status" {
				fmt.Printf("\n%sTertic Agent Mesh Status:%s\n", ColorBold, ColorReset)
				for _, ag := range agentReg.List() {
					fmt.Printf(" • %-12s [%-12s] Status: %sREADY%s | Allowed: %v\n", ag.Name(), ag.Role(), ColorGreen, ColorReset, ag.AllowedTools())
				}
				fmt.Println()
				return
			}
			fmt.Printf("\n%sAvailable Specialized Agents:%s\n", ColorBold, ColorReset)
			for _, ag := range agentReg.List() {
				fmt.Printf(" • %s%-12s%s - %s\n", ColorBold, ag.Role(), ColorReset, ag.Description())
			}
			fmt.Println()
			return
		}

		if subCmd == "models" {
			fmt.Printf("\n%sConfigured Model Providers:%s\n", ColorBold, ColorReset)
			for _, m := range modelReg.List() {
				fmt.Printf(" • %s\n", m)
			}
			fmt.Println()
			return
		}

		if subCmd == "config" {
			fmt.Printf("\n%sActive Model Configuration:%s\n", ColorBold, ColorReset)
			fmt.Printf(" Provider: %s\n Model:    %s\n Mode:     Sequential Android (4GB RAM safe)\n\n", cfg.Provider, cfg.Model)
			return
		}

		// Direct Task Execution
		fmt.Printf("\n%s%sTertic executing:%s %s\n\n", ColorBold, ColorCyan, ColorReset, userPrompt)

		if !isSingleAgentMode && isCodingTask(userPrompt) {
			// Multi-Agent Workflow
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			out, err := multiAgentEngine.ExecuteWorkflow(ctx, userPrompt)
			if err != nil {
				stopMultiAgentProgress(fmt.Sprintf("%s✖%s %sWorkflow failed%s", ColorRed, ColorReset, ColorBold, ColorReset))
				fmt.Fprintf(os.Stderr, "\n%sWorkflow Error: %v%s\n", ColorRed, err, ColorReset)
			} else {
				stopMultiAgentProgress(fmt.Sprintf("%s✔%s %sAll tasks completed & verified%s", ColorGreen, ColorReset, ColorBold, ColorReset))
				fmt.Println()
				printStreamed(out)
			}
			return
		}

		// Single-Agent Direct Execution
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
	fmt.Printf("%sMulti-Agent Autonomous Orchestrator Ready%s (type your task, /agents, or 'exit')\n\n", ColorDim, ColorReset)

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

		if prompt == "/agents" {
			fmt.Printf("\n%sAvailable Specialized Agents:%s\n", ColorBold, ColorReset)
			for _, ag := range agentReg.List() {
				fmt.Printf(" • %s%-12s%s - %s\n", ColorBold, ag.Role(), ColorReset, ag.Description())
			}
			fmt.Println()
			continue
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

		if isCodingTask(prompt) {
			// Multi-Agent workflow execution
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			out, err := multiAgentEngine.ExecuteWorkflow(ctx, prompt)
			cancel()
			if err != nil {
				stopMultiAgentProgress(fmt.Sprintf("%s✖%s %sWorkflow failed%s", ColorRed, ColorReset, ColorBold, ColorReset))
				fmt.Fprintf(os.Stderr, "%sWorkflow Error: %v%s\n\n", ColorRed, err, ColorReset)
			} else {
				stopMultiAgentProgress(fmt.Sprintf("%s✔%s %sAll tasks completed & verified%s", ColorGreen, ColorReset, ColorBold, ColorReset))
				fmt.Println()
				printStreamed(out)
			}
		} else {
			// Conversational turn
			if gemClient != nil {
				if err := runGeminiTurn(gemClient, &gemHistory, prompt); err != nil {
					fmt.Fprintf(os.Stderr, "%sError: %v%s\n\n", ColorRed, err, ColorReset)
				}
			} else {
				if err := runChatTurn(chatClient, &chatHistory, prompt); err != nil {
					fmt.Fprintf(os.Stderr, "%sError: %v%s\n\n", ColorRed, err, ColorReset)
				}
			}
		}
		fmt.Println()
	}
}
