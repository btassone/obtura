package admin

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/btassone/obtura/internal/models"
	"github.com/btassone/obtura/pkg/plugin"
	authPlugin "github.com/btassone/obtura/plugins/auth"
	adminpages "github.com/btassone/obtura/web/templates/admin/pages"
	"github.com/go-chi/chi/v5"
)

// SetupRoutesWithPlugin configures all admin routes with plugin registry
func SetupRoutesWithPlugin(r chi.Router, registry *plugin.Registry) {
	// Middleware to ensure user context is available
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// User should already be in context from RequireAdmin middleware
			next.ServeHTTP(w, req)
		})
	})

	// Admin dashboard
	r.Get("/", handleAdminDashboard)
	
	// Pages management
	r.Route("/pages", func(r chi.Router) {
		r.Get("/", handlePagesIndex)
		r.Get("/new", handleNewPage)
		r.Post("/", handleCreatePage)
		r.Get("/{id}", handleEditPage)
		r.Put("/{id}", handleUpdatePage)
		r.Delete("/{id}", handleDeletePage)
	})
	
	// Themes management
	r.Route("/themes", func(r chi.Router) {
		r.Get("/", handleThemesIndex)
		r.Post("/activate/{id}", handleActivateTheme)
	})
	
	// Plugins management
	r.Route("/plugins", func(r chi.Router) {
		r.Get("/", handlePluginsListWithRegistry(registry))
		r.Post("/{id}/toggle", handlePluginToggleWithRegistry(registry))
		r.Get("/{id}/config", handlePluginConfigWithRegistry(registry))
		r.Post("/{id}/config", handlePluginConfigUpdateWithRegistry(registry))
	})
	
	// Users management
	r.Route("/users", func(r chi.Router) {
		r.Get("/", handleUsersIndex)
		r.Get("/new", handleNewUser)
		r.Post("/", handleCreateUser)
		r.Get("/{id}", handleEditUser)
		r.Put("/{id}", handleUpdateUser)
		r.Delete("/{id}", handleDeleteUser)
	})
	
	// Settings
	r.Route("/settings", func(r chi.Router) {
		r.Get("/", handleSettings)
		r.Post("/", handleUpdateSettings)
	})
	
	// Media management
	r.Route("/media", func(r chi.Router) {
		r.Get("/upload", handleMediaUpload)
		r.Post("/upload", handleProcessUpload)
	})
}

// getUser helper extracts user from context
func getUser(r *http.Request) *models.User {
	// Try to get auth user from context
	if authUser, ok := r.Context().Value(authPlugin.UserContextKey).(plugin.AuthUser); ok {
		// Convert to models.User for templates
		// This is a temporary solution until we refactor templates to use AuthUser interface
		return &models.User{
			Name:  authUser.Name(),
			Email: authUser.Email(),
			Role:  authUser.Role(),
		}
	}
	return nil
}

// Dashboard handler
func handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	component := adminpages.AdminDashboard(user)
	templ.Handler(component).ServeHTTP(w, r)
}

// Pages handlers
func handlePagesIndex(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	component := adminpages.PagesList(user)
	templ.Handler(component).ServeHTTP(w, r)
}

func handleNewPage(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	component := adminpages.NewPage(user)
	templ.Handler(component).ServeHTTP(w, r)
}

func handleCreatePage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create page - TODO"))
}

func handleEditPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Edit page - TODO"))
}

func handleUpdatePage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Update page - TODO"))
}

func handleDeletePage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Delete page - TODO"))
}

func handleThemesIndex(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	component := adminpages.ThemesList(user)
	templ.Handler(component).ServeHTTP(w, r)
}

func handleActivateTheme(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Activate theme - TODO"))
}

func handlePluginsIndex(w http.ResponseWriter, r *http.Request) {
	// This will be replaced in SetupRoutesWithPlugin
	user := getUser(r)
	component := adminpages.PluginsList(user)
	templ.Handler(component).ServeHTTP(w, r)
}

func handleTogglePlugin(w http.ResponseWriter, r *http.Request) {
	// This will be replaced in SetupRoutesWithPlugin
	w.Write([]byte("Toggle plugin - TODO"))
}

func handleUsersIndex(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	component := adminpages.UsersList(user)
	templ.Handler(component).ServeHTTP(w, r)
}

func handleNewUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("New user form - TODO"))
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Create user - TODO"))
}

func handleEditUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Edit user - TODO"))
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Update user - TODO"))
}

func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Delete user - TODO"))
}

func handleSettings(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	component := adminpages.Settings(user)
	templ.Handler(component).ServeHTTP(w, r)
}

func handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Update settings - TODO"))
}

// Media handlers
func handleMediaUpload(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	component := adminpages.MediaUpload(user)
	templ.Handler(component).ServeHTTP(w, r)
}

func handleProcessUpload(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Process upload - TODO"))
}