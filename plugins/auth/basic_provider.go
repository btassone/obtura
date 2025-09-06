package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/btassone/obtura/internal/models"
	"github.com/btassone/obtura/pkg/database"
	"github.com/btassone/obtura/pkg/plugin"
	authTemplates "github.com/btassone/obtura/web/templates/auth"
	"github.com/gorilla/sessions"
)

type contextKey string

const (
	SessionName = "obtura_session"
	UserIDKey   = "user_id"
)

var UserContextKey = contextKey("user")

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user account is inactive")
)

// BasicAuthProvider implements basic email/password authentication
type BasicAuthProvider struct {
	db       *database.DB
	userRepo *models.UserRepository
	store    *sessions.CookieStore
	config   *plugin.AuthConfig
}

// BasicAuthUser wraps a model user to implement AuthUser
type BasicAuthUser struct {
	user *models.User
}

func (u *BasicAuthUser) ID() string    { return fmt.Sprintf("%d", u.user.ID) }
func (u *BasicAuthUser) Email() string { return u.user.Email }
func (u *BasicAuthUser) Name() string  { return u.user.Name }
func (u *BasicAuthUser) Role() string  { return u.user.Role }
func (u *BasicAuthUser) Permissions() []string {
	// TODO: Implement permissions
	if u.user.Role == "admin" {
		return []string{"*"}
	}
	return []string{}
}
func (u *BasicAuthUser) Metadata() map[string]interface{} {
	return map[string]interface{}{
		"avatar": u.user.Avatar,
		"bio":    u.user.Bio,
		"active": u.user.Active,
	}
}

// Provider implementation

func (b *BasicAuthProvider) Name() string { return "basic" }

func (b *BasicAuthProvider) Authenticate(ctx context.Context, credentials map[string]interface{}) (plugin.AuthUser, error) {
	email, _ := credentials["email"].(string)
	password, _ := credentials["password"].(string)

	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	// Find user by email
	user, err := b.userRepo.FindByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.Active {
		return nil, ErrUserInactive
	}

	// Check password
	if !user.CheckPassword(password) {
		return nil, ErrInvalidCredentials
	}

	return &BasicAuthUser{user: user}, nil
}

func (b *BasicAuthProvider) GetUser(r *http.Request) (plugin.AuthUser, bool) {
	// Check context first
	if user, ok := r.Context().Value(UserContextKey).(*BasicAuthUser); ok {
		return user, true
	}

	// Get from session
	session, _ := b.store.Get(r, SessionName)
	userID, ok := session.Values[UserIDKey].(int64)
	if !ok || userID == 0 {
		return nil, false
	}

	// Load user from database
	user, err := b.userRepo.FindByID(userID)
	if err != nil {
		return nil, false
	}

	return &BasicAuthUser{user: user}, true
}

func (b *BasicAuthProvider) Login(w http.ResponseWriter, r *http.Request, user plugin.AuthUser) error {
	// Create session
	session, _ := b.store.Get(r, SessionName)

	// Convert user ID to int64
	var userID int64
	fmt.Sscanf(user.ID(), "%d", &userID)

	session.Values[UserIDKey] = userID
	return session.Save(r, w)
}

func (b *BasicAuthProvider) Logout(w http.ResponseWriter, r *http.Request) error {
	session, _ := b.store.Get(r, SessionName)

	// Delete session values
	session.Values[UserIDKey] = nil
	session.Options.MaxAge = -1

	return session.Save(r, w)
}

func (b *BasicAuthProvider) IsAuthenticated(r *http.Request) bool {
	session, _ := b.store.Get(r, SessionName)
	userID, ok := session.Values[UserIDKey].(int64)
	return ok && userID > 0
}

func (b *BasicAuthProvider) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if user is authenticated
			if !b.IsAuthenticated(r) {
				// Redirect to login with return URL
				returnURL := r.URL.Path
				if r.URL.RawQuery != "" {
					returnURL += "?" + r.URL.RawQuery
				}

				http.Redirect(w, r, "/login?return="+returnURL, http.StatusSeeOther)
				return
			}

			// Get user and add to context
			user, ok := b.GetUser(r)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (b *BasicAuthProvider) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// First require auth
		return b.RequireAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContextKey).(plugin.AuthUser)
			if !ok {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Check if user has one of the required roles
			userRole := user.Role()
			for _, role := range roles {
				if userRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Forbidden", http.StatusForbidden)
		}))
	}
}

// Additional methods for basic auth

func (b *BasicAuthProvider) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	returnURL := r.URL.Query().Get("return")
	if returnURL == "" {
		returnURL = "/admin"
	}

	component := authTemplates.LoginPage("", returnURL)
	templ.Handler(component).ServeHTTP(w, r)
}

func (b *BasicAuthProvider) HandleLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	returnURL := r.FormValue("return")

	if returnURL == "" {
		returnURL = "/admin"
	}

	// Attempt authentication
	user, err := b.Authenticate(r.Context(), map[string]interface{}{
		"email":    email,
		"password": password,
	})

	if err != nil {
		// Show error
		var errorMsg string
		switch err {
		case ErrInvalidCredentials:
			errorMsg = "Invalid email or password"
		case ErrUserInactive:
			errorMsg = "Your account has been deactivated"
		default:
			errorMsg = "An error occurred during login"
		}

		component := authTemplates.LoginPage(errorMsg, returnURL)
		w.WriteHeader(http.StatusUnauthorized)
		templ.Handler(component).ServeHTTP(w, r)
		return
	}

	// Log the user in
	if err := b.Login(w, r, user); err != nil {
		component := authTemplates.LoginPage("Failed to create session", returnURL)
		w.WriteHeader(http.StatusInternalServerError)
		templ.Handler(component).ServeHTTP(w, r)
		return
	}

	// Redirect to return URL
	http.Redirect(w, r, returnURL, http.StatusSeeOther)
}

func (b *BasicAuthProvider) CreateInitialAdmin() error {
	// Check if any admin exists
	user, _ := b.userRepo.FindByEmail("admin@example.com")
	if user != nil {
		return nil // Admin already exists
	}

	// Create admin user
	admin := &models.User{
		Name:            "Admin User",
		Email:           "admin@example.com",
		Password:        "admin123", // This will be hashed by Create
		Role:            "admin",
		Active:          true,
		EmailVerifiedAt: &[]time.Time{time.Now()}[0],
	}

	return b.userRepo.Create(admin)
}

// NewBasicAuthProvider Initialize the provider with config
func NewBasicAuthProvider(db *database.DB, userRepo *models.UserRepository, config *plugin.AuthConfig) *BasicAuthProvider {
	store := sessions.NewCookieStore([]byte(config.SessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   config.SessionMaxAge,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	return &BasicAuthProvider{
		db:       db,
		userRepo: userRepo,
		store:    store,
		config:   config,
	}
}
