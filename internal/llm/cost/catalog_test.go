package cost

import (
	"testing"
)

func TestLLMModelsCatalog_GetModel(t *testing.T) {
	tests := []struct {
		name     string
		lookup   string
		wantOk   bool
		wantName string
	}{
		{"canonical name", "gpt-4o", true, "gpt-4o"},
		{"alias", "gpt-4o-2024-05-13", true, "gpt-4o"},
		{"unknown", "nonexistent-model", false, ""},
	}
	cat := NewDefaultCatalog()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, ok := cat.GetModel(tt.lookup)
			if ok != tt.wantOk {
				t.Errorf("expected ok=%v, got %v", tt.wantOk, ok)
			}
			if ok && model.Name != tt.wantName {
				t.Errorf("expected name=%q, got %q", tt.wantName, model.Name)
			}
		})
	}
}

func TestLLMModelsCatalog_ListModels(t *testing.T) {
	cat := NewDefaultCatalog()
	models := cat.ListModels()
	if len(models) == 0 {
		t.Fatal("expected at least one model in catalog")
	}
	found := false
	for _, m := range models {
		if m.Name == "gpt-4o" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find model 'gpt-4o' in catalog")
	}
}
