package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	lastHeartbeat   time.Time
	heartbeatMu     sync.Mutex
	heartbeatActive bool
)

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

func monitorHeartbeat() {
	for {
		time.Sleep(5 * time.Second)
		heartbeatMu.Lock()
		if time.Since(lastHeartbeat) > 15*time.Second {
			fmt.Println("No se recibió heartbeat en 15 segundos. Cerrando LlamaManager...")
			os.Exit(0)
		}
		heartbeatMu.Unlock()
	}
}
