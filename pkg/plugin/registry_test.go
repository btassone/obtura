package plugin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPlugin is a basic plugin implementation for testing
type TestPlugin struct {
	id           string
	dependencies []string
	initError    error
	startError   error
	stopError    error
	destroyError error
	initialized  bool
	started      bool
	stopped      bool
	destroyed    bool
	service      interface{}
	routes       []Route
	adminRoutes  []Route
}

func (p *TestPlugin) ID() string                       { return p.id }
func (p *TestPlugin) Name() string                     { return p.id }
func (p *TestPlugin) Version() string                  { return "1.0.0" }
func (p *TestPlugin) Description() string              { return "Test plugin" }
func (p *TestPlugin) Author() string                   { return "Test Author" }
func (p *TestPlugin) Website() string                  { return "https://example.com" }
func (p *TestPlugin) Dependencies() []string           { return p.dependencies }
func (p *TestPlugin) Init(ctx context.Context) error   { p.initialized = true; return p.initError }
func (p *TestPlugin) Start(ctx context.Context) error  { p.started = true; return p.startError }
func (p *TestPlugin) Stop(ctx context.Context) error   { p.stopped = true; return p.stopError }
func (p *TestPlugin) Destroy(ctx context.Context) error { p.destroyed = true; return p.destroyError }
func (p *TestPlugin) Config() interface{}              { return struct{}{} }
func (p *TestPlugin) DefaultConfig() interface{}       { return struct{}{} }
func (p *TestPlugin) ValidateConfig() error { return nil }

// TestServicePlugin is a service plugin implementation for testing
type TestServicePlugin struct {
	TestPlugin
	service interface{}
}

func (p *TestServicePlugin) Service() interface{} {
	return p.service
}

// TestRoutablePlugin tests routable plugin functionality
type TestRoutablePlugin struct {
	TestPlugin
	routes      []Route
	adminRoutes []Route
}

func (p *TestRoutablePlugin) Routes() []Route {
	return p.routes
}

func (p *TestRoutablePlugin) AdminRoutes() []Route {
	return p.adminRoutes
}

// TestAdminPlugin tests admin plugin functionality
type TestAdminPlugin struct {
	TestPlugin
	adminNav []NavItem
}

func (p *TestAdminPlugin) AdminNavigation() []NavItem {
	return p.adminNav
}

// TestEventPlugin tests event plugin functionality
type TestEventPlugin struct {
	TestPlugin
	handlers map[string]EventHandler
}

func (p *TestEventPlugin) EventHandlers() map[string]EventHandler {
	return p.handlers
}

// TestHookablePlugin tests hookable plugin functionality
type TestHookablePlugin struct {
	TestPlugin
	hooks map[string]HookHandler
}

func (p *TestHookablePlugin) Hooks() map[string]HookHandler {
	return p.hooks
}

func TestNewRegistry(t *testing.T) {
	router := chi.NewRouter()
	registry := NewRegistry(router)

	assert.NotNil(t, registry)
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry(chi.NewRouter())

	plugin := &TestPlugin{
		id: "test.plugin.1",
	}

	err := registry.Register(plugin)
	require.NoError(t, err)

	// Test duplicate registration
	err = registry.Register(plugin)
	assert.Error(t, err)
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry(chi.NewRouter())

	plugin := &TestPlugin{
		id: "test.plugin.1",
	}

	err := registry.Register(plugin)
	require.NoError(t, err)

	// Test getting existing plugin
	got, err := registry.Get("test.plugin.1")
	assert.NoError(t, err)
	assert.Equal(t, plugin, got)

	// Test getting non-existing plugin
	_, err = registry.Get("test.plugin.missing")
	assert.Error(t, err)
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry(chi.NewRouter())

	plugins := []*TestPlugin{
		{id: "test.plugin.1"},
		{id: "test.plugin.2"},
		{id: "test.plugin.3"},
	}

	for _, p := range plugins {
		err := registry.Register(p)
		require.NoError(t, err)
	}

	list := registry.List()
	assert.Len(t, list, 3)
}

func TestRegistry_Dependencies(t *testing.T) {
	tests := []struct {
		name    string
		plugins []*TestPlugin
		wantErr bool
	}{
		{
			name: "no dependencies",
			plugins: []*TestPlugin{
				{id: "test.plugin.1"},
			},
			wantErr: false,
		},
		{
			name: "valid dependencies",
			plugins: []*TestPlugin{
				{id: "test.plugin.1"},
				{id: "test.plugin.2", dependencies: []string{"test.plugin.1"}},
			},
			wantErr: false,
		},
		{
			name: "missing dependency",
			plugins: []*TestPlugin{
				{id: "test.plugin.1", dependencies: []string{"test.missing"}},
			},
			wantErr: true,
		},
		{
			name: "circular dependency",
			plugins: []*TestPlugin{
				{id: "test.plugin.1", dependencies: []string{"test.plugin.2"}},
				{id: "test.plugin.2", dependencies: []string{"test.plugin.1"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry(chi.NewRouter())

			for _, p := range tt.plugins {
				err := registry.Register(p)
				require.NoError(t, err)
			}

			ctx := context.Background()
			err := registry.Initialize(ctx)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegistry_StartAll(t *testing.T) {
	registry := NewRegistry(chi.NewRouter())

	// Register and init plugins
	plugins := []*TestPlugin{
		{id: "test.plugin.1"},
		{id: "test.plugin.2"},
	}

	for _, p := range plugins {
		err := registry.Register(p)
		require.NoError(t, err)
	}

	ctx := context.Background()
	err := registry.Initialize(ctx)
	require.NoError(t, err)

	err = registry.Start(ctx)
	require.NoError(t, err)
}

func TestRegistry_StopAll(t *testing.T) {
	registry := NewRegistry(chi.NewRouter())

	// Register, init, and start plugins
	plugins := []*TestPlugin{
		{id: "test.plugin.1"},
		{id: "test.plugin.2"},
	}

	for _, p := range plugins {
		err := registry.Register(p)
		require.NoError(t, err)
	}

	ctx := context.Background()
	err := registry.Initialize(ctx)
	require.NoError(t, err)

	err = registry.Start(ctx)
	require.NoError(t, err)

	// Stop all plugins
	err = registry.Stop(ctx)
	require.NoError(t, err)
}

func TestRegistry_ServicePlugin(t *testing.T) {
	registry := NewRegistry(chi.NewRouter())

	service := "test service"
	plugin := &TestServicePlugin{
		TestPlugin: TestPlugin{id: "test.service"},
		service:    service,
	}

	err := registry.Register(plugin)
	require.NoError(t, err)

	ctx := context.Background()
	err = registry.Initialize(ctx)
	require.NoError(t, err)

	// Get the service
	got, ok := registry.GetService("test.service")
	assert.True(t, ok)
	assert.NotNil(t, got)
	assert.Equal(t, service, got)

	// Get non-existing service
	got, ok = registry.GetService("non.existing")
	assert.False(t, ok)
	assert.Nil(t, got)
}

// TestRoutablePlugin tests routable plugin functionality
func TestRegistry_RoutablePlugin(t *testing.T) {
	// Create router first
	router := chi.NewRouter()
	registry := NewRegistry(router)

	routes := []Route{
		{
			Method: http.MethodGet,
			Path:   "/test",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		},
	}

	plugin := &TestRoutablePlugin{
		TestPlugin: TestPlugin{id: "test.routable"},
		routes:     routes,
	}

	err := registry.Register(plugin)
	require.NoError(t, err)

	// Test the route
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// TODO: Implement ExecuteHook in registry
// func TestRegistry_HookablePlugin(t *testing.T) { ... }

// TODO: Implement event handling methods in registry
// func TestRegistry_EventHandling(t *testing.T) { ... }

// TODO: Fix plugin lifecycle test - cannot reassign interface methods
// func TestRegistry_PluginLifecycle(t *testing.T) { ... }