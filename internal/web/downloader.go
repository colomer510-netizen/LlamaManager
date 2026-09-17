package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"llamamanager/internal/config"
)

var (
	downloadStatus struct {
		sync.Mutex
		Active   bool    `json:"active"`
		Progress float64 `json:"progress"` // 0 a 100
		File     string  `json:"file"`
		Error    string  `json:"error"`
	}
)

type passThru struct {
	io.Reader
	total    int64 // Total # of bytes transferred
	length   int64 // Expected length
}

func (pt *passThru) Read(p []byte) (int, error) {
	n, err := pt.Reader.Read(p)
	pt.total += int64(n)
	
	if pt.length > 0 {
		progress := float64(pt.total) / float64(pt.length) * 100
		
		downloadStatus.Lock()
		downloadStatus.Progress = progress
		downloadStatus.Unlock()
	}

	return n, err
}

func startModelDownload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	downloadStatus.Lock()
	if downloadStatus.Active {
		downloadStatus.Unlock()
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Ya hay una descarga en progreso"})
		return
	}
	downloadStatus.Unlock()

	var req struct {
		URL string `json:"url"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.URL == "" || !strings.HasPrefix(req.URL, "http") {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "URL inválida"})
		return
	}

	conf := config.LoadSettings()
	destDir := conf.ModelPath
	if destDir == "" {
		destDir = "."
	}

	// Extraer el nombre del archivo de la URL
	parts := strings.Split(req.URL, "/")
	filename := parts[len(parts)-1]
	if !strings.HasSuffix(filename, ".gguf") {
		filename += ".gguf" // forzar extensión por si acaso
	}
	
	// Limpiar filename de query strings
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	destPath := filepath.Join(destDir, filename)

	downloadStatus.Lock()
	downloadStatus.Active = true
	downloadStatus.Progress = 0
	downloadStatus.File = filename
	downloadStatus.Error = ""
	downloadStatus.Unlock()

	go func() {
		defer func() {
			downloadStatus.Lock()
			downloadStatus.Active = false
			downloadStatus.Unlock()
		}()

		resp, err := http.Get(req.URL)
		if err != nil {
			downloadStatus.Lock()
			downloadStatus.Error = "Error conectando: " + err.Error()
			downloadStatus.Unlock()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			downloadStatus.Lock()
			downloadStatus.Error = fmt.Sprintf("Error HTTP: %d", resp.StatusCode)
			downloadStatus.Unlock()
			return
		}

		out, err := os.Create(destPath)
		if err != nil {
			downloadStatus.Lock()
			downloadStatus.Error = "Error creando archivo: " + err.Error()
			downloadStatus.Unlock()
			return
		}
		defer out.Close()

		pt := &passThru{Reader: resp.Body, length: resp.ContentLength}
		_, err = io.Copy(out, pt)
		
		if err != nil {
			downloadStatus.Lock()
			downloadStatus.Error = "Error durante la descarga: " + err.Error()
			downloadStatus.Unlock()
			os.Remove(destPath) // Limpiar archivo parcial
		}
		
		// Forzar 100% al terminar exitosamente
		if err == nil {
			downloadStatus.Lock()
			downloadStatus.Progress = 100
			downloadStatus.Unlock()
			time.Sleep(2 * time.Second) // Dejar que el UI vea el 100%
		}
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "msg": "Descarga iniciada"})
}

func getDownloadProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	downloadStatus.Lock()
	defer downloadStatus.Unlock()
	
	json.NewEncoder(w).Encode(downloadStatus)
}
