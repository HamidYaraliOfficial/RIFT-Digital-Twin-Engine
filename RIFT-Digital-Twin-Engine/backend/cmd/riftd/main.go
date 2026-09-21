// Command riftd is the RIFT server: it hosts the REST + SSE API layer, runs
// the telemetry pipeline, rules/anomaly/health/alerting loops, and persists
// state under -data (default ./data).
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"rift/internal/api"
)

func main() {
	addr := flag.String("addr", envOr("RIFT_ADDR", ":8080"), "HTTP listen address")
	dataDir := flag.String("data", envOr("RIFT_DATA_DIR", "./data"), "directory for persisted state (snapshots, audit log)")
	flag.Parse()

	if err := os.MkdirAll(*dataDir, 0o755); err != nil {
		log.Fatalf("could not create data directory %s: %v", *dataDir, err)
	}

	server := api.NewServer(*dataDir)

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("shutting down: saving state...")
		cancel()
	}()

	go server.RunBackground(ctx)

	httpServer := &http.Server{Addr: *addr, Handler: server.Handler()}
	go func() {
		<-ctx.Done()
		_ = httpServer.Close()
	}()

	log.Printf("RIFT — Digital Twin Engine listening on %s (data dir: %s)", *addr, *dataDir)
	log.Printf("Default login: admin / admin123  (change this immediately in production)")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
