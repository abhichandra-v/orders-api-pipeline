// Command orders-api is a small, dependency-free HTTP service used as the
// subject of a CI/CD + container-hardening case study. It is deliberately
// production-shaped: structured logging, timeouts, health/readiness probes,
// and graceful shutdown — the things that make a service safe to run on
// Kubernetes, not just to demo.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// version is injected at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := getenv("PORT", "8080")
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Run the server in the background so main can wait for a shutdown signal.
	go func() {
		slog.Info("starting server", "addr", srv.Addr, "version", version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			os.Exit(1)
		}
	}()

	// Block until we receive SIGINT/SIGTERM (k8s sends SIGTERM on pod stop).
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	// Give in-flight requests up to 10s to finish before exiting.
	slog.Info("shutdown signal received, draining connections")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	slog.Info("stopped cleanly")
}

func routes() http.Handler {
	mux := http.NewServeMux()
	store := newOrderStore()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		// In a real service this would check downstream dependencies.
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("GET /orders", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, store.list())
	})
	mux.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Item string `json:"item"`
			Qty  int    `json:"qty"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
			return
		}
		if in.Item == "" || in.Qty <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "item required and qty must be > 0"})
			return
		}
		o := store.add(in.Item, in.Qty)
		writeJSON(w, http.StatusCreated, o)
	})

	return logging(mux)
}

// --- order store (in-memory, concurrency-safe) ---

type order struct {
	ID   int    `json:"id"`
	Item string `json:"item"`
	Qty  int    `json:"qty"`
}

type orderStore struct {
	mu     sync.Mutex
	nextID int
	orders []order
}

func newOrderStore() *orderStore { return &orderStore{nextID: 1} }

func (s *orderStore) add(item string, qty int) order {
	s.mu.Lock()
	defer s.mu.Unlock()
	o := order{ID: s.nextID, Item: item, Qty: qty}
	s.nextID++
	s.orders = append(s.orders, o)
	return o
}

func (s *orderStore) list() []order {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]order, len(s.orders))
	copy(out, s.orders)
	return out
}

// --- helpers ---

// logging is a tiny access-log middleware so every request is observable.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"dur_ms", time.Since(start).Milliseconds(),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
