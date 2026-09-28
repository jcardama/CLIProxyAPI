package registry

import (
	"testing"
)

func TestDetectChangedProviders_KimiAliases(t *testing.T) {
	oldData := &staticModelsJSON{
		Kimi: []*ModelInfo{{ID: "kimi-k2"}},
	}
	newData := &staticModelsJSON{
		Kimi: []*ModelInfo{{ID: "kimi-k2"}, {ID: "kimi-k3"}},
	}

	changed := detectChangedProviders(oldData, newData)
	expected := map[string]bool{
		"kimi":     false,
		"kimi-ai":  false,
		"kimi.ai":  false,
		"kimi.com": false,
	}

	for _, p := range changed {
		if _, ok := expected[p]; ok {
			expected[p] = true
		}
	}

	for p, found := range expected {
		if !found {
			t.Errorf("expected changed provider %q to be reported, got %v", p, changed)
		}
	}
}

func TestAddLocalModels_FillsGapUntilUpstreamListsModel(t *testing.T) {
	data := &staticModelsJSON{Claude: []*ModelInfo{{ID: "claude-opus-5-5"}}}
	addLocalModels(data)
	if got := claudeIDCount(data, "claude-sonnet-5-5"); got != 1 {
		t.Fatalf("catalog without claude-sonnet-5-5: got %d entries, want 1", got)
	}

	upstream := &ModelInfo{ID: "claude-sonnet-5-5", DisplayName: "upstream"}
	data = &staticModelsJSON{Claude: []*ModelInfo{upstream}}
	addLocalModels(data)
	if got := claudeIDCount(data, "claude-sonnet-5-5"); got != 1 || data.Claude[0] != upstream {
		t.Fatalf("catalog with claude-sonnet-5-5: got %d entries, want upstream's entry only", got)
	}
}

func claudeIDCount(data *staticModelsJSON, id string) int {
	n := 0
	for _, m := range data.Claude {
		if m.ID == id {
			n++
		}
	}
	return n
}
