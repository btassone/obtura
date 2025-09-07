package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/btassone/obtura/internal/models"
	"github.com/btassone/obtura/pkg/database"
	"github.com/btassone/obtura/pkg/plugin"
	"github.com/btassone/obtura/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// Mock database for testing
type mockDB struct {
	users []models.User
}

func (m *mockDB) GetGorm() interface{} {
	return testutil.TestDBWithSchema(&testing.T{}, &models.User{})
}

func TestPlugin_Metadata(t *testing.T) {
	db := &database.DB{}
	p := NewPlugin(db)

	assert.Equal(t, "com.obtura.auth", p.ID())
	assert.Equal(t, "Authentication", p.Name())
	assert.Equal(t, "1.0.0", p.Version())
	assert.Equal(t, "Provides authentication and authorization", p.Description())
	assert.Equal(t, "Obtura Team", p.Author())
	assert.Empty(t, p.Dependencies())
}

func TestPlugin_Init(t *testing.T) {
	db := &database.DB{}
	p := NewPlugin(db)

	ctx := context.Background()
	err := p.Init(ctx)
	require.NoError(t, err)

	// Check that providers are registered
	basicProvider, exists := p.GetProvider("basic")
	assert.True(t, exists)
	assert.NotNil(t, basicProvider)

	noAuthProvider, exists := p.GetProvider("none")
	assert.True(t, exists)
	assert.NotNil(t, noAuthProvider)

	// Check active provider
	assert.Equal(t, "basic", p.active)
}

func TestPlugin_RegisterProvider(t *testing.T) {
	db := &database.DB{}
	p := NewPlugin(db)

	// Create a mock provider
	mockProvider := &mockAuthProvider{
		id: "mock",
	}

	p.RegisterProvider(mockProvider)

	provider, exists := p.GetProvider("mock")
	assert.True(t, exists)
	assert.Equal(t, mockProvider, provider)
}

func TestPlugin_GetActiveProvider(t *testing.T) {
	db := &database.DB{}
	p := NewPlugin(db)

	ctx := context.Background()
	err := p.Init(ctx)
	require.NoError(t, err)

	// Test getting active provider
	provider := p.GetActiveProvider()
	assert.NotNil(t, provider)
	assert.Equal(t, "basic", provider.ID())

	// Change active provider
	p.active = "none"
	provider = p.GetActiveProvider()
	assert.NotNil(t, provider)
	assert.Equal(t, "none", provider.ID())

	// Test non-existent active provider
	p.active = "non-existent"
	provider = p.GetActiveProvider()
	assert.Nil(t, provider)
}

func TestPlugin_Middleware(t *testing.T) {
	db := &database.DB{}
	p := NewPlugin(db)

	ctx := context.Background()
	err := p.Init(ctx)
	require.NoError(t, err)

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if user is in context
		user, ok := plugin.GetUserFromContext(r.Context())
		if ok && user != nil {
			w.Header().Set("X-User-ID", string(user.ID))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply middleware
	middleware := p.Middleware()
	handler := middleware(testHandler)

	// Test request without auth
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("X-User-ID"))
}

func TestPlugin_RequireAuth(t *testing.T) {
	db := &database.DB{}
	p := NewPlugin(db)

	ctx := context.Background()
	err := p.Init(ctx)
	require.NoError(t, err)

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authenticated"))
	})

	// Apply RequireAuth middleware
	handler := p.RequireAuth(testHandler)

	// Test without authentication
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// With no-auth provider, it should redirect to login
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/login?redirect=/protected", rec.Header().Get("Location"))
}

func TestPlugin_RequireAdmin(t *testing.T) {
	db := &database.DB{}
	p := NewPlugin(db)

	ctx := context.Background()
	err := p.Init(ctx)
	require.NoError(t, err)

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Admin area"))
	})

	// Apply RequireAdmin middleware
	handler := p.RequireAdmin(testHandler)

	tests := []struct {
		name           string
		setupContext   func() context.Context
		expectedStatus int
	}{
		{
			name: "no user",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "non-admin user",
			setupContext: func() context.Context {
				user := &plugin.User{
					ID:    1,
					Email: "user@example.com",
					Role:  "user",
				}
				return plugin.SetUserInContext(context.Background(), user)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "admin user",
			setupContext: func() context.Context {
				user := &plugin.User{
					ID:    1,
					Email: "admin@example.com",
					Role:  "admin",
				}
				return plugin.SetUserInContext(context.Background(), user)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			req = req.WithContext(tt.setupContext())
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestPlugin_LoginRoutes(t *testing.T) {
	// This test would require setting up a full test database
	// For now, we'll test that routes are registered correctly
	db := &database.DB{}
	p := NewPlugin(db)

	routes := p.Routes()

	// Check that login routes exist
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Method+":"+route.Path] = true
	}

	expectedRoutes := []string{
		"GET:/login",
		"POST:/login",
		"GET:/logout",
	}

	for _, expected := range expectedRoutes {
		assert.True(t, routePaths[expected], "Expected route %s not found", expected)
	}
}

// Mock auth provider for testing
type mockAuthProvider struct {
	id           string
	authFunc     func(ctx context.Context, credentials interface{}) (*plugin.User, error)
	validateFunc func(ctx context.Context, user *plugin.User) error
}

func (m *mockAuthProvider) ID() string   { return m.id }
func (m *mockAuthProvider) Name() string { return "Mock Provider" }

func (m *mockAuthProvider) Authenticate(ctx context.Context, credentials interface{}) (*plugin.User, error) {
	if m.authFunc != nil {
		return m.authFunc(ctx, credentials)
	}
	return nil, plugin.ErrInvalidCredentials
}

func (m *mockAuthProvider) ValidateSession(ctx context.Context, user *plugin.User) error {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, user)
	}
	return nil
}

func (m *mockAuthProvider) Logout(ctx context.Context, user *plugin.User) error {
	return nil
}

func TestAuthContext(t *testing.T) {
	// Test user context functions
	user := &plugin.User{
		ID:    123,
		Email: "test@example.com",
		Role:  "user",
	}

	ctx := plugin.SetUserInContext(context.Background(), user)

	// Get user from context
	retrievedUser, ok := plugin.GetUserFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Email, retrievedUser.Email)
	assert.Equal(t, user.Role, retrievedUser.Role)

	// Test with empty context
	_, ok = plugin.GetUserFromContext(context.Background())
	assert.False(t, ok)
}

func TestPasswordHashing(t *testing.T) {
	password := "test-password-123"

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	// Verify correct password
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	assert.NoError(t, err)

	// Verify incorrect password
	err = bcrypt.CompareHashAndPassword(hash, []byte("wrong-password"))
	assert.Error(t, err)
}

func TestLoginFormValidation(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		expectError bool
	}{
		{
			name:        "valid credentials",
			email:       "test@example.com",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "empty email",
			email:       "",
			password:    "password123",
			expectError: true,
		},
		{
			name:        "empty password",
			email:       "test@example.com",
			password:    "",
			expectError: true,
		},
		{
			name:        "invalid email format",
			email:       "not-an-email",
			password:    "password123",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple validation logic
			err := validateLoginCredentials(tt.email, tt.password)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function for credential validation
func validateLoginCredentials(email, password string) error {
	if email == "" || password == "" {
		return plugin.ErrInvalidCredentials
	}
	if !strings.Contains(email, "@") {
		return plugin.ErrInvalidCredentials
	}
	return nil
}
