package main

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"llamamanager/internal/config"
	"llamamanager/internal/hardware"
	"llamamanager/internal/models"
	"llamamanager/internal/tools"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	serverCmd *exec.Cmd
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) domReady(ctx context.Context) {}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

func (a *App) shutdown(ctx context.Context) {}

// API Methods exposed to frontend

func (a *App) GetModels(modelPath string) []string {
	if modelPath == "" {
		conf := config.LoadSettings()
		modelPath = conf.ModelPath
	}
	found, _ := models.FindGGUFModels(modelPath)
	return found
}

func (a *App) GetHardware() map[string]interface{} {
	specs, err := hardware.GetSystemSpecs()
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	return map[string]interface{}{
		"cpu":          specs.CpuModel,
		"logicalCores": specs.LogicalCores,
		"ramGB":        specs.TotalRAMGB,
	}
}

func (a *App) LoadSettings() config.Settings {
	return config.LoadSettings()
}

func (a *App) SaveSettings(settings config.Settings) map[string]interface{} {
	err := config.SaveSettings(settings)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true}
}


func (a *App) StopLlama() {
	if a.serverCmd != nil && a.serverCmd.Process != nil {
		a.serverCmd.Process.Kill()
		a.serverCmd.Process.Wait()
		a.serverCmd = nil
		wailsRuntime.EventsEmit(a.ctx, "server-log", "🛑 Servidor Llama detenido.")
	}
}

func (a *App) StartLlama(model string, port string) map[string]interface{} {
	if model == "" {
		return map[string]interface{}{"success": false, "error": "El modelo está vacío"}
	}
	
	a.StopLlama() // Detener si ya hay uno corriendo
	
	exePath, err := tools.ResolveBinPath("llama-server.exe")
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}

	conf := config.LoadSettings()
	ctxSize := conf.ContextSize
	if ctxSize == 0 {
		ctxSize = 32768
	}
	if port == "" {
		port = conf.APIPort
		if port == "" {
			port = "8080"
		}
	}

	specs, _ := hardware.GetSystemSpecs()
	threads := 4
	if specs != nil && specs.LogicalCores > 0 {
		threads = specs.LogicalCores - 1
	}

	cmd := exec.Command(exePath, 
		"-m", model, 
		"-c", fmt.Sprintf("%d", ctxSize), 
		"-t", fmt.Sprintf("%d", threads), 
		"-ngl", fmt.Sprintf("%d", conf.GPULayers), 
		"--port", port)
		
	// Hacerlo invisible en Windows
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return map[string]interface{}{"success": false, "error": "Error al iniciar llama-server: " + err.Error()}
	}

	a.serverCmd = cmd
	
	go func() {
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			wailsRuntime.EventsEmit(a.ctx, "server-log", scanner.Text())
		}
	}()

	wailsRuntime.EventsEmit(a.ctx, "server-log", "🚀 Servidor Llama iniciado exitosamente (Modo Oculto).")
	return map[string]interface{}{"success": true}
}

func getBinFolder() string {
	if _, err := os.Stat(filepath.Join("..", "..", "bin")); err == nil {
		return filepath.Join("..", "..", "bin") // Modo Wails dev
	}
	return "bin" // Modo producción (compilado)
}

func (a *App) InstallLocalBinaries(zipPath string) map[string]interface{} {
	a.StopLlama()
	exec.Command("taskkill", "/F", "/IM", "llama-server.exe").Run() // Asegurar cierre de huerfanos
	
	zipPath = strings.Trim(zipPath, "\"")
	binDir := getBinFolder()
	os.MkdirAll(binDir, os.ModePerm)

	err := unzip(zipPath, binDir)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true}
}

func (a *App) AutoInstallBinaries() map[string]interface{} {
	a.StopLlama()
	exec.Command("taskkill", "/F", "/IM", "llama-server.exe").Run() // Asegurar cierre de huerfanos

	resp, err := http.Get("https://api.github.com/repos/ggml-org/llama.cpp/releases/latest")
	if err != nil {
		return map[string]interface{}{"success": false, "error": "Error conectando a GitHub: " + err.Error()}
	}
	defer resp.Body.Close()

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadUrl string `json:"browser_download_url"`
		} `json:"assets"`
	}
	json.NewDecoder(resp.Body).Decode(&release)

	var downloadUrl string
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, "bin-win-cpu-x64.zip") {
			downloadUrl = asset.BrowserDownloadUrl
			break
		}
	}

	if downloadUrl == "" {
		return map[string]interface{}{"success": false, "error": "No se encontró un binario CPU para Windows en la última versión."}
	}

	tempZip := "llama-temp-download.zip"
	out, err := os.Create(tempZip)
	if err != nil {
		return map[string]interface{}{"success": false, "error": "Error creando archivo temporal: " + err.Error()}
	}

	dlResp, err := http.Get(downloadUrl)
	if err != nil {
		out.Close()
		return map[string]interface{}{"success": false, "error": "Error descargando: " + err.Error()}
	}
	defer dlResp.Body.Close()

	io.Copy(out, dlResp.Body)
	out.Close()

	binDir := getBinFolder()
	os.MkdirAll(binDir, os.ModePerm)
	err = unzip(tempZip, binDir)
	os.Remove(tempZip)

	if err != nil {
		return map[string]interface{}{"success": false, "error": "Error extrayendo ZIP: " + err.Error()}
	}

	return map[string]interface{}{"success": true, "msg": "Se descargó e instaló la última versión desde GitHub automáticamente."}
}

func unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el zip: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue 
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

func (a *App) CheckUpdateStatus() map[string]interface{} {
	resp, err := http.Get("https://api.github.com/repos/ggml-org/llama.cpp/releases/latest")
	if err != nil {
		return map[string]interface{}{"success": false, "error": "Error conectando a GitHub"}
	}
	defer resp.Body.Close()

	var release struct {
		TagName   string `json:"tag_name"`
		Published string `json:"published_at"`
	}
	json.NewDecoder(resp.Body).Decode(&release)

	localDate := "No instalado"
	binDir := getBinFolder()
	info, err := os.Stat(filepath.Join(binDir, "llama-server.exe"))
	if err == nil {
		localDate = info.ModTime().Format("2006-01-02 15:04:05")
	}

	return map[string]interface{}{
		"success": true, 
		"latestVersion": release.TagName,
		"latestDate": release.Published,
		"localDate": localDate,
	}
}

func (a *App) RunMultimodalCLI() map[string]interface{} {
	exePath, err := tools.ResolveBinPath("llama-mtmd-cli.exe")
	if err != nil {
		exePath, err = tools.ResolveBinPath("llama-llava-cli.exe")
		if err != nil {
			return map[string]interface{}{"success": false, "error": "No se encontró llama-mtmd-cli ni llama-llava-cli."}
		}
	}
	
	err = tools.RunInteractive(exePath, []string{"--help"})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true}
}

func (a *App) RunQuantize() map[string]interface{} {
	inputFile, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Selecciona el modelo GGUF original",
		Filters: []wailsRuntime.FileFilter{{DisplayName: "GGUF", Pattern: "*.gguf"}},
	})
	if err != nil || inputFile == "" {
		return map[string]interface{}{"success": false, "error": "Cancelado"}
	}

	outputFile, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title: "Guardar modelo cuantizado",
		DefaultFilename: "modelo-Q4_K_M.gguf",
		Filters: []wailsRuntime.FileFilter{{DisplayName: "GGUF", Pattern: "*.gguf"}},
	})
	if err != nil || outputFile == "" {
		return map[string]interface{}{"success": false, "error": "Cancelado"}
	}

	exePath, err := tools.ResolveBinPath("llama-quantize.exe")
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}

	err = tools.RunInteractive(exePath, []string{inputFile, outputFile, "Q4_K_M"})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true}
}

func (a *App) RunRPC() map[string]interface{} {
	exePath, err := tools.ResolveBinPath("ggml-rpc-server.exe")
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	err = tools.RunInteractive(exePath, []string{"--host", "0.0.0.0", "--port", "50052"})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true}
}

func (a *App) RunBench() map[string]interface{} {
	exePath, err := tools.ResolveBinPath("llama-bench.exe")
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	err = tools.RunInteractive(exePath, []string{"-p", "512", "-n", "128"})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true}
}

