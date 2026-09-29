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
	if got := idCount(data.Claude, "claude-sonnet-5-5"); got != 1 {
		t.Fatalf("catalog without claude-sonnet-5-5: got %d entries, want 1", got)
	}

	upstream := &ModelInfo{ID: "claude-sonnet-5-5", DisplayName: "upstream"}
	data = &staticModelsJSON{Claude: []*ModelInfo{upstream}}
	addLocalModels(data)
	if got := idCount(data.Claude, "claude-sonnet-5-5"); got != 1 || data.Claude[0] != upstream {
		t.Fatalf("catalog with claude-sonnet-5-5: got %d entries, want upstream's entry only", got)
	}
}

func idCount(section []*ModelInfo, id string) int {
	n := 0
	for _, m := range section {
		if m.ID == id {
			n++
		}
	}
	return n
}

func TestAddLocalModels_FillsCodexTiersUntilUpstreamListsModel(t *testing.T) {
	data := &staticModelsJSON{CodexPro: []*ModelInfo{{ID: "gpt-6-sol"}}}
	addLocalModels(data)
	for name, section := range map[string][]*ModelInfo{"codex-team": data.CodexTeam, "codex-plus": data.CodexPlus, "codex-pro": data.CodexPro} {
		if got := idCount(section, "gpt-6.1-sol"); got != 1 {
			t.Fatalf("%s without gpt-6.1-sol: got %d entries, want 1", name, got)
		}
	}
	if got := idCount(data.CodexFree, "gpt-6.1-sol"); got != 0 {
		t.Fatalf("codex-free: got %d gpt-6.1-sol entries, want 0", got)
	}

	upstream := &ModelInfo{ID: "gpt-6.1-sol", DisplayName: "upstream"}
	data = &staticModelsJSON{CodexPro: []*ModelInfo{upstream}}
	addLocalModels(data)
	if got := idCount(data.CodexPro, "gpt-6.1-sol"); got != 1 || data.CodexPro[0] != upstream {
		t.Fatalf("codex-pro with gpt-6.1-sol: got %d entries, want upstream's entry only", got)
	}
}
