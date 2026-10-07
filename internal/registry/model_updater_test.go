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

func TestAddLocalModels_FillsCodexTiersUntilUpstreamListsModel(t *testing.T) {
	data := &staticModelsJSON{CodexPro: []*ModelInfo{{ID: "gpt-6.1-sol"}}}
	addLocalModels(data)
	for name, section := range map[string][]*ModelInfo{"codex-team": data.CodexTeam, "codex-plus": data.CodexPlus, "codex-pro": data.CodexPro} {
		if got := idCount(section, "gpt-6.1-terra"); got != 1 {
			t.Fatalf("%s without gpt-6.1-terra: got %d entries, want 1", name, got)
		}
	}
	if got := idCount(data.CodexFree, "gpt-6.1-terra"); got != 0 {
		t.Fatalf("codex-free: got %d gpt-6.1-terra entries, want 0", got)
	}

	upstream := &ModelInfo{ID: "gpt-6.1-terra", DisplayName: "upstream"}
	data = &staticModelsJSON{CodexPro: []*ModelInfo{upstream}}
	addLocalModels(data)
	if got := idCount(data.CodexPro, "gpt-6.1-terra"); got != 1 || data.CodexPro[0] != upstream {
		t.Fatalf("codex-pro with gpt-6.1-terra: got %d entries, want upstream's entry only", got)
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
	if got := idCount(getModels().CodexPro, "gpt-6.1-terra"); got != 1 {
		t.Fatalf("published catalog: got %d gpt-6.1-terra entries, want 1", got)
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
