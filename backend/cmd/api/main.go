package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kebaikankuid/kebaikanku/backend/internal/config"
	"github.com/kebaikankuid/kebaikanku/backend/internal/database"
)

func main() {
	// 1. Load Configurations
	cfg := config.Load()

	// 2. Initialize Database and Auto-migrate
	database.Init(cfg)

	// 3. Setup Router
	r := chi.NewRouter()

	// Standard middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS config
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "https://*.pages.dev", "https://kebaikanku.id", "https://*.kebaikanku.id"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by browsers
	}))

	// 4. Routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","time":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// 5. Start Server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("Backend API server is running on http://localhost%s in %s mode\n", serverAddr, cfg.Env)

	err := http.ListenAndServe(serverAddr, r)
	if err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}
