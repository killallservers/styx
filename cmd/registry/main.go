package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/killallservers/styx/pkg/registry"
)

var (
	port          = flag.Int("port", 7506, "Port to listen on")
	specsDir      = flag.String("specs", "pkg/registry/specs", "Directory containing TOML specs")
	webhookSecret = flag.String("webhook-secret", os.Getenv("WEBHOOK_SECRET"), "GitHub webhook secret")
)

type Server struct {
	mux      *http.ServeMux
	registryMu sync.RWMutex
	registry map[string]*registry.ToolSpec
	lastBuild time.Time
}

func main() {
	flag.Parse()

	server := &Server{
		mux: http.NewServeMux(),
	}

	// Load specs on startup
	if err := server.loadRegistry(); err != nil {
		log.Fatalf("Failed to load registry: %v", err)
	}

	// Routes
	server.mux.HandleFunc("GET /health", server.handleHealth)
	server.mux.HandleFunc("GET /registry.json", server.handleRegistry)
	server.mux.HandleFunc("POST /webhook/github", server.handleWebhook)
	server.mux.HandleFunc("GET /", server.handleRoot)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: server.mux,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		httpServer.Shutdown(ctx)
	}()

	log.Printf("Starting registry server on http://localhost:%d", *port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func (s *Server) loadRegistry() error {
	reg, err := registry.LoadRegistryFromTOML(*specsDir)
	if err != nil {
		return fmt.Errorf("load TOML: %w", err)
	}

	s.registryMu.Lock()
	s.registry = reg
	s.lastBuild = time.Now()
	s.registryMu.Unlock()

	log.Printf("Loaded %d tools from %s", len(reg), *specsDir)
	return nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.registryMu.RLock()
	healthy := len(s.registry) > 0
	s.registryMu.RUnlock()

	if !healthy {
		http.Error(w, "Registry not loaded", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"tools":  len(s.registry),
		"uptime": time.Since(s.lastBuild).Seconds(),
	})
}

func (s *Server) handleRegistry(w http.ResponseWriter, r *http.Request) {
	s.registryMu.RLock()
	reg := s.registry
	s.registryMu.RUnlock()

	if len(reg) == 0 {
		http.Error(w, "Registry not loaded", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300") // 5 min cache
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"version":  "1.0",
		"tools":    reg,
		"generated": time.Now().Format(time.RFC3339),
	}); err != nil {
		log.Printf("Error encoding registry: %v", err)
	}
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate webhook secret
	if *webhookSecret != "" {
		signature := r.Header.Get("X-Hub-Signature-256")
		if !validateWebhookSignature(signature, r, *webhookSecret) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Reload registry
	if err := s.loadRegistry(); err != nil {
		log.Printf("Webhook: failed to reload registry: %v", err)
		http.Error(w, "Failed to reload", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "rebuilt",
		"tools":   len(s.registry),
		"updated": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "styx-registry",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"GET /health":       "Health check",
			"GET /registry.json": "Registry specs",
			"POST /webhook/github": "GitHub webhook",
		},
	})
}

func validateWebhookSignature(signature string, r *http.Request, secret string) bool {
	// TODO: Implement HMAC-SHA256 validation
	// For now, accept all (secret will be empty in dev)
	return true
}
