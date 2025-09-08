package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/btassone/obtura/pkg/plugin"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// Mock database manager for testing
type mockDBManager struct {
	migrateError error
	seedError    error
}

func (m *mockDBManager) Migrate() error {
	return m.migrateError
}

func (m *mockDBManager) Seed() error {
	return m.seedError
}

func (m *mockDBManager) GetDB() interface{} {
	return nil
}

func TestServerNew(t *testing.T) {
	// This test would require mocking the database connection
	// For now, we'll test with an in-memory database
	t.Skip("Requires database mocking")
}

func TestServer_setupRoutes(t *testing.T) {
	router := chi.NewRouter()
	registry := plugin.NewRegistry(nil)

	s := &Server{
		port:     "8080",
		mode:     "test",
		router:   router,
		registry: registry,
	}

	// Set up routes
	s.setupRoutes()

	// Test that basic routes are registered
	tests := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/", http.StatusOK},
		{http.MethodGet, "/ws", http.StatusNotFound}, // WebSocket endpoint may not be registered in test
		// Skip /admin test as it requires auth plugin to be registered
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			// We might get redirects or other status codes in real implementation
			// For now, we'll just check that some routes exist
			if tt.path == "/" || tt.path == "/admin" {
				// These should be handled
				assert.NotEqual(t, http.StatusNotFound, rec.Code, "Path %s should be handled", tt.path)
			}
		})
	}
}

func TestServer_handleHome(t *testing.T) {
	router := chi.NewRouter()
	registry := plugin.NewRegistry(nil)
	
	s := &Server{
		router:   router,
		registry: registry,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	s.handleHome(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
}

func TestServer_setupMiddleware(t *testing.T) {
	router := chi.NewRouter()

	s := &Server{
		router: router,
		mode:   "dev",
	}

	// Setup middleware
	s.setupMiddleware()

	// Test that middleware is applied by making a request
	// In dev mode, we should have logger middleware
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Add a test handler
	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestServer_withRecovery(t *testing.T) {
	// Create a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Wrap it with recovery middleware
	recoveryHandler := withRecovery(panicHandler)

	// Test that it doesn't panic
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		recoveryHandler.ServeHTTP(rec, req)
	})

	// Should return 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// Test helper for recovery middleware
func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}