package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	addr := flag.String("http.addr", ":8080", "HTTP listen address")
	flag.Parse()

	logger := log.New(os.Stdout, "employee-service ", log.LstdFlags|log.LUTC)
	repo := NewMemoryRepository()
	service := NewEmployeeService(repo)
	server := &http.Server{
		Addr: *addr, Handler: NewHTTPHandler(service, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		logger.Printf("HTTP server listening on %s", *addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("HTTP server: %v", err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil { logger.Printf("graceful shutdown failed: %v", err) }
}
