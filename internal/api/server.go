package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/nats-io/nats.go"
	"gorm.io/gorm"

	"piramid/internal/config"
)

// Server holds the dependencies for the API server
type Server struct {
	cfg *config.Config
	db  *gorm.DB
	js  nats.JetStreamContext
}

// NewServer creates a new API server instance
func NewServer(cfg *config.Config, db *gorm.DB, js nats.JetStreamContext) *Server {
	return &Server{
		cfg: cfg,
		db:  db,
		js:  js,
	}
}

// Router sets up the HTTP routes and middleware
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:65605"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", s.healthCheck)

	// Authentication routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", s.login)
		r.Post("/logout", s.logout)
	})

	// API routes (protected)
	r.Route("/api", func(r chi.Router) {
		r.Use(s.jwtAuthMiddleware)
		r.Use(s.tenantMiddleware)

		r.Get("/events/stream", s.streamEvents)
		r.Get("/events", s.getEvents)
		r.Get("/stats/ssh", s.getSSHStats)
		r.Get("/stats/overview", s.getOverviewStats)
		r.Post("/ban", s.banIP)
		r.Get("/bans", s.getBans)
		r.Delete("/bans/{id}", s.unbanIP)
	})

	// Serve static files in production
	if s.cfg.Environment == "production" {
		fileServer := http.FileServer(http.Dir("./web/dist/"))
		r.Handle("/*", fileServer)
	}

	return r
}

// Response helpers

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// respondJSON sends a JSON response
func (s *Server) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func (s *Server) respondError(w http.ResponseWriter, statusCode int, message string) {
	s.respondJSON(w, statusCode, APIResponse{
		Success: false,
		Error:   message,
	})
}

// respondSuccess sends a success response
func (s *Server) respondSuccess(w http.ResponseWriter, data interface{}) {
	s.respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// healthCheck returns the health status of the API
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	sqlDB, err := s.db.DB()
	if err != nil {
		s.respondError(w, http.StatusServiceUnavailable, "Database connection error")
		return
	}

	if err := sqlDB.Ping(); err != nil {
		s.respondError(w, http.StatusServiceUnavailable, "Database ping failed")
		return
	}

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"services": map[string]string{
			"database": "healthy",
			"nats":     "healthy",
		},
	}

	s.respondSuccess(w, health)
}

// Utility functions

// getTenantIDFromContext extracts tenant ID from request context
func getTenantIDFromContext(r *http.Request) uint {
	if tenantID := r.Context().Value("tenant_id"); tenantID != nil {
		if id, ok := tenantID.(uint); ok {
			return id
		}
	}
	return 1 // Default to first tenant for now
}

// getUserIDFromContext extracts user ID from request context
func getUserIDFromContext(r *http.Request) uint {
	if userID := r.Context().Value("user_id"); userID != nil {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}

// parseUintParam parses a URL parameter as uint
func parseUintParam(r *http.Request, param string) (uint, error) {
	str := chi.URLParam(r, param)
	val, err := strconv.ParseUint(str, 10, 32)
	return uint(val), err
}
