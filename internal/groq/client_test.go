package groq

import (
	"testing"
)

func TestGroqClientDefaults(t *testing.T) {
	c := NewClient("test-key", "")
	if c.Model != DefaultModel {
		t.Errorf("expected %s, got %s", DefaultModel, c.Model)
	}
	if c.APIKey != "test-key" {
		t.Errorf("expected test-key, got %s", c.APIKey)
	}
}
