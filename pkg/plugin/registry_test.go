package plugin

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	config       interface{}
}

func (p *TestPlugin) ID() string          { return p.id }
func (p *TestPlugin) Name() string        { return "Test Plugin" }
func (p *TestPlugin) Version() string     { return "1.0.0" }
func (p *TestPlugin) Description() string { return "Test plugin" }
func (p *TestPlugin) Author() string      { return "Test Author" }

func (p *TestPlugin) Init(ctx context.Context) error    { return p.initError }
func (p *TestPlugin) Start(ctx context.Context) error   { return p.startError }
func (p *TestPlugin) Stop(ctx context.Context) error    { return p.stopError }
func (p *TestPlugin) Destroy(ctx context.Context) error { return p.destroyError }

func (p *TestPlugin) Dependencies() []string     { return p.dependencies }
func (p *TestPlugin) Config() interface{}        { return p.config }
func (p *TestPlugin) ValidateConfig() error      { return nil }
func (p *TestPlugin) DefaultConfig() interface{} { return struct{}{} }

func TestRegistry_Register(t *testing.T) {
	tests := []struct {
		name    string
		plugins []*TestPlugin
		wantErr bool
		errMsg  string
	}{
		{
			name: "register single plugin",
			plugins: []*TestPlugin{
				{id: "test.plugin.1"},
			},
			wantErr: false,
		},
		{
			name: "register multiple plugins",
			plugins: []*TestPlugin{
				{id: "test.plugin.1"},
				{id: "test.plugin.2"},
			},
			wantErr: false,
		},
		{
			name: "register duplicate plugin",
			plugins: []*TestPlugin{
				{id: "test.plugin.1"},
				{id: "test.plugin.1"},
			},
			wantErr: true,
			errMsg:  "plugin test.plugin.1 already registered",
		},
		{
			name: "register plugin with satisfied dependencies",
			plugins: []*TestPlugin{
				{id: "test.plugin.1"},
				{id: "test.plugin.2", dependencies: []string{"test.plugin.1"}},
			},
			wantErr: false,
		},
		{
			name: "register plugin with missing dependencies",
			plugins: []*TestPlugin{
				{id: "test.plugin.1", dependencies: []string{"test.plugin.missing"}},
			},
			wantErr: true,
			errMsg:  "plugin test.plugin.1 requires test.plugin.missing which is not registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry(chi.NewRouter())

			var lastErr error
			for _, p := range tt.plugins {
				err := registry.Register(p)
				if err != nil {
					lastErr = err
				}
			}

			if tt.wantErr {
				require.Error(t, lastErr)
				if tt.errMsg != "" {
					assert.Contains(t, lastErr.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, lastErr)
			}
		})
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry(chi.NewRouter())

	plugin := &TestPlugin{id: "test.plugin.1"}
	err := registry.Register(plugin)
	require.NoError(t, err)

	// Test getting existing plugin
	got, exists := registry.Get("test.plugin.1")
	assert.True(t, exists)
	assert.Equal(t, plugin, got)

	// Test getting non-existing plugin
	_, exists = registry.Get("test.plugin.missing")
	assert.False(t, exists)
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry(chi.NewRouter())

	// Register multiple plugins
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

	// Check all plugins are in the list
	ids := make(map[string]bool)
	for _, p := range list {
		ids[p.ID()] = true
	}

	for _, p := range plugins {
		assert.True(t, ids[p.id], "Plugin %s should be in list", p.id)
	}
}

func TestRegistry_InitAll(t *testing.T) {
	tests := []struct {
		name    string
		plugins []*TestPlugin
		wantErr bool
	}{
		{
			name: "init all plugins successfully",
			plugins: []*TestPlugin{
				{id: "test.plugin.1"},
				{id: "test.plugin.2"},
			},
			wantErr: false,
		},
		{
			name: "init fails for one plugin",
			plugins: []*TestPlugin{
				{id: "test.plugin.1"},
				{id: "test.plugin.2", initError: errors.New("init failed")},
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
			err := registry.InitAll(ctx)

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
	err := registry.InitAll(ctx)
	require.NoError(t, err)

	err = registry.StartAll(ctx)
	require.NoError(t, err)
}

func TestRegistry_StopAll(t *testing.T) {
	registry := NewRegistry(chi.NewRouter())

	// Register, init and start plugins
	plugins := []*TestPlugin{
		{id: "test.plugin.1"},
		{id: "test.plugin.2"},
	}

	for _, p := range plugins {
		err := registry.Register(p)
		require.NoError(t, err)
	}

	ctx := context.Background()
	err := registry.InitAll(ctx)
	require.NoError(t, err)

	err = registry.StartAll(ctx)
	require.NoError(t, err)

	// Stop all plugins
	err = registry.StopAll(ctx)
	require.NoError(t, err)
}

// TestServicePlugin tests service plugin functionality
type TestServicePlugin struct {
	TestPlugin
	service interface{}
}

func (p *TestServicePlugin) Service() interface{} {
	return p.service
}

func TestRegistry_ServicePlugin(t *testing.T) {
	registry := NewRegistry(chi.NewRouter())

	// Create a test service
	type TestService struct {
		Value string
	}
	service := &TestService{Value: "test"}

	plugin := &TestServicePlugin{
		TestPlugin: TestPlugin{id: "test.service"},
		service:    service,
	}

	err := registry.Register(plugin)
	require.NoError(t, err)

	// Get the service
	got := registry.GetService("test.service")
	assert.NotNil(t, got)
	assert.Equal(t, service, got)

	// Get non-existing service
	got = registry.GetService("non.existing")
	assert.Nil(t, got)
}

// TestRoutablePlugin tests routable plugin functionality
type TestRoutablePlugin struct {
	TestPlugin
	routes []Route
}

func (p *TestRoutablePlugin) Routes() []Route {
	return p.routes
}

func TestRegistry_RoutablePlugin(t *testing.T) {
	router := chi.NewRouter()
	registry := NewRegistry(router)
	registry.SetRouter(router)

	called := false
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}

	plugin := &TestRoutablePlugin{
		TestPlugin: TestPlugin{id: "test.routable"},
		routes: []Route{
			{
				Method:  "GET",
				Path:    "/test",
				Handler: handler,
			},
		},
	}

	err := registry.Register(plugin)
	require.NoError(t, err)

	// Test the route
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, called)
}

// TestHookablePlugin tests hookable plugin functionality
type TestHookablePlugin struct {
	TestPlugin
	hooks map[string]HookHandler
}

func (p *TestHookablePlugin) Hooks() map[string]HookHandler {
	return p.hooks
}

func TestRegistry_HookablePlugin(t *testing.T) {
	registry := NewRegistry(chi.NewRouter())

	hookCalled := false
	hookData := ""

	plugin := &TestHookablePlugin{
		TestPlugin: TestPlugin{id: "test.hookable"},
		hooks: map[string]HookHandler{
			"test.hook": func(ctx context.Context, data interface{}) (interface{}, error) {
				hookCalled = true
				hookData = data.(string)
				return "result", nil
			},
		},
	}

	err := registry.Register(plugin)
	require.NoError(t, err)

	// Execute the hook
	ctx := context.Background()
	results, err := registry.ExecuteHook(ctx, "test.hook", "test data")
	require.NoError(t, err)

	assert.True(t, hookCalled)
	assert.Equal(t, "test data", hookData)
	assert.Len(t, results, 1)
	assert.Equal(t, "result", results[0])
}

// TestAdminPlugin tests admin plugin functionality
type TestAdminPlugin struct {
	TestPlugin
	adminRoutes []Route
	navItems    []NavItem
}

func (p *TestAdminPlugin) AdminRoutes() []Route {
	return p.adminRoutes
}

func (p *TestAdminPlugin) AdminNavigation() []NavItem {
	return p.navItems
}

func TestRegistry_AdminPlugin(t *testing.T) {
	router := chi.NewRouter()
	registry := NewRegistry(router)
	registry.SetRouter(router)

	adminCalled := false
	adminHandler := func(w http.ResponseWriter, r *http.Request) {
		adminCalled = true
		w.WriteHeader(http.StatusOK)
	}

	plugin := &TestAdminPlugin{
		TestPlugin: TestPlugin{id: "test.admin"},
		adminRoutes: []Route{
			{
				Method:  "GET",
				Path:    "/settings",
				Handler: adminHandler,
			},
		},
		navItems: []NavItem{
			{
				Title: "Test Settings",
				Path:  "/admin/settings",
				Order: 1,
			},
		},
	}

	err := registry.Register(plugin)
	require.NoError(t, err)

	// Test the admin route
	req := httptest.NewRequest("GET", "/admin/settings", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, adminCalled)
}

// TestEventPlugin tests event plugin functionality
type TestEventPlugin struct {
	TestPlugin
	handlers map[string]EventHandler
}

func (p *TestEventPlugin) EventHandlers() map[string]EventHandler {
	return p.handlers
}

func TestRegistry_EventHandling(t *testing.T) {
	registry := NewRegistry(chi.NewRouter())

	// Start the event dispatcher
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go registry.DispatchEvents(ctx)

	eventReceived := make(chan bool)
	var receivedEvent Event

	plugin := &TestEventPlugin{
		TestPlugin: TestPlugin{id: "test.event"},
		handlers: map[string]EventHandler{
			"test.event.type": func(ctx context.Context, event Event) error {
				receivedEvent = event
				eventReceived <- true
				return nil
			},
		},
	}

	err := registry.Register(plugin)
	require.NoError(t, err)

	// Subscribe to events
	registry.SubscribeToEvents(ctx, plugin)

	// Emit an event
	testEvent := Event{
		Name:   "test.event.type",
		Source: "test.source",
		Data:   "test data",
	}
	registry.EmitEvent(testEvent)

	// Wait for event to be received
	select {
	case <-eventReceived:
		assert.Equal(t, testEvent.Name, receivedEvent.Name)
		assert.Equal(t, testEvent.Data, receivedEvent.Data)
	case <-time.After(1 * time.Second):
		t.Fatal("Event was not received within timeout")
	}
}

func TestRegistry_PluginLifecycle(t *testing.T) {
	registry := NewRegistry(chi.NewRouter())

	// Track lifecycle calls
	var calls []string

	plugin := &TestPlugin{
		id:           "test.lifecycle",
		initError:    nil,
		startError:   nil,
		stopError:    nil,
		destroyError: nil,
	}

	// Override lifecycle methods to track calls
	originalInit := plugin.Init
	plugin.Init = func(ctx context.Context) error {
		calls = append(calls, "init")
		return originalInit(ctx)
	}

	originalStart := plugin.Start
	plugin.Start = func(ctx context.Context) error {
		calls = append(calls, "start")
		return originalStart(ctx)
	}

	originalStop := plugin.Stop
	plugin.Stop = func(ctx context.Context) error {
		calls = append(calls, "stop")
		return originalStop(ctx)
	}

	originalDestroy := plugin.Destroy
	plugin.Destroy = func(ctx context.Context) error {
		calls = append(calls, "destroy")
		return originalDestroy(ctx)
	}

	// Register plugin
	err := registry.Register(plugin)
	require.NoError(t, err)

	ctx := context.Background()

	// Initialize
	err = registry.InitAll(ctx)
	require.NoError(t, err)

	// Start
	err = registry.StartAll(ctx)
	require.NoError(t, err)

	// Stop
	err = registry.StopAll(ctx)
	require.NoError(t, err)

	// Destroy
	err = registry.DestroyAll(ctx)
	require.NoError(t, err)

	// Verify lifecycle methods were called in correct order
	expected := []string{"init", "start", "stop", "destroy"}
	assert.Equal(t, expected, calls)
}
