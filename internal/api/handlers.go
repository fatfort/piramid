package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"piramid/internal/database"
	"piramid/internal/messaging"
)

// Authentication handlers

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string         `json:"token"`
	User  *database.User `json:"user"`
}

// login handles user authentication
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Find user by email
	var user database.User
	if err := s.db.Preload("Tenant").Where("email = ? AND active = ?", req.Email, true).First(&user).Error; err != nil {
		s.respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   user.ID,
		"tenant_id": user.TenantID,
		"email":     user.Email,
		"role":      user.Role,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Set HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		HttpOnly: true,
		Secure:   s.cfg.Environment == "production",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   24 * 60 * 60, // 24 hours
		Path:     "/",
	})

	s.respondSuccess(w, LoginResponse{
		Token: tokenString,
		User:  &user,
	})
}

// logout handles user logout
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	// Clear the auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		HttpOnly: true,
		Secure:   s.cfg.Environment == "production",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Path:     "/",
	})

	s.respondSuccess(w, map[string]string{"message": "Logged out successfully"})
}

// Event handlers

// streamEvents handles Server-Sent Events for real-time event streaming
func (s *Server) streamEvents(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	tenantID := getTenantIDFromContext(r)

	// Create a context for the SSE connection
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Subscribe to NATS events
	consumer := messaging.NewConsumer(s.js)

	// Channel to receive events
	eventChan := make(chan []byte, 100)

	// Start consuming events in a goroutine
	go func() {
		consumer.ConsumeEvents(ctx, func(data []byte) error {
			select {
			case eventChan <- data:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}()

	// Send events to client
	for {
		select {
		case eventData := <-eventChan:
			// Parse the event to check if it belongs to this tenant
			var event map[string]interface{}
			if err := json.Unmarshal(eventData, &event); err != nil {
				continue
			}

			// For now, send all events (in a real system, you'd filter by tenant)
			fmt.Fprintf(w, "data: %s\n\n", string(eventData))

			// Flush the data to the client
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

		case <-ctx.Done():
			return
		}
	}
}

// getEvents returns paginated events
func (s *Server) getEvents(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	var events []database.Event
	var total int64

	// Count total events
	s.db.Model(&database.Event{}).Where("tenant_id = ?", tenantID).Count(&total)

	// Get events with pagination
	if err := s.db.Where("tenant_id = ?", tenantID).
		Order("timestamp DESC").
		Offset(offset).
		Limit(limit).
		Find(&events).Error; err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to fetch events")
		return
	}

	response := map[string]interface{}{
		"events": events,
		"pagination": map[string]interface{}{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	s.respondSuccess(w, response)
}

// Statistics handlers

// getSSHStats returns SSH brute-force statistics
func (s *Server) getSSHStats(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)

	// Get SSH events from the last 24 hours
	var stats []struct {
		SrcIP     string  `json:"src_ip"`
		Country   string  `json:"country"`
		Attempts  int     `json:"attempts"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	query := `
		SELECT src_ip, country, latitude, longitude, COUNT(*) as attempts
		FROM events 
		WHERE tenant_id = ? 
		AND event_type = 'ssh' 
		AND timestamp > NOW() - INTERVAL '24 hours'
		GROUP BY src_ip, country, latitude, longitude
		ORDER BY attempts DESC
		LIMIT 100
	`

	if err := s.db.Raw(query, tenantID).Scan(&stats).Error; err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to fetch SSH stats")
		return
	}

	s.respondSuccess(w, stats)
}

// getOverviewStats returns general statistics
func (s *Server) getOverviewStats(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)

	// Get various statistics
	var totalEvents int64
	var totalUniqueIPs int64
	var totalBannedIPs int64
	var recentEvents int64

	// Total events
	s.db.Model(&database.Event{}).Where("tenant_id = ?", tenantID).Count(&totalEvents)

	// Unique source IPs
	s.db.Model(&database.Event{}).Where("tenant_id = ?", tenantID).
		Distinct("src_ip").Count(&totalUniqueIPs)

	// Total banned IPs
	s.db.Model(&database.IPBan{}).Where("tenant_id = ?", tenantID).Count(&totalBannedIPs)

	// Recent events (last hour)
	s.db.Model(&database.Event{}).Where("tenant_id = ? AND timestamp > NOW() - INTERVAL '1 hour'", tenantID).
		Count(&recentEvents)

	// Top countries by events
	var topCountries []struct {
		Country string `json:"country"`
		Count   int64  `json:"count"`
	}

	s.db.Model(&database.Event{}).Where("tenant_id = ?", tenantID).
		Select("country, COUNT(*) as count").
		Group("country").
		Order("count DESC").
		Limit(10).
		Scan(&topCountries)

	stats := map[string]interface{}{
		"total_events":  totalEvents,
		"unique_ips":    totalUniqueIPs,
		"banned_ips":    totalBannedIPs,
		"recent_events": recentEvents,
		"top_countries": topCountries,
		"last_updated":  time.Now().Unix(),
	}

	s.respondSuccess(w, stats)
}

// IP Ban handlers

type BanIPRequest struct {
	IP     string `json:"ip"`
	Reason string `json:"reason"`
}

// banIP handles IP banning
func (s *Server) banIP(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	userID := getUserIDFromContext(r)

	var req BanIPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.IP == "" {
		s.respondError(w, http.StatusBadRequest, "IP address is required")
		return
	}

	// Check if IP is already banned
	var existingBan database.IPBan
	if err := s.db.Where("tenant_id = ? AND ip = ?", tenantID, req.IP).First(&existingBan).Error; err == nil {
		s.respondError(w, http.StatusConflict, "IP is already banned")
		return
	}

	// Create IP ban record
	ban := database.IPBan{
		TenantID: tenantID,
		IP:       req.IP,
		Reason:   req.Reason,
		BannedBy: userID,
	}

	if err := s.db.Create(&ban).Error; err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to ban IP")
		return
	}

	// Publish ban action to NATS
	publisher := messaging.NewPublisher(s.js)
	publisher.PublishBanAction(req.IP, req.Reason)

	s.respondSuccess(w, ban)
}

// getBans returns list of banned IPs
func (s *Server) getBans(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)

	var bans []database.IPBan
	if err := s.db.Where("tenant_id = ?", tenantID).
		Preload("User").
		Order("created_at DESC").
		Find(&bans).Error; err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to fetch banned IPs")
		return
	}

	s.respondSuccess(w, bans)
}

// unbanIP handles IP unbanning
func (s *Server) unbanIP(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)

	banID, err := parseUintParam(r, "id")
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid ban ID")
		return
	}

	// Delete the ban record
	if err := s.db.Where("tenant_id = ? AND id = ?", tenantID, banID).Delete(&database.IPBan{}).Error; err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to unban IP")
		return
	}

	s.respondSuccess(w, map[string]string{"message": "IP unbanned successfully"})
}
