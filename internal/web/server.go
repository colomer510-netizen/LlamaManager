package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"llamamanager/internal/hardware"
	"llamamanager/internal/models"
)

func StartWebServer() {
	// Servir archivos estáticos del frontend
	http.Handle("/", http.FileServer(http.Dir("./public")))

	// Endpoints API
	http.HandleFunc("/api/models", getModels)
	http.HandleFunc("/api/hardware", getHardware)
	http.HandleFunc("/api/install", installLocalBinaries)
	http.HandleFunc("/api/autoinstall", autoInstallBinaries)
	http.HandleFunc("/api/run/chat", runChat)
	http.HandleFunc("/api/run/server", runServer)
	
	fmt.Println("=====================================================")
	fmt.Println("🚀 Servidor Web de LlamaManager iniciado en el puerto 3000")
	fmt.Println("🌐 Abre tu navegador en: http://localhost:3000")
	fmt.Println("=====================================================")

	// Intentar abrir el navegador automáticamente
	openBrowser("http://localhost:3000")

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println("Error iniciando el servidor web:", err)
	}
}

func getModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	dir := r.URL.Query().Get("dir")
	if dir == "" {
		dir = "." // Por defecto busca en la carpeta actual
	}

	foundModels, err := models.FindGGUFModels(dir)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"models": foundModels})
}

// Estructura para recibir la petición de ejecución
type runRequest struct {
	Model string `json:"model"`
}

func runChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	var req runRequest
	json.NewDecoder(r.Body).Decode(&req)

	if req.Model == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "El modelo no puede estar vacío"})
		return
	}

	// Obtener hardware para hilos
	specs, _ := hardware.GetSystemSpecs()
	threads := 4
	if specs != nil && specs.LogicalCores > 0 {
		threads = specs.LogicalCores - 1
	}

	// Ejecutar en nueva ventana
	exePath := filepath.Join("bin", "llama-cli.exe")
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "No se encuentra bin/llama-cli.exe. ¡Por favor usa el botón de Instalar Binarios primero!"})
		return
	}

	// Escribir script bat para evitar problemas de comillas en Windows
	batPath := "run_chat.bat"
	batContent := fmt.Sprintf("@echo off\ntitle LlamaManager Chat\n\"%%~dp0%s\" -m \"%s\" -c 8192 -t %d -ngl 0 --color -i\npause\n", exePath, req.Model, threads)
	os.WriteFile(batPath, []byte(batContent), 0755)

	cmd := exec.Command("cmd", "/c", "start", batPath)
	err := cmd.Start()
	
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func runServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	var req runRequest
	json.NewDecoder(r.Body).Decode(&req)

	if req.Model == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "El modelo no puede estar vacío"})
		return
	}

	specs, _ := hardware.GetSystemSpecs()
	threads := 4
	if specs != nil && specs.LogicalCores > 0 {
		threads = specs.LogicalCores - 1
	}

	exePath := filepath.Join("bin", "llama-server.exe")
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "No se encuentra bin/llama-server.exe. ¡Por favor usa el botón de Instalar Binarios primero!"})
		return
	}

	batPath := "run_server.bat"
	batContent := fmt.Sprintf("@echo off\ntitle LlamaManager Server\n\"%%~dp0%s\" -m \"%s\" -c 8192 -t %d -ngl 0 --port 8080\npause\n", exePath, req.Model, threads)
	os.WriteFile(batPath, []byte(batContent), 0755)

	cmd := exec.Command("cmd", "/c", "start", batPath)
	err := cmd.Start()
	
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func getHardware(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	specs, err := hardware.GetSystemSpecs()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(specs)
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		fmt.Println("No se pudo abrir el navegador automáticamente.")
	}
}
