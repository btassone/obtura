package auth

import (
	"context"
	"testing"
	"time"

	"github.com/btassone/obtura/internal/models"
	"github.com/btassone/obtura/pkg/database"
	"github.com/btassone/obtura/pkg/plugin"
	"github.com/btassone/obtura/test/testutil"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db := testutil.TestDBWithSchema(t, &models.User{})
	return db
}

func TestBasicAuthProvider_Metadata(t *testing.T) {
	provider := &BasicAuthProvider{}

	assert.Equal(t, "basic", provider.ID())
	assert.Equal(t, "Email/Password Authentication", provider.Name())
}

func TestBasicAuthProvider_Authenticate(t *testing.T) {
	db := setupTestDB(t)
	dbWrapper := &database.DB{}
	userRepo := models.NewUserRepository(dbWrapper)

	config := &plugin.AuthConfig{
		SessionSecret: "test-secret",
	}

	provider := NewBasicAuthProvider(dbWrapper, userRepo, config)
	provider.db = db // Override with test DB

	// Create test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	testUser := &models.User{
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}
	err = db.Create(testUser).Error
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name        string
		credentials plugin.BasicCredentials
		wantErr     bool
		errType     error
	}{
		{
			name: "valid credentials",
			credentials: plugin.BasicCredentials{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			credentials: plugin.BasicCredentials{
				Email:    "wrong@example.com",
				Password: "password123",
			},
			wantErr: true,
			errType: plugin.ErrInvalidCredentials,
		},
		{
			name: "invalid password",
			credentials: plugin.BasicCredentials{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			wantErr: true,
			errType: plugin.ErrInvalidCredentials,
		},
		{
			name: "empty credentials",
			credentials: plugin.BasicCredentials{
				Email:    "",
				Password: "",
			},
			wantErr: true,
			errType: plugin.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := provider.Authenticate(ctx, tt.credentials)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.credentials.Email, user.Email)
			}
		})
	}
}

func TestBasicAuthProvider_ValidateSession(t *testing.T) {
	db := setupTestDB(t)
	dbWrapper := &database.DB{}
	userRepo := models.NewUserRepository(dbWrapper)

	config := &plugin.AuthConfig{
		SessionSecret: "test-secret",
	}

	provider := NewBasicAuthProvider(dbWrapper, userRepo, config)
	provider.db = db // Override with test DB

	// Create test user in DB
	testUser := &models.User{
		Email: "test@example.com",
	}
	err := db.Create(testUser).Error
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		user    *plugin.User
		wantErr bool
	}{
		{
			name: "valid user",
			user: &plugin.User{
				ID:    testUser.ID,
				Email: testUser.Email,
			},
			wantErr: false,
		},
		{
			name: "non-existent user",
			user: &plugin.User{
				ID:    99999,
				Email: "fake@example.com",
			},
			wantErr: true,
		},
		{
			name:    "nil user",
			user:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.ValidateSession(ctx, tt.user)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBasicAuthProvider_CreateInitialAdmin(t *testing.T) {
	db := setupTestDB(t)
	dbWrapper := &database.DB{}
	userRepo := models.NewUserRepository(dbWrapper)

	config := &plugin.AuthConfig{
		SessionSecret: "test-secret",
	}

	provider := NewBasicAuthProvider(dbWrapper, userRepo, config)
	provider.db = db // Override with test DB

	// Test creating initial admin when no users exist
	err := provider.CreateInitialAdmin()
	require.NoError(t, err)

	// Verify admin was created
	var admin models.User
	err = db.Where("email = ?", "admin@obtura.local").First(&admin).Error
	require.NoError(t, err)
	assert.Equal(t, "admin@obtura.local", admin.Email)

	// Verify password is correct
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte("admin123"))
	assert.NoError(t, err)

	// Test that it doesn't create duplicate admin
	err = provider.CreateInitialAdmin()
	require.NoError(t, err)

	// Verify still only one admin
	var count int64
	db.Model(&models.User{}).Where("email = ?", "admin@obtura.local").Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestBasicAuthProvider_GenerateJWT(t *testing.T) {
	config := &plugin.AuthConfig{
		SessionSecret: "test-secret-key",
		SessionMaxAge: 3600, // 1 hour
	}

	provider := &BasicAuthProvider{
		config: config,
	}

	user := &plugin.User{
		ID:    123,
		Email: "test@example.com",
		Role:  "user",
	}

	// Generate token
	tokenString, err := provider.GenerateJWT(user)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.SessionSecret), nil
	})
	require.NoError(t, err)
	assert.True(t, token.Valid)

	// Check claims
	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, float64(123), claims["user_id"])
	assert.Equal(t, "test@example.com", claims["email"])
	assert.Equal(t, "user", claims["role"])

	// Check expiration
	exp, ok := claims["exp"].(float64)
	require.True(t, ok)
	expectedExp := time.Now().Add(time.Duration(config.SessionMaxAge) * time.Second).Unix()
	assert.InDelta(t, expectedExp, int64(exp), 5) // Allow 5 second difference
}

func TestBasicAuthProvider_ValidateJWT(t *testing.T) {
	config := &plugin.AuthConfig{
		SessionSecret: "test-secret-key",
		SessionMaxAge: 3600,
	}

	provider := &BasicAuthProvider{
		config: config,
	}

	// Create valid token
	claims := jwt.MapClaims{
		"user_id": float64(123),
		"email":   "test@example.com",
		"role":    "user",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	validToken, err := token.SignedString([]byte(config.SessionSecret))
	require.NoError(t, err)

	// Test valid token
	user, err := provider.ValidateJWT(validToken)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, uint(123), user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "user", user.Role)

	// Test expired token
	expiredClaims := jwt.MapClaims{
		"user_id": float64(123),
		"email":   "test@example.com",
		"role":    "user",
		"exp":     time.Now().Add(-time.Hour).Unix(), // Expired
	}

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString([]byte(config.SessionSecret))
	require.NoError(t, err)

	user, err = provider.ValidateJWT(expiredTokenString)
	assert.Error(t, err)
	assert.Nil(t, user)

	// Test invalid signature
	wrongToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	invalidToken, err := wrongToken.SignedString([]byte("wrong-secret"))
	require.NoError(t, err)

	user, err = provider.ValidateJWT(invalidToken)
	assert.Error(t, err)
	assert.Nil(t, user)

	// Test malformed token
	user, err = provider.ValidateJWT("not.a.valid.token")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestBasicAuthProvider_Logout(t *testing.T) {
	provider := &BasicAuthProvider{}

	user := &plugin.User{
		ID:    123,
		Email: "test@example.com",
	}

	// Logout should always succeed for basic auth
	err := provider.Logout(context.Background(), user)
	assert.NoError(t, err)

	// Test with nil user
	err = provider.Logout(context.Background(), nil)
	assert.NoError(t, err)
}

func TestBasicAuthProvider_WithNilDependencies(t *testing.T) {
	// Test that provider handles nil dependencies gracefully
	provider := &BasicAuthProvider{}

	ctx := context.Background()

	// These should not panic
	user, err := provider.Authenticate(ctx, plugin.BasicCredentials{})
	assert.Error(t, err)
	assert.Nil(t, user)

	err = provider.ValidateSession(ctx, &plugin.User{})
	assert.Error(t, err)

	err = provider.CreateInitialAdmin()
	assert.Error(t, err)
}
