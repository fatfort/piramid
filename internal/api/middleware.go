package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"piramid/internal/database"
)

// jwtAuthMiddleware validates JWT tokens
func (s *Server) jwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header or cookie
		var tokenString string

		// Try Authorization header first
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// If not in header, try cookie
		if tokenString == "" {
			if cookie, err := r.Cookie("auth_token"); err == nil {
				tokenString = cookie.Value
			}
		}

		if tokenString == "" {
			s.respondError(w, http.StatusUnauthorized, "No authentication token provided")
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(s.cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			s.respondError(w, http.StatusUnauthorized, "Invalid authentication token")
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			s.respondError(w, http.StatusUnauthorized, "Invalid token claims")
			return
		}

		// Get user ID and tenant ID from claims
		userID, ok := claims["user_id"].(float64)
		if !ok {
			s.respondError(w, http.StatusUnauthorized, "Invalid user ID in token")
			return
		}

		tenantID, ok := claims["tenant_id"].(float64)
		if !ok {
			s.respondError(w, http.StatusUnauthorized, "Invalid tenant ID in token")
			return
		}

		// Verify user is still active
		var user database.User
		if err := s.db.Where("id = ? AND active = ?", uint(userID), true).First(&user).Error; err != nil {
			s.respondError(w, http.StatusUnauthorized, "User account is inactive")
			return
		}

		// Add user and tenant info to context
		ctx := context.WithValue(r.Context(), "user_id", uint(userID))
		ctx = context.WithValue(ctx, "tenant_id", uint(tenantID))
		ctx = context.WithValue(ctx, "user_email", claims["email"])
		ctx = context.WithValue(ctx, "user_role", claims["role"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// tenantMiddleware ensures tenant context is available
func (s *Server) tenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := getTenantIDFromContext(r)
		if tenantID == 0 {
			s.respondError(w, http.StatusUnauthorized, "No tenant context available")
			return
		}

		// Verify tenant exists and is active
		var tenant database.Tenant
		if err := s.db.Where("id = ?", tenantID).First(&tenant).Error; err != nil {
			s.respondError(w, http.StatusUnauthorized, "Invalid tenant")
			return
		}

		// Add tenant info to context
		ctx := context.WithValue(r.Context(), "tenant", &tenant)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// roleMiddleware checks if user has required role
func (s *Server) roleMiddleware(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value("user_role")
			if userRole == nil {
				s.respondError(w, http.StatusForbidden, "No role information available")
				return
			}

			role, ok := userRole.(string)
			if !ok {
				s.respondError(w, http.StatusForbidden, "Invalid role information")
				return
			}

			// Simple role check (in production, you might want more sophisticated RBAC)
			if role != requiredRole && role != "admin" {
				s.respondError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
