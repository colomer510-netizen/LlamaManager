package main

import (
	"os"
	"os/exec"
	"runtime"

	"llamamanager/internal/web"
)

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	clearScreen()
	web.BroadcastLog("=====================================================")
	web.BroadcastLog("  _      _                       __  __ ")
	web.BroadcastLog(" | |    | |                     |  \\/  |")
	web.BroadcastLog(" | |    | | __ _ _ __ ___   __ _| \\  / | __ _ _ __   __ _  __ _  ___ _ __ ")
	web.BroadcastLog(" | |    | |/ _` | '_ ` _ \\ / _` | |\\/| |/ _` | '_ \\ / _` |/ _` |/ _ \\ '__|")
	web.BroadcastLog(" | |____| | (_| | | | | | | (_| | |  | | (_| | | | | (_| | (_| |  __/ |   ")
	web.BroadcastLog(" |______|_|\\__,_|_| |_| |_|\\__,_|_|  |_|\\__,_|_| |_|\\__,_|\\__, |\\___|_|   ")
	web.BroadcastLog("                                                           __/ |          ")
	web.BroadcastLog("                                                          |___/           ")
	web.BroadcastLog("=====================================================")
	web.BroadcastLog("             Llama.cpp Universal Manager             ")
	web.BroadcastLog("=====================================================")

	// 3. Iniciar Backend Web
	web.StartWebServer()
}
