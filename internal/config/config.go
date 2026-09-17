package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Settings struct {
	ModelPath   string `json:"model_path"`
	APIPort     string `json:"api_port"`
	ContextSize int    `json:"context_size"`
	GPULayers   int    `json:"gpu_layers"`
	ForceCPU    bool   `json:"force_cpu"`
	RPCServers  string `json:"rpc_servers"`
}

var (
	settingsFile = "llama_settings.json"
	mu           sync.Mutex
)

func getSettingsPath() string {
	if filepath.IsAbs(settingsFile) {
		return settingsFile
	}
	// Intentar guardar junto al ejecutable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidate := filepath.Join(exeDir, settingsFile)
		// Si ya existe junto al ejecutable o la carpeta es escribible
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		if _, err := os.Stat(exeDir); err == nil {
			return candidate
		}
	}
	return settingsFile
}

// DefaultSettings returns the fallback settings
func DefaultSettings() Settings {
	return Settings{
		ModelPath:   "",
		APIPort:     "8080",
		ContextSize: 4096,
		GPULayers:   0,
		ForceCPU:    false,
	}
}

// LoadSettings reads the settings from disk or returns defaults
func LoadSettings() Settings {
	mu.Lock()
	defer mu.Unlock()

	targetPath := getSettingsPath()
	data, err := os.ReadFile(targetPath)
	if err != nil {
		// Fallback al directorio actual si es diferente
		data, err = os.ReadFile(settingsFile)
		if err != nil {
			return DefaultSettings()
		}
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
		s.ContextSize = 4096
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

	targetPath := getSettingsPath()
	return os.WriteFile(targetPath, data, 0644)
}
