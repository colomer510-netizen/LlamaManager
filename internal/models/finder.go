package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindGGUFModels escanea el directorio especificado en busca de archivos .gguf
func FindGGUFModels(dirPath string) ([]string, error) {
	var models []string

	// Revisar si el directorio existe
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("el directorio no existe: %s", dirPath)
	}

	// Leer archivos en la raíz del directorio
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".gguf") {
			fullPath := filepath.Join(dirPath, entry.Name())
			models = append(models, fullPath)
		}
	}

	return models, nil
}
