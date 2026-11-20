package ui

import (
	"bufio"
	"strings"
	"testing"

	"github.com/Ashed2127/tertic/internal/commands"
	"github.com/Ashed2127/tertic/internal/theme"
)

func TestFormatActionHumanReadable(t *testing.T) {
	tests := []struct {
		name     string
		tool     string
		args     map[string]interface{}
		expected string
	}{
		{
			name:     "read file",
			tool:     "read_file",
			args:     map[string]interface{}{"path": "src/main.go"},
			expected: "Reading: main.go",
		},
		{
			name:     "write file",
			tool:     "write_file",
			args:     map[string]interface{}{"path": "calculator.html"},
			expected: "Writing: calculator.html",
		},
		{
			name:     "execute test",
			tool:     "execute_bash",
			args:     map[string]interface{}{"command": "go test ./..."},
			expected: "Testing & verifying...",
		},
		{
			name:     "install package",
			tool:     "execute_bash",
			args:     map[string]interface{}{"command": "npm install"},
			expected: "Installing packages...",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := FormatActionHumanReadable(tc.tool, tc.args)
			if res != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, res)
			}
		})
	}
}

func TestFormatInlineMarkdown(t *testing.T) {
	text := "Check `src/index.js` for **key details**."
	formatted := formatInlineMarkdown(text, theme.DarkTheme())
	if !strings.Contains(formatted, "key details") || !strings.Contains(formatted, "src/index.js") {
		t.Fatalf("formatting lost text content: %s", formatted)
	}
}

func TestInteractiveSlashReaderFallback(t *testing.T) {
	prompt := "t ❯ "
	reader := NewInteractiveSlashReader(prompt, commands.AppState{})
	input := strings.NewReader("echo hello\n")
	buf := bufio.NewReader(input)

	// In non-terminal testing context, ReadLine falls back to standard line reading safely
	res, err := reader.ReadLine(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != "echo hello" {
		t.Fatalf("expected 'echo hello', got %q", res)
	}
}
