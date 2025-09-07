package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdminAuth(t *testing.T) {
	// Create a test handler that sets a header when called
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "handler-called")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with AdminAuth middleware
	handler := AdminAuth(testHandler)

	// Test that middleware allows request through (current implementation)
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "handler-called", rec.Header().Get("X-Test"))
	assert.Equal(t, "OK", rec.Body.String())
}

func TestRequireAuth(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "handler-called")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with RequireAuth middleware
	handler := RequireAuth(testHandler)

	// Test that middleware allows request through (current implementation)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "handler-called", rec.Header().Get("X-Test"))
}

// TestMiddlewareChain tests that multiple middleware can be chained
func TestMiddlewareChain(t *testing.T) {
	// Track middleware execution order
	var executionOrder []string

	// Create test middleware that tracks execution
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "middleware1-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "middleware1-after")
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "middleware2-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "middleware2-after")
		})
	}

	// Final handler
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		executionOrder = append(executionOrder, "handler")
		w.WriteHeader(http.StatusOK)
	})

	// Chain middleware
	handler := middleware1(middleware2(AdminAuth(RequireAuth(finalHandler))))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify execution order
	expected := []string{
		"middleware1-before",
		"middleware2-before",
		"handler",
		"middleware2-after",
		"middleware1-after",
	}
	assert.Equal(t, expected, executionOrder)
}

// TestAdminAuthFuture shows what AdminAuth should do when properly implemented
func TestAdminAuthFuture(t *testing.T) {
	t.Skip("Test for future implementation of AdminAuth")

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Admin area"))
	})

	handler := AdminAuth(testHandler)

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "no auth token",
			setupRequest: func(r *http.Request) {
				// No auth header
			},
			expectedStatus: http.StatusSeeOther, // Redirect to login
		},
		{
			name: "invalid auth token",
			setupRequest: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer invalid-token")
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid non-admin token",
			setupRequest: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer valid-user-token")
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "valid admin token",
			setupRequest: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer valid-admin-token")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Admin area",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
			if tt.setupRequest != nil {
				tt.setupRequest(req)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}

// TestRequireAuthFuture shows what RequireAuth should do when properly implemented
func TestRequireAuthFuture(t *testing.T) {
	t.Skip("Test for future implementation of RequireAuth")

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Protected content"))
	})

	handler := RequireAuth(testHandler)

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "no auth",
			setupRequest: func(r *http.Request) {
				// No auth header
			},
			expectedStatus: http.StatusSeeOther, // Redirect to login
		},
		{
			name: "invalid token",
			setupRequest: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer invalid-token")
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid token",
			setupRequest: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer valid-token")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Protected content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.setupRequest != nil {
				tt.setupRequest(req)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}
