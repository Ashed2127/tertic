package commands

import (
	"sort"
	"strings"
)

// Category represents a logical grouping of slash commands
type Category string

const (
	CategoryGeneral     Category = "General"
	CategoryModels      Category = "Models and Providers"
	CategoryAgents      Category = "Agents"
	CategoryVoice       Category = "Voice and Input"
	CategoryConfig      Category = "Configuration"
	CategorySession     Category = "Session Management"
	CategoryDeveloper   Category = "Developer Tools"
)

// AppState represents the current state of Tertic for context-aware filtering
type AppState struct {
	IsRecording bool
	IsExecuting bool
	VoiceActive bool
}

// Command defines a slash command in the centralized registry
type Command struct {
	Name        string   // e.g. "/help"
	Aliases     []string // e.g. ["help"]
	Description string   // Short 1-line description
	Category    Category // Logical category
	Priority    int      // Higher number = higher ranking when suggested
	IsAvailable func(state AppState) bool // Optional context-aware filter
}

// Registry stores and queries all registered slash commands
type Registry struct {
	commands map[string]Command
	ordered  []string
}

// Global default registry
var defaultRegistry = NewRegistry()

// NewRegistry creates a new Command registry
func NewRegistry() *Registry {
	r := &Registry{
		commands: make(map[string]Command),
		ordered:  make([]string, 0),
	}
	r.registerDefaults()
	return r
}

// Default returns the default global command registry
func Default() *Registry {
	return defaultRegistry
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) {
	if !strings.HasPrefix(cmd.Name, "/") {
		cmd.Name = "/" + cmd.Name
	}
	r.commands[cmd.Name] = cmd
	r.ordered = append(r.ordered, cmd.Name)
}

// ListAll returns all registered commands in registration order
func (r *Registry) ListAll() []Command {
	list := make([]Command, 0, len(r.commands))
	seen := make(map[string]bool)
	for _, name := range r.ordered {
		if !seen[name] {
			seen[name] = true
			if cmd, ok := r.commands[name]; ok {
				list = append(list, cmd)
			}
		}
	}
	return list
}

// Filter returns matching commands given an entered query and current application state
func (r *Registry) Filter(query string, state AppState) []Command {
	query = strings.TrimSpace(query)
	cleanQuery := strings.ToLower(query)
	if !strings.HasPrefix(cleanQuery, "/") && cleanQuery != "" {
		cleanQuery = "/" + cleanQuery
	}

	var matches []Command
	for _, cmd := range r.ListAll() {
		if cmd.IsAvailable != nil && !cmd.IsAvailable(state) {
			continue
		}

		cmdLower := strings.ToLower(cmd.Name)
		if cleanQuery == "" || cleanQuery == "/" {
			matches = append(matches, cmd)
			continue
		}

		// Exact prefix match on command name
		if strings.HasPrefix(cmdLower, cleanQuery) {
			matches = append(matches, cmd)
			continue
		}

		// Check aliases
		aliasMatched := false
		trimmedQuery := strings.TrimPrefix(cleanQuery, "/")
		for _, alias := range cmd.Aliases {
			if strings.HasPrefix(strings.ToLower(alias), trimmedQuery) {
				matches = append(matches, cmd)
				aliasMatched = true
				break
			}
		}
		if aliasMatched {
			continue
		}

		// Substring match on name or description if prefix didn't match
		if strings.Contains(cmdLower, cleanQuery) || strings.Contains(strings.ToLower(cmd.Description), trimmedQuery) {
			matches = append(matches, cmd)
		}
	}

	// Sort context-aware: higher priority first, then matching name length / alphabetical
	sort.SliceStable(matches, func(i, j int) bool {
		if state.IsRecording {
			if matches[i].Name == "/stop" {
				return true
			}
			if matches[j].Name == "/stop" {
				return false
			}
		}
		if matches[i].Priority != matches[j].Priority {
			return matches[i].Priority > matches[j].Priority
		}
		// Exact prefix match prioritized
		iPrefix := strings.HasPrefix(strings.ToLower(matches[i].Name), cleanQuery)
		jPrefix := strings.HasPrefix(strings.ToLower(matches[j].Name), cleanQuery)
		if iPrefix && !jPrefix {
			return true
		}
		if !iPrefix && jPrefix {
			return false
		}
		return matches[i].Name < matches[j].Name
	})

	return matches
}

func (r *Registry) registerDefaults() {
	cmds := []Command{
		// General
		{
			Name:        "/help",
			Aliases:     []string{"h", "?"},
			Description: "Show help and available commands",
			Category:    CategoryGeneral,
			Priority:    10,
		},
		{
			Name:        "/clear",
			Aliases:     []string{"cls"},
			Description: "Clear terminal",
			Category:    CategoryGeneral,
			Priority:    5,
		},
		{
			Name:        "/exit",
			Aliases:     []string{"quit", "q"},
			Description: "Exit Tertic",
			Category:    CategoryGeneral,
			Priority:    1,
		},

		// Models and Providers
		{
			Name:        "/model",
			Aliases:     []string{"m"},
			Description: "Switch AI model",
			Category:    CategoryModels,
			Priority:    9,
		},
		{
			Name:        "/models",
			Description: "List configured models",
			Category:    CategoryModels,
			Priority:    8,
		},
		{
			Name:        "/provider",
			Aliases:     []string{"p"},
			Description: "Switch AI provider (OpenRouter/Gemini/Ollama)",
			Category:    CategoryModels,
			Priority:    8,
		},

		// Agents
		{
			Name:        "/agents",
			Aliases:     []string{"ag"},
			Description: "Manage AI agents",
			Category:    CategoryAgents,
			Priority:    8,
		},
		{
			Name:        "/status",
			Description: "Show session status",
			Category:    CategoryAgents,
			Priority:    8,
		},

		// Voice and Input
		{
			Name:        "/voice",
			Aliases:     []string{"v"},
			Description: "Activate voice input",
			Category:    CategoryVoice,
			Priority:    7,
		},
		{
			Name:        "/text",
			Description: "Switch to text input",
			Category:    CategoryVoice,
			Priority:    6,
		},
		{
			Name:        "/input",
			Description: "Change input mode",
			Category:    CategoryVoice,
			Priority:    6,
		},
		{
			Name:        "/listen",
			Aliases:     []string{"record"},
			Description: "Start voice recording",
			Category:    CategoryVoice,
			Priority:    7,
			IsAvailable: func(s AppState) bool {
				return !s.IsRecording
			},
		},
		{
			Name:        "/stop",
			Description: "Stop recording",
			Category:    CategoryVoice,
			Priority:    15, // High priority when recording
			IsAvailable: func(s AppState) bool {
				return true
			},
		},

		// Configuration
		{
			Name:        "/config",
			Aliases:     []string{"cfg"},
			Description: "Configure Tertic",
			Category:    CategoryConfig,
			Priority:    5,
		},
		{
			Name:        "/theme",
			Description: "Change terminal theme",
			Category:    CategoryConfig,
			Priority:    6,
		},

		// Session Management
		{
			Name:        "/history",
			Description: "View conversation history",
			Category:    CategorySession,
			Priority:    5,
		},
		{
			Name:        "/sessions",
			Aliases:     []string{"session"},
			Description: "List or resume saved chat sessions",
			Category:    CategorySession,
			Priority:    6,
		},
		{
			Name:        "/save",
			Description: "Save current conversation session",
			Category:    CategorySession,
			Priority:    5,
		},
		{
			Name:        "/context",
			Description: "View project context",
			Category:    CategorySession,
			Priority:    5,
		},

		// Developer Tools
		{
			Name:        "/verbose",
			Aliases:     []string{"process"},
			Description: "Toggle detailed output",
			Category:    CategoryDeveloper,
			Priority:    4,
		},
	}

	for _, c := range cmds {
		r.Register(c)
	}
}
