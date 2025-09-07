package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestServer represents a test HTTP server
type TestServer struct {
	*httptest.Server
	Client *http.Client
}

// NewTestServer creates a new test server with the given handler
func NewTestServer(t *testing.T, handler http.Handler) *TestServer {
	srv := httptest.NewServer(handler)

	t.Cleanup(func() {
		srv.Close()
	})

	return &TestServer{
		Server: srv,
		Client: srv.Client(),
	}
}

// NewTLSTestServer creates a new TLS test server
func NewTLSTestServer(t *testing.T, handler http.Handler) *TestServer {
	srv := httptest.NewTLSServer(handler)

	t.Cleanup(func() {
		srv.Close()
	})

	return &TestServer{
		Server: srv,
		Client: srv.Client(),
	}
}

// Request makes a request to the test server
func (ts *TestServer) Request(t *testing.T, method, path string, body interface{}) *http.Response {
	req := HTTPRequest(t, method, ts.URL+path, body)
	resp, err := ts.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	return resp
}

// AuthRequest makes an authenticated request to the test server
func (ts *TestServer) AuthRequest(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	req := HTTPRequest(t, method, ts.URL+path, body)
	AuthRequest(req, token)
	resp, err := ts.Client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	return resp
}

// TestRouter is a simple router for testing
type TestRouter struct {
	routes map[string]map[string]http.HandlerFunc
}

// NewTestRouter creates a new test router
func NewTestRouter() *TestRouter {
	return &TestRouter{
		routes: make(map[string]map[string]http.HandlerFunc),
	}
}

// Handle registers a handler for a method and path
func (r *TestRouter) Handle(method, path string, handler http.HandlerFunc) {
	if r.routes[path] == nil {
		r.routes[path] = make(map[string]http.HandlerFunc)
	}
	r.routes[path][method] = handler
}

// ServeHTTP implements http.Handler
func (r *TestRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if methods, ok := r.routes[req.URL.Path]; ok {
		if handler, ok := methods[req.Method]; ok {
			handler(w, req)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

// MockMiddleware creates a simple test middleware
func MockMiddleware(name string, before, after func(*http.Request)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if before != nil {
				before(r)
			}
			next.ServeHTTP(w, r)
			if after != nil {
				after(r)
			}
		})
	}
}

// ChainMiddleware chains multiple middleware together
func ChainMiddleware(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
