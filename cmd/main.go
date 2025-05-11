package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dimakirio/calculatorv1/internal/orchestrator"
	"github.com/dimakirio/calculatorv1/pkg/config"
	"github.com/dimakirio/calculatorv1/pkg/logger"
)

func loggingMiddleware(next http.HandlerFunc, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info(fmt.Sprintf("%s %s %s", r.Method, r.RequestURI, time.Since(start)))
	}
}

func panicMiddleware(next http.HandlerFunc, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error(fmt.Sprintf("Panic recovered: %v", err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	}
}

func main() {
	cfg := config.LoadConfig()
	log := logger.NewLogger(cfg.LogLevel)

	orchestrator := orchestrator.NewOrchestrator(log, cfg)
	
	// Create a new mux and wrap handlers with middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", panicMiddleware(loggingMiddleware(orchestrator.HandleCalculate, log), log))
	mux.HandleFunc("/api/v1/expressions", panicMiddleware(loggingMiddleware(orchestrator.HandleGetExpressions, log), log))
	mux.HandleFunc("/api/v1/expressions/", panicMiddleware(loggingMiddleware(orchestrator.HandleGetExpressionByID, log), log))
	mux.HandleFunc("/api/v1/register", panicMiddleware(loggingMiddleware(orchestrator.HandleRegister, log), log))
	mux.HandleFunc("/api/v1/login", panicMiddleware(loggingMiddleware(orchestrator.HandleLogin, log), log))

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Info(fmt.Sprintf("Starting server on :%s", cfg.ServerPort))
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatal(fmt.Sprintf("Server error: %v", err))

	case sig := <-shutdown:
		log.Info(fmt.Sprintf("Server is shutting down: %v", sig))

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := server.Shutdown(ctx); err != nil {
			log.Error(fmt.Sprintf("Graceful shutdown did not complete in 10s: %v", err))
			if err := server.Close(); err != nil {
				log.Fatal(fmt.Sprintf("Could not stop server: %v", err))
			}
		}
	}
}
