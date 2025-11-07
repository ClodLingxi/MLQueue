package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"MLQueue/internal/config"
	"MLQueue/internal/database"
	"MLQueue/internal/queue"
	"MLQueue/internal/routes"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Printf("Starting MLQueue API Server (Environment: %s)", cfg.Server.Env)

	// Initialize database connections
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	if err := database.InitRedis(cfg); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// Initialize queue manager with worker pool
	queueManager := queue.NewQueueManager(cfg.Queue.WorkerCount)
	queueManager.Start()
	defer queueManager.Stop()

	// Setup routes
	router := routes.SetupRouter(queueManager)

	// Setup V2 routes (Python客户端驱动架构)
	routes.SetupV2Routes(router)

	log.Println("V1 API (云端调度): /v1/*")
	log.Println("V2 API (Python驱动): /v2/*")

	// Create HTTP server
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server is running on http://%s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited successfully")
}
