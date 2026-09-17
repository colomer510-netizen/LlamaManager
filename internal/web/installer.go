package web

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type InstallRequest struct {
	Path string `json:"path"`
}

type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}

type GitHubRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []ReleaseAsset `json:"assets"`
}

func installLocalBinaries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	var req InstallRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Path == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Ruta inválida"})
		return
	}

	archivePath := strings.Trim(req.Path, "\"")
	
	// Crear carpeta bin/ si no existe
	binDir := "./bin"
	os.MkdirAll(binDir, 0755)

	// Extraer (soporta .tar.gz y .zip)
	err = extractArchive(archivePath, binDir)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	if runtime.GOOS != "windows" {
		fixSharedLibraries(binDir)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func autoInstallBinaries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// 1. Obtener lista de releases de GitHub con User-Agent
	req, err := http.NewRequest("GET", "https://api.github.com/repos/ggml-org/llama.cpp/releases?per_page=5", nil)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error creando petición: " + err.Error()})
		return
	}
	req.Header.Set("User-Agent", "LlamaManager-App")

	resp, err := client.Do(req)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error conectando a GitHub: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil || len(releases) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "No se pudieron leer los releases de GitHub."})
		return
	}

	// 2. Buscar el archivo adecuado según el SO
	var downloadUrl string
	var downloadedFileName string

	for _, release := range releases {
		for _, asset := range release.Assets {
			name := strings.ToLower(asset.Name)
			if runtime.GOOS == "linux" {
				// En Linux buscar ubuntu-x64 (tar.gz o zip)
				if strings.Contains(name, "ubuntu-x64") && (strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip")) {
					downloadUrl = asset.BrowserDownloadUrl
					downloadedFileName = asset.Name
					break
				}
			} else if runtime.GOOS == "windows" {
				// En Windows buscar win-cpu-x64
				if strings.Contains(name, "win-cpu-x64") && strings.HasSuffix(name, ".zip") {
					downloadUrl = asset.BrowserDownloadUrl
					downloadedFileName = asset.Name
					break
				}
			} else if runtime.GOOS == "darwin" {
				// En macOS
				if strings.Contains(name, "macos-arm64") {
					downloadUrl = asset.BrowserDownloadUrl
					downloadedFileName = asset.Name
					break
				}
			}
		}
		if downloadUrl != "" {
			break
		}
	}

	if downloadUrl == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "No se encontró un binario compatible para tu sistema en las últimas versiones de GitHub."})
		return
	}

	// 3. Descargar el archivo
	tempFile := "llama-temp-download"
	if strings.HasSuffix(downloadedFileName, ".tar.gz") {
		tempFile += ".tar.gz"
	} else {
		tempFile += ".zip"
	}

	out, err := os.Create(tempFile)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error creando archivo temporal: " + err.Error()})
		return
	}

	dlClient := &http.Client{Timeout: 15 * time.Minute}
	dlReq, _ := http.NewRequest("GET", downloadUrl, nil)
	dlReq.Header.Set("User-Agent", "LlamaManager-App")
	dlResp, err := dlClient.Do(dlReq)
	if err != nil {
		out.Close()
		os.Remove(tempFile)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error descargando: " + err.Error()})
		return
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != http.StatusOK {
		out.Close()
		os.Remove(tempFile)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": fmt.Sprintf("Error HTTP al descargar: %s", dlResp.Status)})
		return
	}

	_, err = io.Copy(out, dlResp.Body)
	out.Close()
	if err != nil {
		os.Remove(tempFile)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error guardando descarga: " + err.Error()})
		return
	}

	// 4. Extraer en bin/
	binDir := "./bin"
	os.MkdirAll(binDir, 0755)
	err = extractArchive(tempFile, binDir)
	os.Remove(tempFile)

	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error extrayendo archivo: " + err.Error()})
		return
	}

	if runtime.GOOS != "windows" {
		fixSharedLibraries(binDir)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true, 
		"msg": fmt.Sprintf("Se descargó e instaló exitosamente %s en bin/", downloadedFileName),
	})
}

// extractArchive detecta el formato y descomprime (.tar.gz o .zip)
func extractArchive(src, dest string) error {
	if strings.HasSuffix(strings.ToLower(src), ".tar.gz") || strings.HasSuffix(strings.ToLower(src), ".tgz") {
		return untarGz(src, dest)
	}
	if strings.HasSuffix(strings.ToLower(src), ".zip") {
		return unzip(src, dest)
	}
	// Intentar primero con tar.gz, si falla intentar con zip
	if err := untarGz(src, dest); err == nil {
		return nil
	}
	return unzip(src, dest)
}

func untarGz(src string, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("no se pudo abrir tar.gz: %w", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("formato gzip inválido: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) && filepath.Clean(target) != filepath.Clean(dest) {
			continue // Evitar Zip/Tar Slip
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()

			if runtime.GOOS != "windows" {
				os.Chmod(target, 0755)
			}
		}
	}
	return nil
}

func unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el zip: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) && filepath.Clean(fpath) != filepath.Clean(dest) {
			continue // Evitar ZipSlip
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}

		// En Linux/Unix, dar permisos de ejecución a los binarios
		if runtime.GOOS != "windows" {
			os.Chmod(fpath, 0755)
		}
	}
	return nil
}

func fixSharedLibraries(destDir string) {
	// 1. Mover archivos de subdirectorios a destDir si los hay
	if entries, err := os.ReadDir(destDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				subPath := filepath.Join(destDir, entry.Name())
				if subEntries, err := os.ReadDir(subPath); err == nil {
					for _, subEntry := range subEntries {
						srcFile := filepath.Join(subPath, subEntry.Name())
						dstFile := filepath.Join(destDir, subEntry.Name())
						if _, err := os.Stat(dstFile); os.IsNotExist(err) {
							os.Rename(srcFile, dstFile)
						}
					}
				}
			}
		}
	}

	// 2. Crear enlaces simbólicos necesarios para librerías compartidas .so
	if entries, err := os.ReadDir(destDir); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if strings.Contains(name, ".so.") {
				// Por ejemplo libllama-common.so.0.2.0 -> libllama-common.so.0 y libllama-common.so
				parts := strings.Split(name, ".so.")
				if len(parts) == 2 {
					baseName := parts[0] + ".so"
					verParts := strings.Split(parts[1], ".")
					if len(verParts) > 0 {
						majorName := baseName + "." + verParts[0]
						majorPath := filepath.Join(destDir, majorName)
						if _, err := os.Lstat(majorPath); os.IsNotExist(err) {
							os.Symlink(name, majorPath)
						}
					}
					basePath := filepath.Join(destDir, baseName)
					if _, err := os.Lstat(basePath); os.IsNotExist(err) {
						os.Symlink(name, basePath)
					}
				}
			}
		}
	}
}
