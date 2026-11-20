package mistral

import (
	"testing"
)

func TestMistralClientDefaults(t *testing.T) {
	c := NewClient("test-mistral-key", "")
	if c.Model != DefaultModel {
		t.Errorf("expected %s, got %s", DefaultModel, c.Model)
	}
	if c.APIKey != "test-mistral-key" {
		t.Errorf("expected test-mistral-key, got %s", c.APIKey)
	}
}
