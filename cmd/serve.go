package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"meditrack-backend/internal/config"
	"meditrack-backend/internal/database"
	"meditrack-backend/internal/router"
)

func Serve() {
	cfg := config.LoadConfig()
	db := database.ConnectPostgres(cfg)
	defer db.Close()

	mux := http.NewServeMux()

	// 🔹 Setup All Routes and Middlewares
	handler := router.SetupRoutes(mux, db)

	// Server Config
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 🔹 Server Execution
	go func() {
		log.Printf("🚀 MediTrack server running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server ListenAndServe error: %v", err)
		}
	}()

	// Graceful Shutdown Setup
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("⚡ Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped cleanly")
}
