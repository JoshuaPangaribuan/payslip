package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/entity/sqlentity"
	"github.com/stretchr/testify/assert"
)

type testHandler struct {
	called bool
	ctx    context.Context
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
	h.ctx = r.Context()
	w.WriteHeader(http.StatusOK)
}

func makeJWTClaims(role int64, sub string) string {
	claims := JWTClaims{
		Role: role,
		Sub:  sub,
	}
	claimsBytes, _ := json.Marshal(claims)
	claimsB64 := base64.StdEncoding.EncodeToString(claimsBytes)
	// JWT: header.claims.signature
	return "header." + strings.TrimRight(claimsB64, "=") + ".signature"
}

func TestJWTMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		header         string
		expectedStatus int
		expectHandler  bool
		expectUserID   string
		expectRole     sqlentity.EmployeeRoles
		expectBody     string
	}{
		{
			name:           "valid admin token",
			header:         "Bearer " + makeJWTClaims(1, "admin-uuid"),
			expectedStatus: http.StatusOK,
			expectHandler:  true,
			expectUserID:   "admin-uuid",
			expectRole:     sqlentity.Admin,
		},
		{
			name:           "valid employee token",
			header:         "Bearer " + makeJWTClaims(0, "employee-uuid"),
			expectedStatus: http.StatusOK,
			expectHandler:  true,
			expectUserID:   "employee-uuid",
			expectRole:     sqlentity.Employee,
		},
		{
			name:           "missing header",
			header:         "",
			expectedStatus: http.StatusForbidden,
			expectHandler:  false,
			expectBody:     "missing authorization header",
		},
		{
			name:           "invalid format",
			header:         "BearerOnlyOnePart",
			expectedStatus: http.StatusForbidden,
			expectHandler:  false,
			expectBody:     "invalid authorization header format",
		},
		{
			name:           "invalid token format",
			header:         "Bearer part1.part2",
			expectedStatus: http.StatusForbidden,
			expectHandler:  false,
			expectBody:     "invalid token format",
		},
		{
			name:           "invalid claims base64",
			header:         "Bearer part1.invalidbase64.part3",
			expectedStatus: http.StatusForbidden,
			expectHandler:  false,
			expectBody:     "invalid token claims",
		},
		{
			name:           "invalid claims json",
			header:         "Bearer part1." + base64.StdEncoding.EncodeToString([]byte("notjson")) + ".part3",
			expectedStatus: http.StatusForbidden,
			expectHandler:  false,
			expectBody:     "invalid token claims format",
		},
		{
			name:           "invalid role",
			header:         "Bearer " + makeJWTClaims(99, "user-uuid"),
			expectedStatus: http.StatusForbidden,
			expectHandler:  false,
			expectBody:     "invalid role in token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &testHandler{}
			wrapped := JWTMiddleware(handler)
			req := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rw := httptest.NewRecorder()
			wrapped.ServeHTTP(rw, req)

			assert.Equal(t, tt.expectedStatus, rw.Code, "status code")
			if tt.expectHandler {
				assert.True(t, handler.called, "handler should be called")
				ctx := handler.ctx
				userID := ctx.Value(UserIDKey)
				role := ctx.Value(UserRoleKey)
				assert.Equal(t, tt.expectUserID, userID)
				assert.Equal(t, tt.expectRole, role)
			} else {
				assert.False(t, handler.called, "handler should not be called")
				if tt.expectBody != "" {
					assert.Contains(t, rw.Body.String(), tt.expectBody)
				}
			}
		})
	}
}
