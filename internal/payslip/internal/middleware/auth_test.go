package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JoshuaPangaribuan/payslip/internal/payslip/internal/entity/sqlentity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAdminOnly(t *testing.T) {
	tests := []struct {
		name       string
		role       interface{}
		expectCode int
		shouldCall bool
	}{
		{"admin role", sqlentity.Admin, http.StatusOK, true},
		{"employee role", sqlentity.Employee, http.StatusForbidden, false},
		{"missing role", nil, http.StatusForbidden, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			h := AdminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest("GET", "/", nil)
			ctx := req.Context()
			if tt.role != nil {
				ctx = context.WithValue(ctx, UserRoleKey, tt.role)
			}
			req = req.WithContext(ctx)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)
			assert.Equal(t, tt.expectCode, rw.Code)
			assert.Equal(t, tt.shouldCall, called)
		})
	}
}

func TestEmployeeOnly(t *testing.T) {
	tests := []struct {
		name       string
		role       interface{}
		expectCode int
		shouldCall bool
	}{
		{"employee role", sqlentity.Employee, http.StatusOK, true},
		{"admin role", sqlentity.Admin, http.StatusForbidden, false},
		{"missing role", nil, http.StatusForbidden, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			h := EmployeeOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest("GET", "/", nil)
			ctx := req.Context()
			if tt.role != nil {
				ctx = context.WithValue(ctx, UserRoleKey, tt.role)
			}
			req = req.WithContext(ctx)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)
			assert.Equal(t, tt.expectCode, rw.Code)
			assert.Equal(t, tt.shouldCall, called)
		})
	}
}

func TestWithUserRoleAndGetUserRole(t *testing.T) {
	ctx := context.Background()
	role := sqlentity.Admin
	mw := WithUserRole(role)
	var gotRole interface{}
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRole = GetUserRole(r.Context())
	}))
	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req.WithContext(ctx))
	assert.Equal(t, role, gotRole)
}

func TestWithIPAddressAndGetIPAddress(t *testing.T) {
	t.Run("X-Real-IP header", func(t *testing.T) {
		h := WithIPAddress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetIPAddress(r.Context())
			assert.Equal(t, "1.2.3.4", ip)
		}))
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Real-IP", "1.2.3.4")
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
	})

	t.Run("X-Forwarded-For header", func(t *testing.T) {
		h := WithIPAddress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetIPAddress(r.Context())
			assert.Equal(t, "5.6.7.8", ip)
		}))
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Forwarded-For", "5.6.7.8, 9.10.11.12")
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
	})

	t.Run("RemoteAddr", func(t *testing.T) {
		h := WithIPAddress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetIPAddress(r.Context())
			assert.Equal(t, "127.0.0.1", ip)
		}))
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
	})

	t.Run("No IP", func(t *testing.T) {
		h := WithIPAddress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetIPAddress(r.Context())
			// The default RemoteAddr in httptest.NewRequest is 192.0.2.1
			assert.Equal(t, "192.0.2.1", ip)
		}))
		req := httptest.NewRequest("GET", "/", nil)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
	})
}

func TestGetUserID(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "user-123")
	assert.Equal(t, "user-123", GetUserID(ctx))
	assert.Equal(t, "", GetUserID(context.Background()))
}

func TestGetRequestID(t *testing.T) {
	id := uuid.New().String()
	ctx := context.WithValue(context.Background(), RequestIDKey, id)
	assert.Equal(t, id, GetRequestID(ctx))
	// Should generate new UUID if not present
	newID := GetRequestID(context.Background())
	_, err := uuid.Parse(newID)
	assert.NoError(t, err)
}

func TestWithRequestID(t *testing.T) {
	t.Run("header present", func(t *testing.T) {
		h := WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := GetRequestID(r.Context())
			assert.Equal(t, "req-abc", id)
		}))
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-Request-ID", "req-abc")
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
	})

	t.Run("header missing", func(t *testing.T) {
		h := WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := GetRequestID(r.Context())
			_, err := uuid.Parse(id)
			assert.NoError(t, err)
		}))
		req := httptest.NewRequest("GET", "/", nil)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
	})
}
