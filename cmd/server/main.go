package main

import (
	"fmt"
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
	fmt.Println("=====================================================")
	fmt.Println("  _      _                       __  __ ")
	fmt.Println(" | |    | |                     |  \\/  |")
	fmt.Println(" | |    | | __ _ _ __ ___   __ _| \\  / | __ _ _ __   __ _  __ _  ___ _ __ ")
	fmt.Println(" | |    | |/ _` | '_ ` _ \\ / _` | |\\/| |/ _` | '_ \\ / _` |/ _` |/ _ \\ '__|")
	fmt.Println(" | |____| | (_| | | | | | | (_| | |  | | (_| | | | | (_| | (_| |  __/ |   ")
	fmt.Println(" |______|_|\\__,_|_| |_| |_|\\__,_|_|  |_|\\__,_|_| |_|\\__,_|\\__, |\\___|_|   ")
	fmt.Println("                                                           __/ |          ")
	fmt.Println("                                                          |___/           ")
	fmt.Println("=====================================================")
	fmt.Println("             Llama.cpp Universal Manager             ")
	fmt.Println("=====================================================\n")

	// 3. Iniciar Backend Web
	web.StartWebServer()
}
