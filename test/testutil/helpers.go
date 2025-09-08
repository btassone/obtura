package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestPlugin is a minimal plugin implementation for testing
type TestPlugin struct {
	IDFunc             func() string
	NameFunc           func() string
	DescriptionFunc    func() string
	VersionFunc        func() string
	AuthorFunc         func() string
	WebsiteFunc        func() string
	DependenciesFunc   func() []string
	SettingsFieldsFunc func() []map[string]interface{}
	ConfigFunc         func() interface{}
	ValidateConfigFunc func() error
	InitFunc           func(context.Context) error
	StartFunc          func(context.Context) error
	StopFunc           func(context.Context) error
	DestroyFunc        func(context.Context) error
	RegisterFunc       func(router http.Handler) http.Handler
	DefaultConfigFunc  func() interface{}
}

func (p *TestPlugin) ID() string {
	if p.IDFunc != nil {
		return p.IDFunc()
	}
	return "test-plugin"
}

func (p *TestPlugin) Name() string {
	if p.NameFunc != nil {
		return p.NameFunc()
	}
	return "test-plugin"
}

func (p *TestPlugin) Description() string {
	if p.DescriptionFunc != nil {
		return p.DescriptionFunc()
	}
	return "Test plugin"
}

func (p *TestPlugin) Version() string {
	if p.VersionFunc != nil {
		return p.VersionFunc()
	}
	return "1.0.0"
}

func (p *TestPlugin) Author() string {
	if p.AuthorFunc != nil {
		return p.AuthorFunc()
	}
	return "Test Author"
}

func (p *TestPlugin) Website() string {
	if p.WebsiteFunc != nil {
		return p.WebsiteFunc()
	}
	return "https://example.com"
}

func (p *TestPlugin) Dependencies() []string {
	if p.DependenciesFunc != nil {
		return p.DependenciesFunc()
	}
	return []string{}
}

func (p *TestPlugin) SettingsFields() []map[string]interface{} {
	if p.SettingsFieldsFunc != nil {
		return p.SettingsFieldsFunc()
	}
	return nil
}

func (p *TestPlugin) Config() interface{} {
	if p.ConfigFunc != nil {
		return p.ConfigFunc()
	}
	return struct{}{}
}

func (p *TestPlugin) DefaultConfig() interface{} {
	if p.DefaultConfigFunc != nil {
		return p.DefaultConfigFunc()
	}
	return struct{}{}
}

func (p *TestPlugin) ValidateConfig() error {
	if p.ValidateConfigFunc != nil {
		return p.ValidateConfigFunc()
	}
	return nil
}

func (p *TestPlugin) Init(ctx context.Context) error {
	if p.InitFunc != nil {
		return p.InitFunc(ctx)
	}
	return nil
}

func (p *TestPlugin) Start(ctx context.Context) error {
	if p.StartFunc != nil {
		return p.StartFunc(ctx)
	}
	return nil
}

func (p *TestPlugin) Stop(ctx context.Context) error {
	if p.StopFunc != nil {
		return p.StopFunc(ctx)
	}
	return nil
}

func (p *TestPlugin) Destroy(ctx context.Context) error {
	if p.DestroyFunc != nil {
		return p.DestroyFunc(ctx)
	}
	return nil
}

func (p *TestPlugin) RegisterRoutes(router http.Handler) http.Handler {
	if p.RegisterFunc != nil {
		return p.RegisterFunc(router)
	}
	return router
}

// HTTPRequest creates a test HTTP request
func HTTPRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = bytes.NewBufferString(v)
		case []byte:
			bodyReader = bytes.NewBuffer(v)
		default:
			data, err := json.Marshal(body)
			require.NoError(t, err)
			bodyReader = bytes.NewBuffer(data)
		}
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil && method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

// AssertEventually asserts that a condition is eventually true
func AssertEventually(t *testing.T, condition func() bool, timeout time.Duration, tick time.Duration, msgAndArgs ...interface{}) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for {
		select {
		case <-timer.C:
			t.Fatalf("condition never satisfied: %v", msgAndArgs...)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// RequireJSONResponse checks that a response has JSON content type and unmarshals it
func RequireJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, target interface{}) {
	require.Equal(t, "application/json", resp.Header().Get("Content-Type"))
	err := json.Unmarshal(resp.Body.Bytes(), target)
	require.NoError(t, err)
}

// RequireHTMLResponse checks that a response has HTML content type
func RequireHTMLResponse(t *testing.T, resp *httptest.ResponseRecorder) {
	contentType := resp.Header().Get("Content-Type")
	require.Contains(t, contentType, "text/html")
}

// CleanupTimeout returns a context that will be cancelled when test cleanup runs
func CleanupTimeout(t *testing.T, timeout time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(func() {
		cancel()
	})
	return ctx
}
