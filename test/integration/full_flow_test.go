package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	// "github.com/btassone/obtura/internal/database"
	// "github.com/btassone/obtura/internal/server"
	"github.com/btassone/obtura/pkg/plugin"
	"github.com/btassone/obtura/test/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer represents a test server instance
type TestServer struct {
	*httptest.Server
	router   *chi.Mux
	registry *plugin.Registry
}

// setupTestServer creates a fully configured test server
func setupTestServer(t *testing.T) *TestServer {
	// Create router and registry
	router := chi.NewRouter()
	registry := plugin.NewRegistry(router)

	// Create test server
	ts := httptest.NewServer(router)

	testServer := &TestServer{
		Server:   ts,
		router:   router,
		registry: registry,
	}

	// Register cleanup
	t.Cleanup(func() {
		ts.Close()
	})

	return testServer
}

func TestHomePage(t *testing.T) {
	ts := setupTestServer(t)

	// Add home route
	ts.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<h1>Welcome to Obtura</h1>"))
	})

	// Make request
	resp, err := http.Get(ts.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")

	// Read body
	body := make([]byte, 1024)
	n, _ := resp.Body.Read(body)
	assert.Contains(t, string(body[:n]), "Welcome to Obtura")
}

func TestHealthEndpoint(t *testing.T) {
	ts := setupTestServer(t)

	// Add health route
	ts.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		health := map[string]interface{}{
			"status": "healthy",
			"time":   time.Now().Unix(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	})

	// Make request
	resp, err := http.Get(ts.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Parse response
	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err)

	assert.Equal(t, "healthy", health["status"])
	assert.NotZero(t, health["time"])
}

func TestPluginRoutes(t *testing.T) {
	ts := setupTestServer(t)

	// Create and register a test plugin with routes
	plugin := &testutil.TestPlugin{
		NameFunc: func() string { return "Test Plugin" },
	}

	// Add plugin route
	ts.router.Get("/plugin/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Plugin route working"))
	})

	err := ts.registry.Register(plugin)
	require.NoError(t, err)

	// Test plugin route
	resp, err := http.Get(ts.URL + "/plugin/test")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := make([]byte, 100)
	n, _ := resp.Body.Read(body)
	assert.Equal(t, "Plugin route working", string(body[:n]))
}

func TestAdminAuthentication(t *testing.T) {
	ts := setupTestServer(t)
	jwt := testutil.DefaultTestJWT()

	// Mock admin middleware
	adminMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Simple token validation for testing
			parts := strings.Split(auth, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Add admin route
	ts.router.Route("/admin", func(r chi.Router) {
		r.Use(adminMiddleware)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Admin Dashboard"))
		})
	})

	// Test without auth
	resp, err := http.Get(ts.URL + "/admin/")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Test with auth
	token, err := jwt.GenerateAdminToken(1)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", ts.URL+"/admin/", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoginFlow(t *testing.T) {
	ts := setupTestServer(t)

	// Add login routes
	ts.router.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<form method="post" action="/login">
			<input name="email" type="email">
			<input name="password" type="password">
			<button type="submit">Login</button>
		</form>`))
	})

	ts.router.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		// Simple validation for testing
		if email == "admin@example.com" && password == "admin123" {
			// In real app, generate JWT token
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"token": "test-jwt-token",
				"role":  "admin",
			})
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid credentials",
			})
		}
	})

	// Test GET login page
	resp, err := http.Get(ts.URL + "/login")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test POST login with valid credentials
	formData := strings.NewReader("email=admin@example.com&password=admin123")
	resp, err = http.Post(ts.URL+"/login", "application/x-www-form-urlencoded", formData)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp map[string]string
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	require.NoError(t, err)
	assert.Equal(t, "test-jwt-token", loginResp["token"])
	assert.Equal(t, "admin", loginResp["role"])

	// Test POST login with invalid credentials
	formData = strings.NewReader("email=wrong@example.com&password=wrong")
	resp, err = http.Post(ts.URL+"/login", "application/x-www-form-urlencoded", formData)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestStaticAssets(t *testing.T) {
	ts := setupTestServer(t)

	// Add static file handler
	ts.router.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		// In real app, this would serve actual files
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css")
			w.Write([]byte("body { margin: 0; }"))
		} else if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
			w.Write([]byte("console.log('Hello');"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// Test CSS file
	resp, err := http.Get(ts.URL + "/static/styles.css")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/css", resp.Header.Get("Content-Type"))

	// Test JS file
	resp, err = http.Get(ts.URL + "/static/app.js")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/javascript", resp.Header.Get("Content-Type"))

	// Test non-existent file
	resp, err = http.Get(ts.URL + "/static/missing.txt")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAPIEndpoints(t *testing.T) {
	ts := setupTestServer(t)

	// Add API routes
	ts.router.Route("/api/v1", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				next.ServeHTTP(w, req)
			})
		})

		r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
			users := []map[string]interface{}{
				{"id": 1, "name": "User 1"},
				{"id": 2, "name": "User 2"},
			}
			json.NewEncoder(w).Encode(users)
		})

		r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
			var user map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
				return
			}

			user["id"] = 3
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(user)
		})
	})

	// Test GET users
	resp, err := http.Get(ts.URL + "/api/v1/users")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var users []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&users)
	require.NoError(t, err)
	assert.Len(t, users, 2)

	// Test POST user
	newUser := map[string]string{"name": "New User"}
	body, _ := json.Marshal(newUser)

	resp, err = http.Post(ts.URL+"/api/v1/users", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdUser map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdUser)
	require.NoError(t, err)
	assert.Equal(t, "New User", createdUser["name"])
	assert.Equal(t, float64(3), createdUser["id"])
}

func TestMiddlewareOrder(t *testing.T) {
	ts := setupTestServer(t)

	var order []string

	// Create middleware that tracks execution order
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "middleware1-before")
			next.ServeHTTP(w, r)
			order = append(order, "middleware1-after")
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "middleware2-before")
			next.ServeHTTP(w, r)
			order = append(order, "middleware2-after")
		})
	}

	// Add route with middleware
	ts.router.Use(middleware1)
	ts.router.Use(middleware2)
	ts.router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	// Make request
	resp, err := http.Get(ts.URL + "/test")
	require.NoError(t, err)
	resp.Body.Close()

	// Verify order
	expected := []string{
		"middleware1-before",
		"middleware2-before",
		"handler",
		"middleware2-after",
		"middleware1-after",
	}
	assert.Equal(t, expected, order)
}

func TestErrorHandling(t *testing.T) {
	ts := setupTestServer(t)

	// Add error handling middleware
	ts.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "Internal server error",
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	})

	// Add route that panics
	ts.router.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	// Add route that returns error
	ts.router.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Bad request",
		})
	})

	// Test panic recovery
	resp, err := http.Get(ts.URL + "/panic")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var errResp map[string]string
	json.NewDecoder(resp.Body).Decode(&errResp)
	assert.Equal(t, "Internal server error", errResp["error"])

	// Test normal error
	resp, err = http.Get(ts.URL + "/error")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	json.NewDecoder(resp.Body).Decode(&errResp)
	assert.Equal(t, "Bad request", errResp["error"])
}
