package config

import (
	"encoding/json"
	"os"
	"sync"
)

type Settings struct {
	ModelPath   string `json:"model_path"`
	APIPort     string `json:"api_port"`
	ContextSize int    `json:"context_size"`
	GPULayers   int    `json:"gpu_layers"`
	ForceCPU    bool   `json:"force_cpu"`
}

var (
	settingsFile = "llama_settings.json"
	mu           sync.Mutex
)

// DefaultSettings returns the fallback settings
func DefaultSettings() Settings {
	return Settings{
		ModelPath:   "",
		APIPort:     "8080",
		ContextSize: 32768,
		GPULayers:   0,
		ForceCPU:    false,
	}
}

// LoadSettings reads the settings from disk or returns defaults
func LoadSettings() Settings {
	mu.Lock()
	defer mu.Unlock()

	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return DefaultSettings()
	}

	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings()
	}

	// Apply defaults for empty required fields
	if s.APIPort == "" {
		s.APIPort = "8080"
	}
	if s.ContextSize == 0 {
		s.ContextSize = 32768
	}

	return s
}

// SaveSettings writes the settings to disk
func SaveSettings(s Settings) error {
	mu.Lock()
	defer mu.Unlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsFile, data, 0644)
}
