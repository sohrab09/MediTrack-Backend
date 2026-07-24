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
	"meditrack-backend/internal/handlers/login"
	"meditrack-backend/internal/handlers/medicine_category"
	"meditrack-backend/internal/handlers/medicine_units"
	"meditrack-backend/internal/handlers/register"
	"meditrack-backend/internal/handlers/users"
	"meditrack-backend/router"
)

func Serve() {
	cfg := config.LoadConfig()
	db := database.ConnectPostgres(cfg)
	defer db.Close()

	mux := http.NewServeMux()

	// 🔹 Base API Version / Prefix Variable
	apiPrefix := "/api/v1"

	handle := func(pattern string, handler http.HandlerFunc) {
		var fullPattern string
		spaceIndex := -1
		for i := 0; i < len(pattern); i++ {
			if pattern[i] == ' ' {
				spaceIndex = i
				break
			}
		}

		if spaceIndex != -1 {
			method := pattern[:spaceIndex]
			path := pattern[spaceIndex+1:]
			fullPattern = method + " " + apiPrefix + path
		} else {
			fullPattern = apiPrefix + pattern
		}

		mux.HandleFunc(fullPattern, handler)
	}

	// Health Check
	handle("GET /health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Auth
	handle("POST /auth/login", login.LoginHandler(db))
	handle("POST /auth/register", register.RegisterHandler(db))

	// Users
	handle("GET /users", users.GetUsers(db))
	handle("GET /users/{id}", users.GetUser(db))
	handle("PUT /users/{id}", users.UpdateUser(db))
	handle("DELETE /users/{id}", users.DeleteUser(db))

	// Medicine Categories
	handle("POST /medicine-categories", medicine_category.CreateMedicineCategories(db))
	handle("GET /medicine-categories", medicine_category.GetMedicineCategories(db))
	handle("GET /medicine-categories/{id}", medicine_category.GetMedicineCategoryByID(db))
	handle("PUT /medicine-categories/{id}", medicine_category.UpdateMedicineCategory(db))
	handle("DELETE /medicine-categories/{id}", medicine_category.DeleteMedicineCategory(db))

	// Medicine Units
	handle("POST /medicine-units", medicine_units.CreateMedicineUnits(db))
	handle("GET /medicine-units", medicine_units.GetMedicineUnits(db))
	handle("GET /medicine-units/{id}", medicine_units.GetMedicineUnitByID(db))
	handle("PUT /medicine-units/{id}", medicine_units.UpdateMedicineUnits(db))
	handle("DELETE /medicine-units/{id}", medicine_units.DeleteMedicineUnits(db))

	// Global Router / Middleware Integration
	globalHandler := router.GlobalRouter(mux)

	// Server Config
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      globalHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 🔹 Server Execution with Graceful Shutdown
	go func() {
		log.Printf("🚀 MediTrack server running on port %s (Prefix: %s)", cfg.Port, apiPrefix)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server ListenAndServe error: %v", err)
		}
	}()

	// Listen for OS Interrupt Signals (SIGINT, SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("⚡ Shutting down server gracefully...")

	// Timeout Context for Shutdown (5 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped cleanly")
}
