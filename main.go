package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
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
	repo, closeRepo, err := buildRepository(logger)
	if err != nil {
		logger.Fatalf("initialize repository: %v", err)
	}
	defer closeRepo()
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

func buildRepository(logger *log.Logger) (EmployeeRepository, func(), error) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		logger.Printf("DB_HOST is not set; using in-memory repository")
		return NewMemoryRepository(), func() {}, nil
	}

	port := envOrDefault("DB_PORT", "5432")
	name := envOrDefault("DB_NAME", "employees")
	user := envOrDefault("DB_USER", "employee_user")
	password := os.Getenv("DB_PASSWORD")
	sslmode := envOrDefault("DB_SSLMODE", "disable")
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host, port, name, user, password, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("open PostgreSQL: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		if err := db.PingContext(ctx); err == nil {
			break
		} else if ctx.Err() != nil {
			_ = db.Close()
			return nil, nil, fmt.Errorf("wait for PostgreSQL: %w", ctx.Err())
		}
		<-ticker.C
	}
	logger.Printf("connected to PostgreSQL at %s:%s database=%s", host, port, name)

	repo, err := NewPostgresRepository(db)
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	return repo, func() { _ = db.Close() }, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
