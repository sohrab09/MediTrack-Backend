package cmd

import (
	"log"
	"meditrack-backend/internal/config"
	"meditrack-backend/internal/database"
	addmedicinecategories "meditrack-backend/internal/handlers/add_medicine_categories"
	deleteuser "meditrack-backend/internal/handlers/delete-user"
	deletemedicinecategories "meditrack-backend/internal/handlers/delete_medicine_categories"
	getmedicinecategories "meditrack-backend/internal/handlers/get_medicine_categories"
	"meditrack-backend/internal/handlers/getuser"
	"meditrack-backend/internal/handlers/getusers"
	"meditrack-backend/internal/handlers/login"
	"meditrack-backend/internal/handlers/register"
	updateuser "meditrack-backend/internal/handlers/update-user"
	updatemedicinecategories "meditrack-backend/internal/handlers/update_medicine_categories"
	"meditrack-backend/router"
	"net/http"
	"time"
)

func Serve() {
	cfg := config.LoadConfig()
	db := database.ConnectPostgres(cfg)
	defer db.Close()

	mux := http.NewServeMux()
	globalHandler := router.GlobalRouter(mux)

	// Auth
	mux.HandleFunc("POST /api/v1/auth/login", login.LoginHandler(db))
	mux.HandleFunc("POST /api/v1/auth/register", register.RegisterHandler(db))

	// Users
	mux.HandleFunc("GET /api/v1/users", getusers.GetUsers(db))
	mux.HandleFunc("GET /api/v1/users/{id}", getuser.GetUser(db))
	mux.HandleFunc("PUT /api/v1/users/{id}", updateuser.UpdateUser(db))
	mux.HandleFunc("DELETE /api/v1/users/{id}", deleteuser.DeleteUser(db))

	// Medicine Categories
	mux.HandleFunc("POST /api/v1/medicine-categories", addmedicinecategories.CreateMedicineCategories(db))
	mux.HandleFunc("GET /api/v1/medicine-categories", getmedicinecategories.GetMedicineCategories(db))
	mux.HandleFunc("PUT /api/v1/medicine-categories/{id}", updatemedicinecategories.UpdateMedicineCategory(db))
	mux.HandleFunc("DELETE /api/v1/medicine-categories/{id}", deletemedicinecategories.DeleteMedicineCategory(db))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      globalHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 MediTrack server running on port %s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
