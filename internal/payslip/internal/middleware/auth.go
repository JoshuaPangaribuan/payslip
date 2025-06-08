package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/entity/sqlentity"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkgerror"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserRoleKey  contextKey = "user_role"
	IPAddressKey contextKey = "ip_address"
	UserIDKey    contextKey = "user_id"
	RequestIDKey contextKey = "request_id"
)

// AdminOnly is a middleware that ensures only admin users can access the endpoint
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(UserRoleKey).(sqlentity.EmployeeRoles)
		if !ok {
			err := pkgerror.NewBusinessError("unauthorized")
			pkgerror.WriteError(w, *err)
			return
		}

		if role != sqlentity.Admin {
			err := pkgerror.NewBusinessError("admin access required")
			pkgerror.WriteError(w, *err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// EmployeeOnly is a middleware that ensures only employee users can access the endpoint
func EmployeeOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(UserRoleKey).(sqlentity.EmployeeRoles)
		if !ok {
			err := pkgerror.NewBusinessError("unauthorized")
			pkgerror.WriteError(w, *err)
			return
		}

		if role != sqlentity.Employee {
			err := pkgerror.NewBusinessError("employee access required")
			pkgerror.WriteError(w, *err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// WithUserRole adds the user role to the request context
func WithUserRole(role sqlentity.EmployeeRoles) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), UserRoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserRole retrieves the user role from the context
func GetUserRole(ctx context.Context) interface{} {
	return ctx.Value(UserRoleKey)
}

// WithIPAddress adds the IP address to the context
func WithIPAddress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIPAddress(r)
		ctx := context.WithValue(r.Context(), IPAddressKey, ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetIPAddress retrieves the IP address from the context
func GetIPAddress(ctx context.Context) string {
	if ip, ok := ctx.Value(IPAddressKey).(string); ok {
		return ip
	}
	return ""
}

// getIPAddress extracts the real IP address from the request
func getIPAddress(r *http.Request) string {
	// Try to get IP from X-Real-IP header
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Try to get IP from X-Forwarded-For header
	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// If no proxy headers, get IP from RemoteAddr
	ip = r.RemoteAddr
	if ip != "" {
		// Remove port if present
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
		return ip
	}

	return ""
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok && id != "" {
		return id
	}
	// Generate a new UUID if none exists
	return uuid.New().String()
}

// WithRequestID adds a request ID to the context
func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get request ID from header or generate new one
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
