package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/btassone/obtura/internal/admin"
	"github.com/btassone/obtura/internal/database"
	"github.com/btassone/obtura/pkg/plugin"
	authPlugin "github.com/btassone/obtura/plugins/auth"
	docsPlugin "github.com/btassone/obtura/plugins/docs"
	helloPlugin "github.com/btassone/obtura/plugins/hello"
	hubPlugin "github.com/btassone/obtura/plugins/hub"
	"github.com/btassone/obtura/web/templates/pages"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	port     string
	mode     string
	router   *chi.Mux
	db       *database.Manager
	registry *plugin.Registry
}

func New(port, mode string) (*Server, error) {
	// Initialize database
	dbManager, err := database.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations automatically in development
	if mode == "dev" {
		if err := dbManager.Migrate(); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	// Create router
	router := chi.NewRouter()

	s := &Server{
		port:     port,
		mode:     mode,
		router:   router,
		db:       dbManager,
	}
	
	// Setup middleware FIRST
	s.setupMiddleware()
	
	// Create plugin registry WITHOUT router (to avoid early route registration)
	registry := plugin.NewRegistry(nil)
	s.registry = registry

	// Register core plugins
	authPlug := authPlugin.NewPlugin(dbManager.DB())
	if err := registry.Register(authPlug); err != nil {
		return nil, fmt.Errorf("failed to register auth plugin: %w", err)
	}
	
	// Register documentation plugin
	docsPlug := docsPlugin.NewPlugin()
	if err := registry.Register(docsPlug); err != nil {
		return nil, fmt.Errorf("failed to register docs plugin: %w", err)
	}
	
	// Register hello plugin (example)
	helloPlug := helloPlugin.NewPlugin()
	if err := registry.Register(helloPlug); err != nil {
		return nil, fmt.Errorf("failed to register hello plugin: %w", err)
	}
	
	// Register plugin hub - must be last so it can see all other plugins
	hubPlug := hubPlugin.NewPlugin(registry)
	if err := registry.Register(hubPlug); err != nil {
		return nil, fmt.Errorf("failed to register hub plugin: %w", err)
	}

	// Initialize plugins
	ctx := context.Background()
	if err := registry.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize plugins: %w", err)
	}

	// Start plugins
	if err := registry.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start plugins: %w", err)
	}

	// Setup routes LAST
	s.setupRoutes()
	
	// Set router on registry AFTER all middleware and routes are setup
	s.registry.SetRouter(s.router)
	
	return s, nil
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	if s.mode == "dev" {
		s.router.Use(middleware.NoCache)
	}
}

func (s *Server) setupRoutes() {
	// Static files
	fileServer := http.FileServer(http.Dir("web/static"))
	s.router.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Home route
	s.router.Get("/", s.handleHome)
	
	// Plugin routes are automatically registered by the registry
	
	// Admin routes - protected with auth middleware
	s.router.Route("/admin", func(r chi.Router) {
		// Get auth plugin
		if authPlug, err := s.registry.Get("com.obtura.auth"); err == nil {
			if auth, ok := authPlug.(*authPlugin.Plugin); ok {
				// Apply auth middleware to all admin routes
				r.Use(auth.RequireAdmin())
				
				// Setup admin routes
				admin.SetupRoutesWithPlugin(r, s.registry)
			}
		}
	})
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	component := pages.HomePage()
	templ.Handler(component).ServeHTTP(w, r)
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%s", s.port)
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) Close() error {
	// Stop plugins
	ctx := context.Background()
	if s.registry != nil {
		s.registry.Stop(ctx)
	}
	
	// Close database
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
