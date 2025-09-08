package plugin

import (
	"context"
	"errors"
	"net/http"
)

// AuthProvider defines the interface for authentication providers
type AuthProvider interface {
	// Name returns the provider name (e.g., "basic", "oauth", "ldap")
	Name() string
	
	// Authenticate attempts to authenticate with credentials
	Authenticate(ctx context.Context, credentials map[string]interface{}) (AuthUser, error)
	
	// GetUser gets the current authenticated user from request
	GetUser(r *http.Request) (AuthUser, bool)
	
	// Login handles the login process
	Login(w http.ResponseWriter, r *http.Request, user AuthUser) error
	
	// Logout handles the logout process
	Logout(w http.ResponseWriter, r *http.Request) error
	
	// IsAuthenticated checks if the request is authenticated
	IsAuthenticated(r *http.Request) bool
	
	// RequireAuth returns middleware that requires authentication
	RequireAuth() func(http.Handler) http.Handler
	
	// RequireRole returns middleware that requires specific role
	RequireRole(roles ...string) func(http.Handler) http.Handler
}

// AuthUser represents an authenticated user
type AuthUser interface {
	ID() string
	Email() string
	Name() string
	Role() string
	Permissions() []string
	Metadata() map[string]interface{}
}

// AuthPlugin is the interface that auth plugins must implement
type AuthPlugin interface {
	Plugin
	RoutablePlugin
	AdminPlugin
	
	// GetProvider returns an auth provider by name
	GetProvider(name string) (AuthProvider, bool)
	
	// GetActiveProvider returns the currently active provider
	GetActiveProvider() AuthProvider
	
	// SetActiveProvider sets the active provider
	SetActiveProvider(name string) error
	
	// RegisterProvider registers a new auth provider
	RegisterProvider(provider AuthProvider) error
	
	// Providers returns all registered providers
	Providers() map[string]AuthProvider
}

// AuthConfig represents auth plugin configuration
type AuthConfig struct {
	ActiveProvider string                 `json:"active_provider"`
	SessionSecret  string                 `json:"session_secret"`
	SessionMaxAge  int                    `json:"session_max_age"`
	Providers      map[string]interface{} `json:"providers"`
}

// NoAuthProvider implements a provider that allows all access
type NoAuthProvider struct{}

func (n *NoAuthProvider) Name() string { return "none" }

func (n *NoAuthProvider) Authenticate(ctx context.Context, credentials map[string]interface{}) (AuthUser, error) {
	return &GuestUser{}, nil
}

func (n *NoAuthProvider) GetUser(r *http.Request) (AuthUser, bool) {
	return &GuestUser{}, true
}

func (n *NoAuthProvider) Login(w http.ResponseWriter, r *http.Request, user AuthUser) error {
	return nil
}

func (n *NoAuthProvider) Logout(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (n *NoAuthProvider) IsAuthenticated(r *http.Request) bool {
	return true
}

func (n *NoAuthProvider) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next
	}
}

func (n *NoAuthProvider) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next
	}
}

// GuestUser represents an unauthenticated user
type GuestUser struct{}

func (g *GuestUser) ID() string                        { return "guest" }
func (g *GuestUser) Email() string                     { return "guest@example.com" }
func (g *GuestUser) Name() string                      { return "Guest" }
func (g *GuestUser) Role() string                      { return "guest" }
func (g *GuestUser) Permissions() []string             { return []string{} }
func (g *GuestUser) Metadata() map[string]interface{}  { return map[string]interface{}{} }

// Common errors for auth
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUnauthorized       = errors.New("unauthorized")
)

// User represents a concrete user type for tests and basic usage
type User struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// BasicCredentials represents username/password authentication
type BasicCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Context key for auth
type contextKey string

const userContextKey contextKey = "user"

// SetUserInContext adds a user to the context
func SetUserInContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// GetUserFromContext retrieves a user from the context
func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey).(*User)
	return user, ok
}