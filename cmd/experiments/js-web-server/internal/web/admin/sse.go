package admin

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/repository"
	"github.com/rs/zerolog/log"
)

// SSEHandler handles Server-Sent Events for real-time updates
type SSEHandler struct {
	logger *engine.RequestLogger
	repos  repository.RepositoryManager

	// SSE support
	sseClients map[string]chan string
	sseMutex   sync.RWMutex
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(logger *engine.RequestLogger, repos repository.RepositoryManager) *SSEHandler {
	sh := &SSEHandler{
		logger:     logger,
		repos:      repos,
		sseClients: make(map[string]chan string),
	}

	// Start monitoring for new requests
	go sh.monitorNewRequests()

	return sh
}

// ServeSSE handles Server-Sent Events for real-time updates
func (sh *SSEHandler) ServeSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create unique client ID
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())
	clientChan := make(chan string, 10)

	// Register client
	sh.sseMutex.Lock()
	sh.sseClients[clientID] = clientChan
	sh.sseMutex.Unlock()

	// Clean up on disconnect
	defer func() {
		sh.sseMutex.Lock()
		delete(sh.sseClients, clientID)
		close(clientChan)
		sh.sseMutex.Unlock()
		log.Debug().Str("clientID", clientID).Msg("SSE client disconnected")
	}()

	log.Debug().Str("clientID", clientID).Msg("SSE client connected")

	// Send initial ping
	fmt.Fprintf(w, "data: {\"type\":\"connected\",\"clientId\":\"%s\"}\n\n", clientID)
	w.(http.Flusher).Flush()

	// Listen for messages or client disconnect
	for {
		select {
		case message, ok := <-clientChan:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", message)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// BroadcastSSE sends a message to all connected SSE clients
func (sh *SSEHandler) BroadcastSSE(message string) {
	sh.sseMutex.RLock()
	defer sh.sseMutex.RUnlock()

	for clientID, clientChan := range sh.sseClients {
		select {
		case clientChan <- message:
		default:
			// Channel is full, skip this client
			log.Warn().Str("clientID", clientID).Msg("SSE client channel full, skipping message")
		}
	}
}

// monitorNewRequests watches for new HTTP requests and broadcasts updates
func (sh *SSEHandler) monitorNewRequests() {
	lastRequestCount := 0
	lastExecutionCount := 0

	ticker := time.NewTicker(1 * time.Second) // Check every second
	defer ticker.Stop()

	for range ticker.C {
		// Check for new HTTP requests
		stats := sh.logger.GetStats()
		if totalRequests, ok := stats["totalRequests"].(int); ok && totalRequests > lastRequestCount {
			message := fmt.Sprintf("{\"type\":\"newRequest\",\"count\":%d}", totalRequests)
			sh.BroadcastSSE(message)
			lastRequestCount = totalRequests
		}

		// Check for new script executions
		if result, err := sh.repos.Executions().ListExecutions(context.Background(), repository.ExecutionFilter{}, repository.PaginationOptions{Limit: 1, Offset: 0}); err == nil {
			if result.Total > lastExecutionCount {
				message := fmt.Sprintf("{\"type\":\"newExecution\",\"count\":%d}", result.Total)
				sh.BroadcastSSE(message)
				lastExecutionCount = result.Total
			}
		}
	}
}
