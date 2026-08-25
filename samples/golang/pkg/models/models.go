package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Model represents a single model entry.
type Model struct {
	ModelID     string `json:"modelId"`
	DisplayName string `json:"displayName"`
	ShortName   string `json:"shortName"`
}

// Provider represents a model provider with its models.
type Provider struct {
	Provider    string  `json:"provider"`
	DisplayName string  `json:"displayName"`
	Models      []Model `json:"models"`
}

// ModelsConfig is the top-level config structure.
type ModelsConfig struct {
	Providers []Provider `json:"providers"`
}

// LoadModels reads and deserializes the JSON config file.
func LoadModels(configPath string) (*ModelsConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read models config: %w", err)
	}
	var cfg ModelsConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse models config: %w", err)
	}
	return &cfg, nil
}

// FindConfigPath locates the config/models.json file.
// Priority: STAROPS_MODELS_CONFIG env > walk up from cwd > default.
func FindConfigPath() string {
	if envPath := os.Getenv("STAROPS_MODELS_CONFIG"); envPath != "" {
		return envPath
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "config/models.json"
	}

	dir := cwd
	for {
		candidate := filepath.Join(dir, "config", "models.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "config/models.json"
}

// ParseModelFlag parses "provider:modelId" format.
func ParseModelFlag(flag string) (provider, modelID string, err error) {
	parts := strings.SplitN(flag, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid model flag %q: format should be provider:modelId", flag)
	}
	return parts[0], parts[1], nil
}

// ValidateModel checks if the provider+modelId combination exists in config.
func ValidateModel(cfg *ModelsConfig, provider, modelID string) bool {
	for _, p := range cfg.Providers {
		if p.Provider == provider {
			for _, m := range p.Models {
				if m.ModelID == modelID {
					return true
				}
			}
		}
	}
	return false
}

// BuildModelJSON generates a JSON string: {"provider":"x","modelID":"y"}.
func BuildModelJSON(provider, modelID string) string {
	v := struct {
		Provider string `json:"provider"`
		ModelID  string `json:"modelID"`
	}{
		Provider: provider,
		ModelID:  modelID,
	}
	data, _ := json.Marshal(v)
	return string(data)
}

// BuildConfigWithModel generates a full config JSON string with model and disableThreadData.
func BuildConfigWithModel(provider, modelID string) string {
	v := struct {
		Model struct {
			Provider string `json:"provider"`
			ModelID  string `json:"modelID"`
		} `json:"model"`
		DisableThreadData bool `json:"disableThreadData"`
	}{
		Model: struct {
			Provider string `json:"provider"`
			ModelID  string `json:"modelID"`
		}{
			Provider: provider,
			ModelID:  modelID,
		},
		DisableThreadData: false,
	}
	data, _ := json.Marshal(v)
	return string(data)
}

// DisplayMenu returns a numbered menu string of all models.
func DisplayMenu(cfg *ModelsConfig) string {
	var sb strings.Builder
	sb.WriteString("可选模型列表：\n")
	idx := 1
	for _, p := range cfg.Providers {
		for _, m := range p.Models {
			sb.WriteString(fmt.Sprintf("  %d. %s: %s\n", idx, p.DisplayName, m.DisplayName))
			idx++
		}
	}
	return sb.String()
}

// ListModelFlags returns all valid --model flag values with descriptions.
func ListModelFlags(cfg *ModelsConfig) string {
	var sb strings.Builder
	sb.WriteString("可传入 --model 的模型取值（格式: provider:modelId）：\n")
	for _, p := range cfg.Providers {
		for _, m := range p.Models {
			flag := p.Provider + ":" + m.ModelID
			sb.WriteString(fmt.Sprintf("  %-30s %s: %s\n", flag, p.DisplayName, m.DisplayName))
		}
	}
	return sb.String()
}

// GetModelByIndex returns the provider and modelId at the given 0-based index.
// The ordering is the same flat order as DisplayMenu.
func GetModelByIndex(cfg *ModelsConfig, index int) (provider, modelID string, err error) {
	idx := 0
	for _, p := range cfg.Providers {
		for _, m := range p.Models {
			if idx == index {
				return p.Provider, m.ModelID, nil
			}
			idx++
		}
	}
	return "", "", fmt.Errorf("model index %d out of range (total %d models)", index, idx)
}
