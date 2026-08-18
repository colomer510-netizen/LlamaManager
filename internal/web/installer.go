package web

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type InstallRequest struct {
	Path string `json:"path"`
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

	zipPath := strings.Trim(req.Path, "\"")
	
	// Crear carpeta bin/ si no existe
	binDir := "./bin"
	os.MkdirAll(binDir, os.ModePerm)

	// Extraer
	err = unzip(zipPath, binDir)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func autoInstallBinaries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	// 1. Obtener última release de GitHub
	resp, err := http.Get("https://api.github.com/repos/ggml-org/llama.cpp/releases/latest")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error conectando a GitHub: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadUrl string `json:"browser_download_url"`
		} `json:"assets"`
	}
	json.NewDecoder(resp.Body).Decode(&release)

	// 2. Buscar el archivo win-cpu-x64
	var downloadUrl string
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, "bin-win-cpu-x64.zip") {
			downloadUrl = asset.BrowserDownloadUrl
			break
		}
	}

	if downloadUrl == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "No se encontró un binario CPU para Windows en la última versión."})
		return
	}

	// 3. Descargar el archivo
	tempZip := "llama-temp-download.zip"
	out, err := os.Create(tempZip)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error creando archivo temporal: " + err.Error()})
		return
	}

	dlResp, err := http.Get(downloadUrl)
	if err != nil {
		out.Close()
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error descargando: " + err.Error()})
		return
	}
	defer dlResp.Body.Close()

	io.Copy(out, dlResp.Body)
	out.Close()

	// 4. Extraer
	binDir := "./bin"
	os.MkdirAll(binDir, os.ModePerm)
	err = unzip(tempZip, binDir)
	
	// Limpiar
	os.Remove(tempZip)

	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Error extrayendo ZIP: " + err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "msg": "Se descargó e instaló la última versión desde GitHub automáticamente."})
}

// Función auxiliar para extraer ZIP
func unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el zip: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue // Evitar ZipSlip
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
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
	}
	return nil
}
