package web

import (
    "time"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"path/filepath"
	"sync"
	"strconv"

	"llamamanager/internal/models"
	"llamamanager/internal/config"
	"llamamanager/internal/hardware"
	"llamamanager/internal/tools"
)


var (
	activeLlamaCmd *exec.Cmd
	activeLlamaMu  sync.Mutex
	
	lastHeartbeat   time.Time
	heartbeatMu     sync.Mutex
	heartbeatActive bool
)

// handleHeartbeat actualiza el timestamp de la última vez que la web nos contactó
func handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	heartbeatMu.Lock()
	lastHeartbeat = time.Now()
	if !heartbeatActive {
		heartbeatActive = true
		go monitorHeartbeat()
	}
	heartbeatMu.Unlock()
	w.WriteHeader(http.StatusOK)
}

// monitorHeartbeat cierra el servidor si cierras la pestaña del navegador
func monitorHeartbeat() {
	for {
		time.Sleep(5 * time.Second)
		heartbeatMu.Lock()
		if time.Since(lastHeartbeat) > 120*time.Second {
			// Matar el modelo si estaba corriendo
			activeLlamaMu.Lock()
			if activeLlamaCmd != nil && activeLlamaCmd.Process != nil {
				activeLlamaCmd.Process.Kill()
			}
			activeLlamaMu.Unlock()
			
			// Cerrar el backend de Go
			os.Exit(0)
		}
		heartbeatMu.Unlock()
	}
}

func StartWebServer() {
	// Servir archivos estáticos del frontend
	http.Handle("/", http.FileServer(http.Dir("./public")))

	// Endpoints API
	http.HandleFunc("/api/models", getModels)
	http.HandleFunc("/api/hardware", getHardware)
	http.HandleFunc("/api/hardware/optimize", handleOptimize)
	http.HandleFunc("/api/install", installLocalBinaries)
	http.HandleFunc("/api/autoinstall", autoInstallBinaries)
		http.HandleFunc("/api/run/server", runServer)
	http.HandleFunc("/api/run/stop", stopServer)
	http.HandleFunc("/api/agent/chat", handleAgentChat)
	http.HandleFunc("/api/agent/execute", handleAgentExecute)
	http.HandleFunc("/api/agent/reset", handleAgentReset)
	http.HandleFunc("/api/models/download", startModelDownload)
	http.HandleFunc("/api/models/progress", getDownloadProgress)
	http.HandleFunc("/api/shutdown", shutdownServer)
	http.HandleFunc("/api/settings", handleSettings)
	http.HandleFunc("/api/agent/run_terminal", runAgentTerminal)
	http.HandleFunc("/api/logs/stream", handleLogStream)
http.HandleFunc("/api/heartbeat", handleHeartbeat)
	
	BroadcastLog("=====================================================")
	BroadcastLog("🚀 Servidor Web de LlamaManager iniciado en el puerto 3000")
	BroadcastLog("🌐 Abre tu navegador en: http://localhost:3000")
	BroadcastLog("=====================================================")

	// Intentar abrir el navegador automáticamente
	openBrowser("http://localhost:3000")

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println("Error iniciando el servidor web:", err)
	}
}

func handleOptimize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	specs, _ := hardware.GetSystemSpecs()
	
	ctxSize := 4096
	gpuLayers := 0

	if specs != nil {
		if specs.TotalRAMGB >= 16 {
			ctxSize = 8192
		} else if specs.TotalRAMGB <= 8 {
			ctxSize = 2048
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"msg": "Hardware analizado y configurado localmente",
		"context_size": ctxSize,
		"gpu_layers": gpuLayers,
	})
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

// Helper para lanzar comandos directamente en la terminal sin escribir archivos .bat
func launchCommand(title, command string) error {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	// Comando completo que se ejecutará en cmd
	fullCommand := fmt.Sprintf("title %s && %s", title, command)

	// Intentar usar Windows Terminal
	cmd := exec.Command("wt", "-w", "0", "new-tab", "-d", cwd, "cmd", "/c", fullCommand)
	err = cmd.Start()
	if err == nil {
		return nil
	}
	
	// Fallback a ventana separada
	cmd = exec.Command("cmd", "/c", "start", title, "cmd", "/c", fullCommand)
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
	exePath, err := tools.ResolveBinPath("llama-cli")
	isUnified := false
	if err != nil {
		exePath, err = tools.ResolveBinPath("llama")
		isUnified = true
	}
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "No se encuentra bin/llama-cli ni llama. ¡Por favor usa el botón de Instalar Binarios primero!"})
		return
	}

	conf := config.LoadSettings()
	ctxSize := conf.ContextSize
	if ctxSize == 0 {
		ctxSize = 32768
	}

	// Ejecutar directamente el comando sin archivos .bat
	subcommand := ""
	if isUnified {
		// llama unificado por defecto actúa como cli si se le pasa -m, 
		// pero para chat interactivo se recomienda `llama run` o `llama cli` o simplemente pasarle -cnv.
		// En versiones recientes unificadas: `llama -m ...` sigue funcionando para CLI.
	}
	command := fmt.Sprintf("\"%s\" %s-m \"%s\" -c %d -t %d -ngl %d -cnv && pause", exePath, subcommand, req.Model, ctxSize, threads, conf.GPULayers)

	err = launchCommand("LlamaManager_Chat", command)
	
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

	// Intentar encontrar llama-server o llama (binario unificado)
	exePath, err := tools.ResolveBinPath("llama-server")
	isUnified := false
	if err != nil {
		exePath, err = tools.ResolveBinPath("llama")
		isUnified = true
	}
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "No se encuentra llama-server ni llama. ¡Instala los binarios!"})
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

	// Ejecutar directamente el comando sin archivos .bat
		activeLlamaMu.Lock()
	if activeLlamaCmd != nil && activeLlamaCmd.Process != nil {
		activeLlamaMu.Unlock()
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Ya hay un servidor en ejecución. Por favor, detenlo primero."})
		return
	}

	args := []string{}
	if isUnified {
		args = append(args, "serve")
	}
	args = append(args, "--host", "0.0.0.0", "-m", req.Model, "-c", strconv.Itoa(ctxSize), "-t", strconv.Itoa(threads), "-ngl", strconv.Itoa(conf.GPULayers), "--port", port)

	cmd := exec.Command(exePath, args...)
	cmd.Env = os.Environ()
	if conf.GPULayers == 0 {
		cmd.Env = append(cmd.Env, "GGML_VK_VISIBLE_DEVICES=", "CUDA_VISIBLE_DEVICES=")
	}
	if runtime.GOOS != "windows" {
		cwd, _ := os.Getwd()
		cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH="+filepath.Join(cwd, "bin"))
	}
	
	// Enviar logs del modelo a la consola web en tiempo real
	stdoutWriter, stderrWriter := NewLogPipe("llama")
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	err = cmd.Start()
	if err != nil {
		activeLlamaMu.Lock()
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}
	
	activeLlamaCmd = cmd
	activeLlamaMu.Unlock()

	go func() {
		cmd.Wait()
		activeLlamaMu.Lock()
		if activeLlamaCmd == cmd {
			activeLlamaCmd = nil
		}
		activeLlamaMu.Unlock()
	}()

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

func stopServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}
	activeLlamaMu.Lock()
	defer activeLlamaMu.Unlock()

	if activeLlamaCmd != nil && activeLlamaCmd.Process != nil {
		activeLlamaCmd.Process.Kill()
		activeLlamaCmd = nil
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "msg": "Servidor detenido correctamente."})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "No hay un servidor activo."})
}

func shutdownServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "msg": "Apagando el gestor..."})
	
	// Matar el modelo si estaba corriendo
	activeLlamaMu.Lock()
	if activeLlamaCmd != nil && activeLlamaCmd.Process != nil {
		activeLlamaCmd.Process.Kill()
	}
	activeLlamaMu.Unlock()

	// Salir del programa limpiamente
	go func() {
		os.Exit(0)
	}()
}

func runAgentTerminal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	// Buscar el ejecutable del agente
	agentPath := "./agent"
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		agentPath = filepath.Join("bin", "agent")
		if _, err := os.Stat(agentPath); os.IsNotExist(err) {
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "No se encontró el ejecutable del agente."})
			return
		}
	}

	absPath, _ := filepath.Abs(agentPath)
	err := tools.RunInteractive(absPath, []string{})
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}
