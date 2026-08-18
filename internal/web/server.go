package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"llamamanager/internal/config"
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
	http.HandleFunc("/api/shutdown", shutdownServer)
	http.HandleFunc("/api/settings", handleSettings)
	
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

func handleSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		json.NewEncoder(w).Encode(config.LoadSettings())
		return
	} else if r.Method == http.MethodPost {
		var s config.Settings
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
		if err := config.SaveSettings(s); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
}

func getModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	dir := r.URL.Query().Get("dir")
	if dir == "" {
		conf := config.LoadSettings()
		dir = conf.ModelPath
		if dir == "" {
			dir = "."
		}
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
	Port  string `json:"port"`
}

// Helper para lanzar el bat intentando usar una pestaña de Windows Terminal
func launchBat(batPath string) error {
	// Obtener el directorio actual para indicarle a la nueva pestaña dónde buscar
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	// Intentar usar Windows Terminal para abrir en una nueva pestaña
	// -d define el directorio de inicio de esa pestaña
	cmd := exec.Command("wt", "-w", "0", "new-tab", "-d", cwd, "cmd", "/c", batPath)
	err = cmd.Start()
	if err == nil {
		return nil // Éxito abriendo la pestaña
	}
	
	// Fallback: Si no hay Windows Terminal, usamos el start tradicional en ventana separada
	cmd = exec.Command("cmd", "/c", "start", "/min", batPath)
	return cmd.Start()
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

	conf := config.LoadSettings()
	ctxSize := conf.ContextSize
	if ctxSize == 0 {
		ctxSize = 32768
	}

	// Escribir script bat para evitar problemas de comillas en Windows
	batPath := "run_chat.bat"
	batContent := fmt.Sprintf("@echo off\ntitle LlamaManager Chat\n\"%%~dp0%s\" -m \"%s\" -c %d -t %d -ngl 0 -cnv\npause\n", exePath, req.Model, ctxSize, threads)
	os.WriteFile(batPath, []byte(batContent), 0755)

	err := launchBat(batPath)
	
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

	conf := config.LoadSettings()
	port := req.Port
	if port == "" {
		port = conf.APIPort
		if port == "" {
			port = "8080"
		}
	}
	
	ctxSize := conf.ContextSize
	if ctxSize == 0 {
		ctxSize = 32768
	}

	batPath := "run_server.bat"
	batContent := fmt.Sprintf("@echo off\ntitle LlamaManager Server\n\"%%~dp0%s\" -m \"%s\" -c %d -t %d -ngl 0 --port %s\npause\n", exePath, req.Model, ctxSize, threads, port)
	os.WriteFile(batPath, []byte(batContent), 0755)

	err := launchBat(batPath)
	
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

func shutdownServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "msg": "Apagando el gestor..."})
	
	// Salir del programa limpiamente
	go func() {
		os.Exit(0)
	}()
}
