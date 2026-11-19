package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MarshalJSON helper for arguments
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// UnmarshalJSON helper for arguments
func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ToolValidator validates tool calls before execution to protect Android environment
type ToolValidator struct {
	AllowedTools map[string]bool
	DangerousCmds []string
}

func NewToolValidator() *ToolValidator {
	return &ToolValidator{
		AllowedTools: map[string]bool{
			"execute_bash": true,
			"write_file":   true,
			"read_file":    true,
			"edit_file":    true,
			"list_dir":     true,
		},
		DangerousCmds: []string{
			"rm -rf /",
			":(){ :|:& };:",
			"mkfs",
			"> /dev/sda",
			"dd if=/dev/zero",
		},
	}
}

// Validate inspects tool call and arguments, rejecting invalid or unsafe operations
func (v *ToolValidator) Validate(name string, args map[string]interface{}) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if !v.AllowedTools[name] {
		return fmt.Errorf("unauthorized or unknown tool: %s", name)
	}

	if args == nil {
		args = make(map[string]interface{})
	}

	switch name {
	case "execute_bash":
		cmd, ok := args["command"].(string)
		if !ok || strings.TrimSpace(cmd) == "" {
			return fmt.Errorf("missing or empty 'command' argument for execute_bash")
		}
		// Security check: reject destructive root wiper commands
		trimmedCmd := strings.TrimSpace(cmd)
		for _, dangerous := range v.DangerousCmds {
			if strings.Contains(trimmedCmd, dangerous) {
				return fmt.Errorf("command rejected: dangerous operation detected (%s)", dangerous)
			}
		}

	case "write_file":
		path, ok1 := args["path"].(string)
		if !ok1 || strings.TrimSpace(path) == "" {
			return fmt.Errorf("missing or empty 'path' argument for write_file")
		}
		_, ok2 := args["content"].(string)
		if !ok2 {
			return fmt.Errorf("missing 'content' argument for write_file")
		}

	case "read_file":
		path, ok := args["path"].(string)
		if !ok || strings.TrimSpace(path) == "" {
			return fmt.Errorf("missing or empty 'path' argument for read_file")
		}

	case "edit_file":
		path, ok1 := args["path"].(string)
		if !ok1 || strings.TrimSpace(path) == "" {
			return fmt.Errorf("missing or empty 'path' argument for edit_file")
		}
		target, ok2 := args["target"].(string)
		if !ok2 || target == "" {
			return fmt.Errorf("missing or empty 'target' argument for edit_file")
		}
		_, ok3 := args["replacement"].(string)
		if !ok3 {
			return fmt.Errorf("missing 'replacement' argument for edit_file")
		}

	case "list_dir":
		// path is optional, defaults to "."
	}

	return nil
}
