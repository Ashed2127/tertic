package theme

import (
	"testing"
)

func TestThemes(t *testing.T) {
	themes := AvailableThemes()
	if len(themes) < 4 {
		t.Fatalf("expected at least 4 themes, got %d", len(themes))
	}

	for _, name := range themes {
		ok := SetTheme(name)
		if !ok {
			t.Fatalf("failed to set theme %s", name)
		}
		cur := Current()
		if cur.Name != name {
			t.Fatalf("expected theme %s, got %s", name, cur.Name)
		}
	}

	// Invalid theme
	if SetTheme("nonexistent") {
		t.Fatalf("expected false for nonexistent theme")
	}

	// Reset to dark
	SetTheme("dark")
}
