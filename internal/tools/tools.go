package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ToolDefinition represents a Gemini function declaration
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// GetToolDefinitions returns the Gemini-compatible function declarations
func GetToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "execute_bash",
			Description: "Execute a bash command in Termux/Linux workspace with a timeout. Captures stdout and stderr.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The exact bash command line to execute.",
					},
				},
				"required": []string{"command"},
			},
		},
		{
			Name:        "write_file",
			Description: "Create a new file or completely overwrite an existing file with the provided content.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The relative or absolute path of the file to write.",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The entire text content to write into the file.",
					},
				},
				"required": []string{"path", "content"},
			},
		},
		{
			Name:        "read_file",
			Description: "Read the contents of a file.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The path to the file to read.",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "edit_file",
			Description: "Perform an exact search and replace inside a file to modify parts of the code.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The path to the file to modify.",
					},
					"target": map[string]interface{}{
						"type":        "string",
						"description": "The exact string segment to replace.",
					},
					"replacement": map[string]interface{}{
						"type":        "string",
						"description": "The new replacement string.",
					},
				},
				"required": []string{"path", "target", "replacement"},
			},
		},
		{
			Name:        "list_dir",
			Description: "List files and directories in a given path.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The directory path to list (defaults to current dir if empty).",
					},
				},
			},
		},
	}
}

// ExecuteTool dispatches the function call to the corresponding local implementation
func ExecuteTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	case "execute_bash":
		cmdStr, ok := args["command"].(string)
		if !ok || strings.TrimSpace(cmdStr) == "" {
			return "", fmt.Errorf("command parameter is required")
		}
		return RunBash(cmdStr)

	case "write_file":
		path, ok1 := args["path"].(string)
		content, ok2 := args["content"].(string)
		if !ok1 || !ok2 {
			return "", fmt.Errorf("path and content parameters are required")
		}
		return WriteFile(path, content)

	case "read_file":
		path, ok := args["path"].(string)
		if !ok || strings.TrimSpace(path) == "" {
			return "", fmt.Errorf("path parameter is required")
		}
		return ReadFile(path)

	case "edit_file":
		path, ok1 := args["path"].(string)
		target, ok2 := args["target"].(string)
		replacement, ok3 := args["replacement"].(string)
		if !ok1 || !ok2 || !ok3 {
			return "", fmt.Errorf("path, target, and replacement parameters are required")
		}
		return EditFile(path, target, replacement)

	case "list_dir":
		path := "."
		if p, ok := args["path"].(string); ok && strings.TrimSpace(p) != "" {
			path = p
		}
		return ListDir(path)

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

// RunBash runs a command with a 60-second timeout
func RunBash(command string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	output := stdoutBuf.String()
	if stderrBuf.Len() > 0 {
		if len(output) > 0 {
			output += "\n--- STDERR ---\n"
		}
		output += stderrBuf.String()
	}

	// Limit output to prevent flooding context window (up to 30KB)
	if len(output) > 30000 {
		output = output[:15000] + "\n...[truncated]...\n" + output[len(output)-15000:]
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return output + "\n[Error: Command timed out after 60 seconds]", nil
		}
		return output + fmt.Sprintf("\n[Exit code: %v]", err), nil
	}

	if strings.TrimSpace(output) == "" {
		return "[Command completed with no output]", nil
	}

	return output, nil
}

// WriteFile writes content to disk, creating parent directories if needed
func WriteFile(path string, content string) (string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}

// ReadFile reads up to 50KB of a file
func ReadFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	buf := make([]byte, 50000)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	content := string(buf[:n])
	if n == 50000 {
		content += "\n...[File truncated at 50KB]..."
	}
	return content, nil
}

// EditFile searches for exact target substring and replaces it
func EditFile(path string, target string, replacement string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)
	if !strings.Contains(content, target) {
		return "", fmt.Errorf("target string not found in %s", path)
	}

	count := strings.Count(content, target)
	newContent := strings.Replace(content, target, replacement, 1)

	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write modified file: %w", err)
	}

	return fmt.Sprintf("Successfully replaced target in %s (found %d occurrences, replaced first occurrence)", path, count), nil
}

// ListDir lists directory items
func ListDir(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var lines []string
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if entry.IsDir() {
			lines = append(lines, fmt.Sprintf("[DIR]  %s/", entry.Name()))
		} else {
			lines = append(lines, fmt.Sprintf("[FILE] %-30s (%d bytes)", entry.Name(), info.Size()))
		}
	}
	return strings.Join(lines, "\n"), nil
}
