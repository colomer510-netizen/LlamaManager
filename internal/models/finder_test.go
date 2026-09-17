package models

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandUserPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("no se pudo obtener home: %v", err)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"~", home},
		{"~/Modelos", filepath.Join(home, "Modelos")},
		{"  ~/Modelos  ", filepath.Join(home, "Modelos")},
		{"\"~/Modelos\"", filepath.Join(home, "Modelos")},
		{"'~/Modelos'", filepath.Join(home, "Modelos")},
		{"", ""},
	}

	for _, tt := range tests {
		got := ExpandUserPath(tt.input)
		if got != tt.expected {
			t.Errorf("ExpandUserPath(%q) = %q; se esperaba %q", tt.input, got, tt.expected)
		}
	}
}

func TestFindGGUFModels(t *testing.T) {
	// Crear estructura temporal de prueba
	tempDir, err := os.MkdirTemp("", "llama_models_test_*")
	if err != nil {
		t.Fatalf("error creando tempDir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	subDir := filepath.Join(tempDir, "subfolder")
	os.MkdirAll(subDir, 0755)

	// Crear archivos .gguf y no .gguf
	f1 := filepath.Join(tempDir, "model1.gguf")
	f2 := filepath.Join(subDir, "model2.GGUF")
	f3 := filepath.Join(tempDir, "notes.txt")

	os.WriteFile(f1, []byte("fake gguf 1"), 0644)
	os.WriteFile(f2, []byte("fake gguf 2"), 0644)
	os.WriteFile(f3, []byte("some text"), 0644)

	// Probar escaneo del directorio raíz
	models, err := FindGGUFModels(tempDir)
	if err != nil {
		t.Fatalf("FindGGUFModels falló: %v", err)
	}

	if len(models) != 2 {
		t.Errorf("Se esperaban 2 modelos, se encontraron %d: %v", len(models), models)
	}

	// Probar pasando archivo directo
	singleModel, err := FindGGUFModels(f1)
	if err != nil {
		t.Fatalf("FindGGUFModels con archivo individual falló: %v", err)
	}
	if len(singleModel) != 1 || singleModel[0] != f1 {
		t.Errorf("Resultado inesperado para archivo único: %v", singleModel)
	}

	// Probar directorio inexistente
	_, err = FindGGUFModels(filepath.Join(tempDir, "no_existe"))
	if err == nil {
		t.Errorf("Se esperaba error con directorio inexistente")
	}
}
