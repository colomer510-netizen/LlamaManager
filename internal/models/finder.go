package models

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ExpandUserPath limpia, expande tildes (~), variables de entorno y resuelve rutas absolutas.
func ExpandUserPath(path string) string {
	cleanPath := strings.TrimSpace(path)
	cleanPath = strings.Trim(cleanPath, "\"'`")

	if cleanPath == "" {
		return ""
	}

	// Expandir variables de entorno (ej: $HOME/Modelos)
	cleanPath = os.ExpandEnv(cleanPath)

	// Expandir tilde ~
	if cleanPath == "~" || strings.HasPrefix(cleanPath, "~/") || strings.HasPrefix(cleanPath, "~\\") {
		home, err := os.UserHomeDir()
		if err == nil {
			if cleanPath == "~" {
				cleanPath = home
			} else {
				cleanPath = filepath.Join(home, cleanPath[2:])
			}
		}
	}

	// Convertir a ruta absoluta
	absPath, err := filepath.Abs(cleanPath)
	if err == nil {
		return filepath.Clean(absPath)
	}

	return filepath.Clean(cleanPath)
}

// ValidateModelDirectory verifica si una ruta existe y es un directorio accesible.
func ValidateModelDirectory(dirPath string) (string, error) {
	expanded := ExpandUserPath(dirPath)
	if expanded == "" {
		return "", fmt.Errorf("la ruta no puede estar vacía")
	}

	_, err := os.Stat(expanded)
	if err != nil {
		if os.IsNotExist(err) {
			return expanded, fmt.Errorf("la carpeta o archivo no existe: %s", expanded)
		}
		return expanded, fmt.Errorf("error accediendo a la ruta: %w", err)
	}

	return expanded, nil
}

// FindGGUFModels escanea el directorio especificado en busca de archivos .gguf
// Soporta rutas con ~, relativas, absolutas y búsqueda recursiva en subcarpetas.
func FindGGUFModels(dirPath string) ([]string, error) {
	if dirPath == "" {
		dirPath = "."
	}

	expandedPath, err := ValidateModelDirectory(dirPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(expandedPath)
	if err != nil {
		return nil, err
	}

	// Si el usuario especificó directamente un archivo .gguf
	if !info.IsDir() {
		if strings.HasSuffix(strings.ToLower(info.Name()), ".gguf") {
			return []string{expandedPath}, nil
		}
		return nil, fmt.Errorf("el archivo seleccionado no tiene extensión .gguf: %s", expandedPath)
	}

	var foundModels []string
	baseDepth := strings.Count(filepath.Clean(expandedPath), string(os.PathSeparator))
	maxDepth := 4 // Límite de profundidad de subcarpetas para evitar congelamientos

	err = filepath.WalkDir(expandedPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // Ignorar carpetas sin permisos y continuar
		}

		// Ignorar carpetas ocultas y dependencias
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "bin" {
				if path != expandedPath {
					return filepath.SkipDir
				}
			}
			currentDepth := strings.Count(filepath.Clean(path), string(os.PathSeparator)) - baseDepth
			if currentDepth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		// Detectar archivos .gguf (case-insensitive)
		if strings.HasSuffix(strings.ToLower(d.Name()), ".gguf") {
			foundModels = append(foundModels, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error escaneando modelos: %w", err)
	}

	// Ordenar alfabéticamente
	sort.Strings(foundModels)

	return foundModels, nil
}
