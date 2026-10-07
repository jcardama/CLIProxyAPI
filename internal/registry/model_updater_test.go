package registry

import (
	"testing"
)

func TestDetectChangedProviders_CodexConfigurationUpdate(t *testing.T) {
	oldData := &staticModelsJSON{
		CodexFree: []*ModelInfo{{ID: "gpt-6-luna"}},
	}
	newData := &staticModelsJSON{
		CodexFree: []*ModelInfo{{ID: "gpt-6-luna", SupportConfigurationUpdate: true}},
	}

	changed := detectChangedProviders(oldData, newData)
	if len(changed) != 1 || changed[0] != "codex" {
		t.Fatalf("configuration_update-only change: got providers %v, want [codex]", changed)
	}
}

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
	data := &staticModelsJSON{Claude: []*ModelInfo{{ID: "claude-sonnet-5-5"}}}
	addLocalModels(data)
	if got := idCount(data.Claude, "claude-haiku-5-5"); got != 1 {
		t.Fatalf("catalog without claude-haiku-5-5: got %d entries, want 1", got)
	}

	upstream := &ModelInfo{ID: "claude-haiku-5-5", DisplayName: "upstream"}
	data = &staticModelsJSON{Claude: []*ModelInfo{upstream}}
	addLocalModels(data)
	if got := idCount(data.Claude, "claude-haiku-5-5"); got != 1 || data.Claude[0] != upstream {
		t.Fatalf("catalog with claude-haiku-5-5: got %d entries, want upstream's entry only", got)
	}
}

func TestPublishCatalogBytes_KeepsLocalModelsAcrossRemoteRefresh(t *testing.T) {
	previous := getModels()
	t.Cleanup(func() {
		modelsCatalogStore.mu.Lock()
		modelsCatalogStore.data = previous
		modelsCatalogStore.mu.Unlock()
	})

	if _, err := publishCatalogBytes(embeddedModelsJSON); err != nil {
		t.Fatalf("publish embedded catalog: %v", err)
	}
	if got := idCount(getModels().Claude, "claude-haiku-5-5"); got != 1 {
		t.Fatalf("published catalog: got %d claude-haiku-5-5 entries, want 1", got)
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
