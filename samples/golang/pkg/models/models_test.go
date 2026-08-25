package models

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadModels(t *testing.T) {
	// Create a temporary JSON file for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "models.json")
	content := `{
		"providers": [
			{
				"provider": "test-provider",
				"displayName": "Test Provider",
				"models": [
					{"modelId": "model-a", "displayName": "Model A", "shortName": "A"}
				]
			}
		]
	}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := LoadModels(configPath)
	if err != nil {
		t.Fatalf("LoadModels returned error: %v", err)
	}
	if len(cfg.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(cfg.Providers))
	}
	if cfg.Providers[0].Provider != "test-provider" {
		t.Errorf("expected provider 'test-provider', got %q", cfg.Providers[0].Provider)
	}
	if len(cfg.Providers[0].Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(cfg.Providers[0].Models))
	}
	if cfg.Providers[0].Models[0].ModelID != "model-a" {
		t.Errorf("expected modelId 'model-a', got %q", cfg.Providers[0].Models[0].ModelID)
	}
}

func TestLoadModels_FileNotFound(t *testing.T) {
	_, err := LoadModels("/nonexistent/path/models.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestParseModelFlag_Valid(t *testing.T) {
	provider, modelID, err := ParseModelFlag("qwen:qwen3.8-max")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != "qwen" {
		t.Errorf("expected provider 'qwen', got %q", provider)
	}
	if modelID != "qwen3.8-max" {
		t.Errorf("expected modelID 'qwen3.8-max', got %q", modelID)
	}
}

func TestParseModelFlag_NoColon(t *testing.T) {
	_, _, err := ParseModelFlag("nocolon")
	if err == nil {
		t.Fatal("expected error for input without colon")
	}
}

func TestParseModelFlag_EmptyProvider(t *testing.T) {
	_, _, err := ParseModelFlag(":modelId")
	if err == nil {
		t.Fatal("expected error for empty provider")
	}
}

func TestParseModelFlag_EmptyModelID(t *testing.T) {
	_, _, err := ParseModelFlag("provider:")
	if err == nil {
		t.Fatal("expected error for empty modelId")
	}
}

func TestValidateModel(t *testing.T) {
	cfg := &ModelsConfig{
		Providers: []Provider{
			{
				Provider:    "qwen",
				DisplayName: "Qwen",
				Models: []Model{
					{ModelID: "qwen3.8-max", DisplayName: "Qwen 3.8-Max", ShortName: "3.8-Max"},
				},
			},
		},
	}

	if !ValidateModel(cfg, "qwen", "qwen3.8-max") {
		t.Error("expected valid model to return true")
	}
	if ValidateModel(cfg, "qwen", "nonexistent") {
		t.Error("expected invalid modelId to return false")
	}
	if ValidateModel(cfg, "unknown", "qwen3.8-max") {
		t.Error("expected invalid provider to return false")
	}
}

func TestBuildModelJSON(t *testing.T) {
	result := BuildModelJSON("qwen", "qwen3.8-max")
	expected := `{"provider":"qwen","modelID":"qwen3.8-max"}`
	if result != expected {
		t.Errorf("BuildModelJSON mismatch:\n  got:  %s\n  want: %s", result, expected)
	}
}

func TestBuildConfigWithModel(t *testing.T) {
	result := BuildConfigWithModel("qwen", "qwen3.8-max")
	expected := `{"model":{"provider":"qwen","modelID":"qwen3.8-max"},"disableThreadData":false}`
	if result != expected {
		t.Errorf("BuildConfigWithModel mismatch:\n  got:  %s\n  want: %s", result, expected)
	}
}

func TestDisplayMenu(t *testing.T) {
	cfg := &ModelsConfig{
		Providers: []Provider{
			{
				Provider:    "qwen",
				DisplayName: "Qwen",
				Models: []Model{
					{ModelID: "qwen3.8-max", DisplayName: "Qwen 3.8-Max"},
					{ModelID: "qwen3.7-max", DisplayName: "Qwen 3.7-Max"},
				},
			},
			{
				Provider:    "glm",
				DisplayName: "GLM",
				Models: []Model{
					{ModelID: "glm-5.2", DisplayName: "GLM 5.2"},
				},
			},
		},
	}

	menu := DisplayMenu(cfg)
	expected := "可选模型列表：\n" +
		"  1. Qwen: Qwen 3.8-Max\n" +
		"  2. Qwen: Qwen 3.7-Max\n" +
		"  3. GLM: GLM 5.2\n"
	if menu != expected {
		t.Errorf("DisplayMenu mismatch:\n  got:\n%s\n  want:\n%s", menu, expected)
	}
}

func TestGetModelByIndex_Valid(t *testing.T) {
	cfg := &ModelsConfig{
		Providers: []Provider{
			{
				Provider:    "qwen",
				DisplayName: "Qwen",
				Models: []Model{
					{ModelID: "qwen3.8-max", DisplayName: "Qwen 3.8-Max"},
					{ModelID: "qwen3.7-max", DisplayName: "Qwen 3.7-Max"},
				},
			},
			{
				Provider:    "glm",
				DisplayName: "GLM",
				Models: []Model{
					{ModelID: "glm-5.2", DisplayName: "GLM 5.2"},
				},
			},
		},
	}

	// Index 0 → first model (qwen/qwen3.8-max)
	provider, modelID, err := GetModelByIndex(cfg, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != "qwen" || modelID != "qwen3.8-max" {
		t.Errorf("index 0: expected qwen/qwen3.8-max, got %s/%s", provider, modelID)
	}

	// Index 2 → third model (glm/glm-5.2)
	provider, modelID, err = GetModelByIndex(cfg, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider != "glm" || modelID != "glm-5.2" {
		t.Errorf("index 2: expected glm/glm-5.2, got %s/%s", provider, modelID)
	}
}

func TestGetModelByIndex_OutOfRange(t *testing.T) {
	cfg := &ModelsConfig{
		Providers: []Provider{
			{
				Provider: "qwen",
				Models: []Model{
					{ModelID: "qwen3.8-max"},
				},
			},
		},
	}

	_, _, err := GetModelByIndex(cfg, 1)
	if err == nil {
		t.Fatal("expected error for out-of-range index")
	}

	_, _, err = GetModelByIndex(cfg, -1)
	if err == nil {
		t.Fatal("expected error for negative index")
	}
}

func TestLoadModels_RealConfig(t *testing.T) {
	// Try to load the real config/models.json from project root
	// Walk up from the test working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("cannot get working directory")
	}

	dir := cwd
	var configPath string
	for {
		candidate := filepath.Join(dir, "config", "models.json")
		if _, err := os.Stat(candidate); err == nil {
			configPath = candidate
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	if configPath == "" {
		t.Skip("config/models.json not found in parent directories")
	}

	cfg, err := LoadModels(configPath)
	if err != nil {
		t.Fatalf("LoadModels with real config failed: %v", err)
	}
	if len(cfg.Providers) == 0 {
		t.Fatal("expected at least one provider in real config")
	}
	// Verify the first provider is qwen with qwen3.8-max
	if cfg.Providers[0].Provider != "qwen" {
		t.Errorf("expected first provider to be 'qwen', got %q", cfg.Providers[0].Provider)
	}
	if len(cfg.Providers[0].Models) == 0 {
		t.Fatal("expected qwen provider to have models")
	}
	if cfg.Providers[0].Models[0].ModelID != "qwen3.8-max" {
		t.Errorf("expected first model to be 'qwen3.8-max', got %q", cfg.Providers[0].Models[0].ModelID)
	}
}
