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

	"google.golang.org/grpc"
)

func main() {
	httpAddr := flag.String("http.addr", ":8080", "HTTP listen address")
	grpcAddr := flag.String("grpc.addr", ":9090", "gRPC listen address")
	flag.Parse()

	logger := log.New(os.Stdout, "employee-service ", log.LstdFlags|log.LUTC)
	repo := NewMemoryRepository()
	service := NewEmployeeService(repo)

	grpcListener, err := listenGRPC(*grpcAddr)
	if err != nil {
		logger.Fatalf("gRPC listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	registerEmployeeGRPCServer(grpcServer, newEmployeeGRPCServer(service))

	httpServer := &http.Server{
		Addr: *httpAddr, Handler: NewHTTPHandler(service, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Printf("HTTP server listening on %s", *httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("HTTP server: %v", err)
			stop()
		}
	}()
	go func() {
		logger.Printf("gRPC server listening on %s", *grpcAddr)
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Printf("gRPC server: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Printf("HTTP graceful shutdown failed: %v", err)
	}
	grpcDone := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcDone)
	}()
	select {
	case <-grpcDone:
	case <-shutdownCtx.Done():
		grpcServer.Stop()
	}
}
