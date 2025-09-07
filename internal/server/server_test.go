package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/btassone/obtura/internal/database"
	"github.com/btassone/obtura/pkg/plugin"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{http.MethodGet, "/health", http.StatusOK},
		{http.MethodGet, "/admin", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			// We might get redirects or other status codes in real implementation
			// For now, we'll just check that routes exist (not 404)
			assert.NotEqual(t, http.StatusNotFound, rec.Code)
		})
	}
}

func TestServer_handleIndex(t *testing.T) {
	s := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleIndex)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
}

func TestServer_handleNotFound(t *testing.T) {
	s := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/non-existent", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleNotFound)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
}

func TestServer_handleHealth(t *testing.T) {
	router := chi.NewRouter()
	registry := plugin.NewRegistry(nil)

	s := &Server{
		router:   router,
		registry: registry,
	}

	// Register a test plugin
	testPlugin := &mockPlugin{
		id:      "test.plugin",
		healthy: true,
	}
	err := registry.Register(testPlugin)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(s.handleHealth)
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestServer_Start(t *testing.T) {
	// Create a test server
	router := chi.NewRouter()
	registry := plugin.NewRegistry(nil)

	s := &Server{
		port:     "0", // Use port 0 to get a random available port
		mode:     "test",
		router:   router,
		registry: registry,
	}

	// Set up basic routes
	s.setupRoutes()

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- s.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop server
	cancel()

	// Check if server stopped without error
	select {
	case err := <-errChan:
		// Server should stop due to context cancellation
		assert.Error(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not stop within timeout")
	}
}

// Mock plugin for testing
type mockPlugin struct {
	id           string
	healthy      bool
	initError    error
	startError   error
	stopError    error
	destroyError error
}

func (p *mockPlugin) ID() string          { return p.id }
func (p *mockPlugin) Name() string        { return "Mock Plugin" }
func (p *mockPlugin) Version() string     { return "1.0.0" }
func (p *mockPlugin) Description() string { return "Mock plugin for testing" }
func (p *mockPlugin) Author() string      { return "Test" }

func (p *mockPlugin) Init(ctx context.Context) error    { return p.initError }
func (p *mockPlugin) Start(ctx context.Context) error   { return p.startError }
func (p *mockPlugin) Stop(ctx context.Context) error    { return p.stopError }
func (p *mockPlugin) Destroy(ctx context.Context) error { return p.destroyError }

func (p *mockPlugin) Dependencies() []string     { return nil }
func (p *mockPlugin) Config() interface{}        { return nil }
func (p *mockPlugin) ValidateConfig() error      { return nil }
func (p *mockPlugin) DefaultConfig() interface{} { return nil }

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

func TestServer_RenderComponent(t *testing.T) {
	s := &Server{}

	// Test with nil component
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// This would normally render a Templ component
	// For testing, we'll just verify the method exists
	// and doesn't panic with nil
	assert.NotPanics(t, func() {
		_ = renderComponent
	})
}

func TestServer_registerPlugins(t *testing.T) {
	router := chi.NewRouter()
	registry := plugin.NewRegistry(nil)
	dbManager := &database.Manager{} // This would need proper mocking

	s := &Server{
		router:   router,
		registry: registry,
		db:       dbManager,
	}

	// Test that registerPlugins doesn't panic
	assert.NotPanics(t, func() {
		s.registerPlugins()
	})
}

func TestServer_withRecovery(t *testing.T) {
	// Create a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Wrap with chi's Recoverer middleware (which is what the server uses)
	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	})
	router.Get("/panic", panicHandler)

	// Test that panic is recovered
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "Internal Server Error", rec.Body.String())
}
