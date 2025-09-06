package plugin

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
)

// Registry manages all plugins
type Registry struct {
	mu       sync.RWMutex
	plugins  map[string]Plugin
	services map[string]interface{}
	hooks    map[string][]HookHandler
	events   chan Event
	router   *chi.Mux
	routes   []Route // Store routes until router is set
	
	// Plugin states
	initialized map[string]bool
	started     map[string]bool
	
	// Configuration
	configManager *ConfigManager
}

// NewRegistry creates a new plugin registry
func NewRegistry(router *chi.Mux) *Registry {
	// Create config manager with file storage
	configStorage, err := NewJSONFileConfigStorage("./configs/plugins")
	if err != nil {
		// Fall back to memory storage
		configStorage = NewMemoryConfigStorage()
	}
	
	return &Registry{
		plugins:       make(map[string]Plugin),
		services:      make(map[string]interface{}),
		hooks:         make(map[string][]HookHandler),
		events:        make(chan Event, 100),
		router:        router,
		routes:        make([]Route, 0),
		initialized:   make(map[string]bool),
		started:       make(map[string]bool),
		configManager: NewConfigManagerWithStorage(configStorage),
	}
}

// Register adds a plugin to the registry
func (r *Registry) Register(p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	id := p.ID()
	if _, exists := r.plugins[id]; exists {
		return fmt.Errorf("plugin %s already registered", id)
	}
	
	// Check dependencies
	for _, dep := range p.Dependencies() {
		if _, exists := r.plugins[dep]; !exists {
			return fmt.Errorf("plugin %s requires %s which is not registered", id, dep)
		}
	}
	
	r.plugins[id] = p
	
	// Register default config and schema
	r.configManager.SetConfig(id, p.DefaultConfig())
	if schema := GenerateSchemaFromStruct(p.Config()); schema != nil {
		r.configManager.RegisterSchema(id, schema)
	}
	
	// Register services if this is a service plugin
	if sp, ok := p.(ServicePlugin); ok {
		r.services[id] = sp.Service()
	}
	
	// Register hooks if this is a hookable plugin
	if hp, ok := p.(HookablePlugin); ok {
		for hookName, handler := range hp.Hooks() {
			r.hooks[hookName] = append(r.hooks[hookName], handler)
		}
	}
	
	// Register routes if this is a routable plugin
	if rp, ok := p.(RoutablePlugin); ok {
		for _, route := range rp.Routes() {
			r.registerRoute(route)
		}
	}
	
	// Register admin routes if this is an admin plugin
	if ap, ok := p.(AdminPlugin); ok {
		for _, route := range ap.AdminRoutes() {
			// Prefix admin routes
			route.Path = "/admin" + route.Path
			r.registerRoute(route)
		}
	}
	
	return nil
}

// registerRoute registers a single route
func (r *Registry) registerRoute(route Route) {
	// If router is not set, store routes for later
	if r.router == nil {
		r.routes = append(r.routes, route)
		return
	}
	
	var handler http.Handler = route.Handler
	
	// Apply route middlewares in reverse order
	for i := len(route.Middlewares) - 1; i >= 0; i-- {
		handler = route.Middlewares[i](handler)
	}
	
	// Convert back to HandlerFunc for chi
	handlerFunc := func(w http.ResponseWriter, req *http.Request) {
		handler.ServeHTTP(w, req)
	}
	
	switch route.Method {
	case http.MethodGet:
		r.router.Get(route.Path, handlerFunc)
	case http.MethodPost:
		r.router.Post(route.Path, handlerFunc)
	case http.MethodPut:
		r.router.Put(route.Path, handlerFunc)
	case http.MethodDelete:
		r.router.Delete(route.Path, handlerFunc)
	case http.MethodPatch:
		r.router.Patch(route.Path, handlerFunc)
	default:
		r.router.HandleFunc(route.Path, handlerFunc)
	}
}

// SetRouter sets the router and registers all stored routes
func (r *Registry) SetRouter(router *chi.Mux) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.router = router
	
	// Register all stored routes
	for _, route := range r.routes {
		r.registerRoute(route)
	}
	
	// Clear stored routes
	r.routes = nil
}


// GetService returns a service by plugin ID
func (r *Registry) GetService(id string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.services[id]
	return s, ok
}

// IsEnabled checks if a plugin is enabled/active
func (r *Registry) IsEnabled(pluginID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// A plugin is considered enabled if it's started
	return r.started[pluginID]
}

// Get returns a plugin by ID
func (r *Registry) Get(pluginID string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	p, ok := r.plugins[pluginID]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}
	return p, nil
}

// List returns all registered plugins
func (r *Registry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	list := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		list = append(list, p)
	}
	return list
}

// Initialize initializes all plugins
func (r *Registry) Initialize(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Initialize in dependency order
	for _, p := range r.plugins {
		if err := r.initializePlugin(ctx, p); err != nil {
			return err
		}
	}
	
	return nil
}

// initializePlugin initializes a single plugin and its dependencies
func (r *Registry) initializePlugin(ctx context.Context, p Plugin) error {
	id := p.ID()
	
	// Already initialized
	if r.initialized[id] {
		return nil
	}
	
	// Initialize dependencies first
	for _, depID := range p.Dependencies() {
		if dep, ok := r.plugins[depID]; ok {
			if err := r.initializePlugin(ctx, dep); err != nil {
				return err
			}
		}
	}
	
	// Initialize this plugin
	if err := p.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", id, err)
	}
	
	r.initialized[id] = true
	return nil
}

// Start starts all plugins
func (r *Registry) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for _, p := range r.plugins {
		if err := r.startPlugin(ctx, p); err != nil {
			return err
		}
	}
	
	// Start event processor
	go r.processEvents(ctx)
	
	return nil
}

// startPlugin starts a single plugin
func (r *Registry) startPlugin(ctx context.Context, p Plugin) error {
	id := p.ID()
	
	// Already started
	if r.started[id] {
		return nil
	}
	
	// Start dependencies first
	for _, depID := range p.Dependencies() {
		if dep, ok := r.plugins[depID]; ok {
			if err := r.startPlugin(ctx, dep); err != nil {
				return err
			}
		}
	}
	
	// Start this plugin
	if err := p.Start(ctx); err != nil {
		return fmt.Errorf("failed to start plugin %s: %w", id, err)
	}
	
	r.started[id] = true
	return nil
}

// Stop stops all plugins
func (r *Registry) Stop(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Stop in reverse order
	var errs []error
	for _, p := range r.plugins {
		if r.started[p.ID()] {
			if err := p.Stop(ctx); err != nil {
				errs = append(errs, fmt.Errorf("failed to stop plugin %s: %w", p.ID(), err))
			}
			r.started[p.ID()] = false
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("errors stopping plugins: %v", errs)
	}
	
	return nil
}

// ExecuteHook executes all handlers for a hook
func (r *Registry) ExecuteHook(ctx context.Context, hookName string, data interface{}) (interface{}, error) {
	r.mu.RLock()
	handlers := r.hooks[hookName]
	r.mu.RUnlock()
	
	result := data
	for _, handler := range handlers {
		var err error
		result, err = handler(ctx, result)
		if err != nil {
			return nil, err
		}
	}
	
	return result, nil
}

// EmitEvent emits an event
func (r *Registry) EmitEvent(event Event) {
	select {
	case r.events <- event:
	default:
		// Event queue full, drop event
		// In production, you might want to log this
	}
}

// processEvents processes events in the background
func (r *Registry) processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-r.events:
			r.handleEvent(event)
		}
	}
}

// handleEvent handles a single event
func (r *Registry) handleEvent(event Event) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Notify all event plugins
	for _, p := range r.plugins {
		if ep, ok := p.(EventPlugin); ok {
			if handler, exists := ep.EventHandlers()[event.Name]; exists {
				go handler(event.Context, event)
			}
		}
	}
}

// GetConfig gets plugin configuration
func (r *Registry) GetConfig(pluginID string) (interface{}, bool) {
	return r.configManager.GetConfig(pluginID)
}

// SetConfig sets plugin configuration
func (r *Registry) SetConfig(pluginID string, config interface{}) error {
	p, ok := r.plugins[pluginID]
	if !ok {
		return fmt.Errorf("plugin %s not found", pluginID)
	}
	
	// Update plugin's config
	if err := r.configManager.SetConfig(pluginID, config); err != nil {
		return err
	}
	
	// Load config into plugin
	if err := r.configManager.LoadConfig(pluginID, p.Config()); err != nil {
		return err
	}
	
	// Validate configuration
	if err := p.ValidateConfig(); err != nil {
		return err
	}
	
	return nil
}

// GetConfigManager returns the config manager
func (r *Registry) GetConfigManager() *ConfigManager {
	return r.configManager
}