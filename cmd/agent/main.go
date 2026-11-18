package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ashed2127/tertic/internal/gemini"
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
 Autonomous AI Coding Agent for Termux (Gemini Flash)
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

// resolveAPIKey checks environment variables, ~/.tertic/config.env, or prompts the user interactively
func resolveAPIKey() (string, error) {
	// 1. Direct env var
	if key := os.Getenv("GEMINI_API_KEY"); strings.TrimSpace(key) != "" {
		return strings.TrimSpace(key), nil
	}
	if key := os.Getenv("GOOGLE_API_KEY"); strings.TrimSpace(key) != "" {
		return strings.TrimSpace(key), nil
	}

	// 2. Saved Tertic config (~/.tertic/config.env)
	home, err := os.UserHomeDir()
	if err == nil {
		configFile := filepath.Join(home, ".tertic", "config.env")
		if data, err := os.ReadFile(configFile); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "export GEMINI_API_KEY=") {
					val := strings.TrimPrefix(line, "export GEMINI_API_KEY=")
					val = strings.Trim(val, `"' `)
					if val != "" {
						return val, nil
					}
				}
				if strings.HasPrefix(line, "export GOOGLE_API_KEY=") {
					val := strings.TrimPrefix(line, "export GOOGLE_API_KEY=")
					val = strings.Trim(val, `"' `)
					if val != "" {
						return val, nil
					}
				}
			}
		}
	}

	// 3. Prompt user interactively
	fmt.Printf("%s%s[Tertic Setup]%s Google Gemini API Key required.\n", ColorBold, ColorYellow, ColorReset)
	fmt.Printf("Get your free key at: %shttps://aistudio.google.com/app/apikey%s\n\n", ColorCyan, ColorReset)
	fmt.Print("Paste your Google API Key: ")

	reader := bufio.NewReader(os.Stdin)
	key, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read API key from input: %w", err)
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return "", fmt.Errorf("API key cannot be empty")
	}

	// Save to ~/.tertic/config.env for future use
	if home != "" {
		terticDir := filepath.Join(home, ".tertic")
		_ = os.MkdirAll(terticDir, 0755)
		configFile := filepath.Join(terticDir, "config.env")

		f, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			_, _ = f.WriteString(fmt.Sprintf("\nexport GEMINI_API_KEY=\"%s\"\n", key))
			_ = f.Close()
			fmt.Printf("%s✓ API key saved! You won't have to enter it again.%s\n\n", ColorGreen, ColorReset)
		}
	}

	return key, nil
}

func runAgentTurn(client *gemini.Client, history *[]gemini.Content, userPrompt string) error {
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
			return fmt.Errorf("Gemini API error: %w", err)
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
			fmt.Printf("\n%s%s✓ Done!%s\n", ColorBold, ColorGreen, ColorReset)
			break
		}
	}
	return nil
}

func main() {
	apiKey, err := resolveAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	// Allow model override via GEMINI_MODEL env var if desired
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = gemini.DefaultModel
	}
	client := gemini.NewClient(apiKey, model)
	var history []gemini.Content

	// If prompt passed directly in command line (e.g. tertic "build an app")
	if len(os.Args) > 1 {
		userPrompt := strings.Join(os.Args[1:], " ")
		fmt.Printf("\n%s%s🚀 Tertic Agent executing:%s %s\n\n", ColorBold, ColorCyan, ColorReset, userPrompt)
		if err := runAgentTurn(client, &history, userPrompt); err != nil {
			fmt.Fprintf(os.Stderr, "\n%sError: %v%s\n", ColorRed, err, ColorReset)
		}
		return
	}

	// Interactive Mode: when user just runs `tertic`
	fmt.Print(ColorCyan, Banner, ColorReset)
	fmt.Printf("%sInteractive Agent Mode%s (type your instructions or 'exit' to quit)\n\n", ColorBold, ColorReset)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s%stertic > %s", ColorBold, ColorGreen, ColorReset)
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

		if err := runAgentTurn(client, &history, prompt); err != nil {
			fmt.Fprintf(os.Stderr, "%sError: %v%s\n\n", ColorRed, err, ColorReset)
		}
		fmt.Println()
	}
}
