package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/btassone/obtura/internal/models"
	"github.com/btassone/obtura/pkg/database"
	"github.com/btassone/obtura/pkg/plugin"
)

// Plugin implements the auth plugin
type Plugin struct {
	db          *database.DB
	providers   map[string]plugin.AuthProvider
	active      string
	config      *plugin.AuthConfig
	userRepo    *models.UserRepository
}

// NewPlugin creates a new auth plugin
func NewPlugin(db *database.DB) *Plugin {
	return &Plugin{
		db:        db,
		providers: make(map[string]plugin.AuthProvider),
		userRepo:  models.NewUserRepository(db),
		config: &plugin.AuthConfig{
			ActiveProvider: "basic",
			SessionSecret:  "dev-secret-key-change-in-production",
			SessionMaxAge:  86400 * 7, // 7 days
			Providers:      make(map[string]interface{}),
		},
	}
}

// Plugin interface implementation

func (p *Plugin) ID() string          { return "com.obtura.auth" }
func (p *Plugin) Name() string        { return "Authentication" }
func (p *Plugin) Version() string     { return "1.0.0" }
func (p *Plugin) Description() string { return "Provides authentication and authorization" }
func (p *Plugin) Author() string      { return "Obtura Team" }
func (p *Plugin) Dependencies() []string { return []string{} }

func (p *Plugin) Init(ctx context.Context) error {
	// Register default providers
	p.RegisterProvider(NewBasicAuthProvider(p.db, p.userRepo, p.config))
	p.RegisterProvider(&plugin.NoAuthProvider{})
	
	// Set active provider
	p.active = p.config.ActiveProvider
	
	return nil
}

func (p *Plugin) Start(ctx context.Context) error {
	// Create initial admin user if needed
	if p.active == "basic" {
		provider, _ := p.GetProvider("basic")
		if basicAuth, ok := provider.(*BasicAuthProvider); ok {
			if err := basicAuth.CreateInitialAdmin(); err != nil {
				// Log error but don't fail startup
				fmt.Printf("Note: Could not create initial admin: %v\n", err)
			}
		}
	}
	return nil
}

func (p *Plugin) Stop(ctx context.Context) error  { return nil }
func (p *Plugin) Destroy(ctx context.Context) error { return nil }

func (p *Plugin) Config() interface{}        { return p.config }
func (p *Plugin) ValidateConfig() error      { return nil }
func (p *Plugin) DefaultConfig() interface{} { return p.config }

// RoutablePlugin implementation

func (p *Plugin) Routes() []plugin.Route {
	routes := []plugin.Route{
		{
			Method:  http.MethodGet,
			Path:    "/login",
			Handler: p.handleLogin,
		},
		{
			Method:  http.MethodPost,
			Path:    "/login",
			Handler: p.handleLoginPost,
		},
		{
			Method:  http.MethodPost,
			Path:    "/logout",
			Handler: p.handleLogout,
		},
	}
	
	return routes
}

// AdminPlugin implementation

func (p *Plugin) AdminRoutes() []plugin.Route {
	return []plugin.Route{
		{
			Method:  http.MethodGet,
			Path:    "/auth",
			Handler: p.handleAdminAuth,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/provider",
			Handler: p.handleChangeProvider,
		},
	}
}

func (p *Plugin) AdminNavigation() []plugin.NavItem {
	return []plugin.NavItem{
		{
			Title: "Authentication",
			Path:  "/admin/auth",
			Icon:  "shield",
			Order: 100,
		},
	}
}

// AuthPlugin implementation

func (p *Plugin) GetProvider(name string) (plugin.AuthProvider, bool) {
	provider, ok := p.providers[name]
	return provider, ok
}

func (p *Plugin) GetActiveProvider() plugin.AuthProvider {
	if provider, ok := p.providers[p.active]; ok {
		return provider
	}
	// Fallback to no auth
	return &plugin.NoAuthProvider{}
}

func (p *Plugin) SetActiveProvider(name string) error {
	if _, ok := p.providers[name]; !ok {
		return fmt.Errorf("provider %s not found", name)
	}
	p.active = name
	p.config.ActiveProvider = name
	return nil
}

func (p *Plugin) RegisterProvider(provider plugin.AuthProvider) error {
	name := provider.Name()
	if _, exists := p.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}
	p.providers[name] = provider
	return nil
}

func (p *Plugin) Providers() map[string]plugin.AuthProvider {
	return p.providers
}

// HTTP handlers

func (p *Plugin) handleLogin(w http.ResponseWriter, r *http.Request) {
	provider := p.GetActiveProvider()
	
	// If already authenticated, redirect to admin
	if provider.IsAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	
	// For basic auth, show login page
	if p.active == "basic" {
		if basicAuth, ok := provider.(*BasicAuthProvider); ok {
			basicAuth.ShowLoginPage(w, r)
			return
		}
	}
	
	// For other providers, let them handle it
	http.Error(w, "Login not available", http.StatusNotImplemented)
}

func (p *Plugin) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	provider := p.GetActiveProvider()
	
	if p.active == "basic" {
		if basicAuth, ok := provider.(*BasicAuthProvider); ok {
			basicAuth.HandleLogin(w, r)
			return
		}
	}
	
	http.Error(w, "Login not available", http.StatusNotImplemented)
}

func (p *Plugin) handleLogout(w http.ResponseWriter, r *http.Request) {
	provider := p.GetActiveProvider()
	provider.Logout(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (p *Plugin) handleAdminAuth(w http.ResponseWriter, r *http.Request) {
	// TODO: Show auth settings page
	w.Write([]byte("Auth settings page - TODO"))
}

func (p *Plugin) handleChangeProvider(w http.ResponseWriter, r *http.Request) {
	providerName := r.FormValue("provider")
	if err := p.SetActiveProvider(providerName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	http.Redirect(w, r, "/admin/auth", http.StatusSeeOther)
}

// Middleware returns the auth middleware for the current provider
func (p *Plugin) Middleware() func(http.Handler) http.Handler {
	return p.GetActiveProvider().RequireAuth()
}

// RequireAdmin returns middleware that requires admin role
func (p *Plugin) RequireAdmin() func(http.Handler) http.Handler {
	return p.GetActiveProvider().RequireRole("admin")
}