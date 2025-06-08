package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/entity/sqlentity"
	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkgerror"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	Role int64  `json:"role"`
	Sub  string `json:"sub"`
	jwt.RegisteredClaims
}

// JWTMiddleware extracts the role from JWT token and puts it in the context
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			err := pkgerror.NewBusinessError("missing authorization header")
			pkgerror.WriteError(w, *err)
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			err := pkgerror.NewBusinessError("invalid authorization header format")
			pkgerror.WriteError(w, *err)
			return
		}

		// Get the token
		token := parts[1]

		// Split the token into parts
		tokenParts := strings.Split(token, ".")
		if len(tokenParts) != 3 {
			err := pkgerror.NewBusinessError("invalid token format")
			pkgerror.WriteError(w, *err)
			return
		}

		// Decode the claims (second part of the token)
		claimsBytes, err := base64Decode(tokenParts[1])
		if err != nil {
			err := pkgerror.NewBusinessError("invalid token claims")
			pkgerror.WriteError(w, *err)
			return
		}

		// Parse the claims
		var claims JWTClaims
		if err := json.Unmarshal(claimsBytes, &claims); err != nil {
			err := pkgerror.NewBusinessError("invalid token claims format")
			pkgerror.WriteError(w, *err)
			return
		}

		// Convert role string to EmployeeRoles
		var role sqlentity.EmployeeRoles
		switch claims.Role {
		case 1:
			role = sqlentity.Admin
		case 0:
			role = sqlentity.Employee
		default:
			err := pkgerror.NewBusinessError("invalid role in token")
			pkgerror.WriteError(w, *err)
			return
		}

		// Add user ID and role to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.Sub)
		ctx = context.WithValue(ctx, UserRoleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// base64Decode decodes a base64 string
func base64Decode(s string) ([]byte, error) {
	// Add padding if needed
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}

	// Replace URL-safe characters
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")

	// Decode
	return base64.StdEncoding.DecodeString(s)
}
