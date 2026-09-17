package web

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ============================================================
// Sistema de Log Streaming en Tiempo Real (SSE)
// ============================================================

const maxLogLines = 500

// logEntry representa una línea de log con timestamp
type logEntry struct {
	Time    time.Time
	Message string
}

// LogBuffer almacena logs en un buffer circular thread-safe
type LogBuffer struct {
	mu         sync.RWMutex
	entries    []logEntry
	subscribers map[chan logEntry]struct{}
}

var globalLogBuffer = &LogBuffer{
	entries:     make([]logEntry, 0, maxLogLines),
	subscribers: make(map[chan logEntry]struct{}),
}

// BroadcastLog envía un mensaje de log a todos los suscriptores SSE
// y lo almacena en el buffer circular.
func BroadcastLog(msg string) {
	entry := logEntry{
		Time:    time.Now(),
		Message: msg,
	}

	globalLogBuffer.mu.Lock()
	// Buffer circular: si estamos llenos, eliminamos el más viejo
	if len(globalLogBuffer.entries) >= maxLogLines {
		globalLogBuffer.entries = globalLogBuffer.entries[1:]
	}
	globalLogBuffer.entries = append(globalLogBuffer.entries, entry)

	// Enviar a todos los suscriptores
	for ch := range globalLogBuffer.subscribers {
		select {
		case ch <- entry:
		default:
			// Si el canal está lleno, no bloqueamos
		}
	}
	globalLogBuffer.mu.Unlock()

	// También imprimir en la consola del servidor (para debug)
	fmt.Println(msg)
}

// subscribe crea un canal para recibir logs en tiempo real
func (lb *LogBuffer) subscribe() chan logEntry {
	ch := make(chan logEntry, 50)
	lb.mu.Lock()
	lb.subscribers[ch] = struct{}{}
	lb.mu.Unlock()
	return ch
}

// unsubscribe elimina un suscriptor
func (lb *LogBuffer) unsubscribe(ch chan logEntry) {
	lb.mu.Lock()
	delete(lb.subscribers, ch)
	lb.mu.Unlock()
	close(ch)
}

// getHistory devuelve todas las entradas del buffer
func (lb *LogBuffer) getHistory() []logEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	result := make([]logEntry, len(lb.entries))
	copy(result, lb.entries)
	return result
}

// LogWriter implementa io.Writer para capturar stdout/stderr
// y enviar cada línea al sistema de logs.
type LogWriter struct {
	prefix string
}

// NewLogWriter crea un nuevo LogWriter con un prefijo opcional
func NewLogWriter(prefix string) *LogWriter {
	return &LogWriter{prefix: prefix}
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	text := string(p)
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if lw.prefix != "" {
			BroadcastLog(fmt.Sprintf("[%s] %s", lw.prefix, line))
		} else {
			BroadcastLog(line)
		}
	}
	return len(p), nil
}

// NewLogPipe crea un par de io.Writer que captura líneas y las envía al log.
// Retorna un io.Writer que se puede asignar a cmd.Stdout o cmd.Stderr.
func NewLogPipe(prefix string) (io.Writer, io.Writer) {
	stdoutWriter := NewLogWriter(prefix)
	stderrWriter := NewLogWriter(prefix)
	return stdoutWriter, stderrWriter
}

// NewLogPipeReader crea un pipe y lee líneas en una goroutine.
// Útil para procesos que necesitan un pipe real (no solo un Writer).
func NewLogPipeReader(prefix string) (*io.PipeReader, *io.PipeWriter) {
	pr, pw := io.Pipe()
	go func() {
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) != "" {
				if prefix != "" {
					BroadcastLog(fmt.Sprintf("[%s] %s", prefix, line))
				} else {
					BroadcastLog(line)
				}
			}
		}
	}()
	return pr, pw
}

// handleLogStream es el handler HTTP para Server-Sent Events (SSE).
// El navegador se conecta a /api/logs/stream y recibe logs en tiempo real.
func handleLogStream(w http.ResponseWriter, r *http.Request) {
	// Configurar headers SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Enviar historial existente
	history := globalLogBuffer.getHistory()
	for _, entry := range history {
		fmt.Fprintf(w, "data: %s\n\n", entry.Message)
	}
	flusher.Flush()

	// Suscribirse a nuevos logs
	ch := globalLogBuffer.subscribe()
	defer globalLogBuffer.unsubscribe(ch)

	// Detectar cuando el cliente se desconecta
	ctx := r.Context()

	for {
		select {
		case entry := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", entry.Message)
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}
