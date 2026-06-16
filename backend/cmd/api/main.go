package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/kebaikankuid/kebaikanku/backend/internal/config"
	"github.com/kebaikankuid/kebaikanku/backend/internal/database"
	"github.com/kebaikankuid/kebaikanku/backend/internal/domain"
	"gorm.io/gorm"
)

const (
	waitlistCooldown = 20 * time.Second
)

var (
	waitlistRequestWindow = map[string]time.Time{}
	waitlistRequestGuard  sync.Mutex
	waitlistEmailRegex    = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
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
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*", "https://*.pages.dev", "https://kebaikanku.id", "https://*.kebaikanku.id"},
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
	r.Post("/api/v1/waitlist", handleWaitlistSignup)

	// 5. Start Server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("Backend API server is running on http://localhost%s in %s mode\n", serverAddr, cfg.Env)

	err := http.ListenAndServe(serverAddr, r)
	if err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}

type waitlistPayload struct {
	Email       string `json:"email"`
	Website     string `json:"website"`
	Source      string `json:"source"`
	SubmittedAt int64  `json:"submitted_at"`
}

type apiErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiResponse struct {
	Success bool             `json:"success"`
	Data    map[string]any   `json:"data,omitempty"`
	Error   *apiErrorResponse `json:"error,omitempty"`
}

func handleWaitlistSignup(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	var req waitlistPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWaitlistError(w, http.StatusBadRequest, "INVALID_JSON", "Payload must be valid JSON.")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if !waitlistEmailRegex.MatchString(email) {
		writeWaitlistError(w, http.StatusBadRequest, "INVALID_EMAIL", "Email is required and must be valid.")
		return
	}

	if strings.TrimSpace(req.Website) != "" {
		writeWaitlistError(w, http.StatusBadRequest, "BOT_DETECTED", "Request rejected as suspicious.")
		return
	}

	if req.SubmittedAt == 0 || time.Now().UTC().Sub(time.Unix(req.SubmittedAt, 0).UTC()) < 2*time.Second {
		writeWaitlistError(w, http.StatusBadRequest, "BOT_DETECTED", "Form submission was too fast.")
		return
	}

	clientIP := getClientIP(r)
	if blockedByRateLimit(clientIP) {
		writeWaitlistError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many waitlist requests. Please try again later.")
		return
	}

	var exists domain.Waitlist
	err := database.DB.Where("email = ?", email).First(&exists).Error
	if err == nil {
		writeWaitlistError(w, http.StatusConflict, "DUPLICATE_EMAIL", "Email is already registered in the waitlist.")
		return
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		writeWaitlistError(w, http.StatusInternalServerError, "DB_ERROR", "Could not process waitlist registration.")
		return
	}

	entry := domain.Waitlist{
		ID:        randomID(),
		Email:     email,
		Source:    strings.TrimSpace(req.Source),
		IPAddress: clientIP,
		UserAgent: r.Header.Get("User-Agent"),
	}
	if err := database.DB.Create(&entry).Error; err != nil {
		writeWaitlistError(w, http.StatusConflict, "DUPLICATE_EMAIL", "Email is already registered in the waitlist.")
		return
	}

	response := apiResponse{
		Success: true,
		Data: map[string]any{
			"id": entry.ID,
		},
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

func writeWaitlistError(w http.ResponseWriter, status int, code, message string) {
	w.WriteHeader(status)
	response := apiResponse{
		Success: false,
		Error: &apiErrorResponse{
			Code:    code,
			Message: message,
		},
	}
	_ = json.NewEncoder(w).Encode(response)
}

func blockedByRateLimit(ip string) bool {
	now := time.Now().UTC()
	waitlistRequestGuard.Lock()
	defer waitlistRequestGuard.Unlock()

	lastRequest, ok := waitlistRequestWindow[ip]
	if ok && now.Sub(lastRequest) < waitlistCooldown {
		return true
	}
	waitlistRequestWindow[ip] = now
	return false
}

func getClientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}

func randomID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[0:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:16]))
}
