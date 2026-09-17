package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigSaveAndLoad(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "llama_config_test_*")
	if err != nil {
		t.Fatalf("error creando tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldSettingsFile := settingsFile
	settingsFile = filepath.Join(tempDir, "test_settings.json")
	defer func() { settingsFile = oldSettingsFile }()

	// 1. Probar carga de valores por defecto
	def := LoadSettings()
	if def.APIPort != "8080" || def.ContextSize != 32768 {
		t.Errorf("Valores por defecto incorrectos: %+v", def)
	}

	// 2. Guardar nueva configuración
	custom := Settings{
		ModelPath:   "/home/user/ai_models",
		APIPort:     "9090",
		ContextSize: 16384,
		GPULayers:   33,
		ForceCPU:    true,
	}

	err = SaveSettings(custom)
	if err != nil {
		t.Fatalf("Error guardando settings: %v", err)
	}

	// 3. Leer la configuración guardada
	loaded := LoadSettings()
	if loaded.ModelPath != custom.ModelPath || loaded.APIPort != "9090" || loaded.GPULayers != 33 {
		t.Errorf("Configuración cargada no coincide: %+v", loaded)
	}
}
