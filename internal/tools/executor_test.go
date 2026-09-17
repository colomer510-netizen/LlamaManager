package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveBinPath(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	subBinDir := filepath.Join(binDir, "sub-release")
	os.MkdirAll(subBinDir, 0755)
	
	// Crear un binario nativo de linux y uno en subcarpeta
	nativeLinuxTool := filepath.Join(binDir, "llama-server")
	subFolderTool := filepath.Join(subBinDir, "llama-cli")
	
	os.WriteFile(nativeLinuxTool, []byte("dummy binary"), 0755)
	os.WriteFile(subFolderTool, []byte("dummy binary 2"), 0755)

	originalCWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalCWD)

	// 1. Buscar binario directo
	resolved, err := ResolveBinPath("llama-server")
	if err != nil {
		t.Errorf("Error buscando llama-server: %v", err)
	}
	if !strings.HasSuffix(resolved, "llama-server") {
		t.Errorf("Ruta incorrecta: %s", resolved)
	}

	// 2. Buscar binario pasando .exe cuando en disco es nativo sin .exe
	resolvedExe, err := ResolveBinPath("llama-server.exe")
	if err != nil {
		t.Errorf("Error buscando llama-server.exe (variación): %v", err)
	}
	if !strings.HasSuffix(resolvedExe, "llama-server") {
		t.Errorf("Ruta incorrecta: %s", resolvedExe)
	}

	// 3. Buscar binario en subcarpeta (sub-release)
	resolvedSub, err := ResolveBinPath("llama-cli")
	if err != nil {
		t.Errorf("Error buscando llama-cli en subdirectorio: %v", err)
	}
	if !strings.HasSuffix(resolvedSub, filepath.Join("sub-release", "llama-cli")) {
		t.Errorf("Ruta de subdirectorio incorrecta: %s", resolvedSub)
	}
}
