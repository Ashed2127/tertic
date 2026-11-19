package tools

import "testing"

func TestToolValidator(t *testing.T) {
	v := NewToolValidator()

	// Empty name
	if err := v.Validate("", nil); err == nil {
		t.Errorf("expected error for empty tool name")
	}

	// Unknown tool
	if err := v.Validate("delete_database", nil); err == nil {
		t.Errorf("expected error for unknown tool")
	}

	// write_file without path
	if err := v.Validate("write_file", map[string]interface{}{"content": "hello"}); err == nil {
		t.Errorf("expected error for write_file without path")
	}

	// write_file valid
	if err := v.Validate("write_file", map[string]interface{}{"path": "test.txt", "content": "hello"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// dangerous bash
	if err := v.Validate("execute_bash", map[string]interface{}{"command": "rm -rf / --no-preserve-root"}); err == nil {
		t.Errorf("expected error for dangerous command")
	}

	// valid bash
	if err := v.Validate("execute_bash", map[string]interface{}{"command": "ls -la"}); err != nil {
		t.Errorf("unexpected error for valid bash: %v", err)
	}
}
