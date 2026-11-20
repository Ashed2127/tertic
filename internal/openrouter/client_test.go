package openrouter

import (
	"testing"
)

func TestFallbackModelList(t *testing.T) {
	if len(FallbackFreeModels) == 0 {
		t.Fatal("expected at least one fallback model")
	}
	for _, m := range FallbackFreeModels {
		if m == "" {
			t.Errorf("empty model in FallbackFreeModels")
		}
	}
}
