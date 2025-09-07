package testutil

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

// TestJWT represents test JWT configuration
type TestJWT struct {
	Secret    string
	ExpiresIn time.Duration
}

// DefaultTestJWT returns a default JWT configuration for testing
func DefaultTestJWT() *TestJWT {
	return &TestJWT{
		Secret:    "test-secret-key-for-testing-only",
		ExpiresIn: time.Hour,
	}
}

// GenerateToken generates a test JWT token
func (j *TestJWT) GenerateToken(claims jwt.MapClaims) (string, error) {
	if claims["exp"] == nil {
		claims["exp"] = time.Now().Add(j.ExpiresIn).Unix()
	}
	if claims["iat"] == nil {
		claims["iat"] = time.Now().Unix()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.Secret))
}

// GenerateAdminToken generates a test admin JWT token
func (j *TestJWT) GenerateAdminToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    "admin",
		"email":   "admin@test.com",
	}
	return j.GenerateToken(claims)
}

// GenerateUserToken generates a test user JWT token
func (j *TestJWT) GenerateUserToken(userID uint, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    "user",
		"email":   email,
	}
	return j.GenerateToken(claims)
}

// GenerateExpiredToken generates an expired JWT token
func (j *TestJWT) GenerateExpiredToken(claims jwt.MapClaims) (string, error) {
	claims["exp"] = time.Now().Add(-time.Hour).Unix()
	claims["iat"] = time.Now().Add(-2 * time.Hour).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.Secret))
}

// AuthRequest adds authentication to an HTTP request
func AuthRequest(req *http.Request, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
}

// RequireAuth creates an authenticated request for testing
func RequireAuth(t *testing.T, method, path string, body interface{}, token string) *http.Request {
	req := HTTPRequest(t, method, path, body)
	AuthRequest(req, token)
	return req
}

// AssertUnauthorized checks that a response indicates unauthorized access
func AssertUnauthorized(t *testing.T, statusCode int) {
	require.Equal(t, http.StatusUnauthorized, statusCode, "expected unauthorized response")
}

// AssertForbidden checks that a response indicates forbidden access
func AssertForbidden(t *testing.T, statusCode int) {
	require.Equal(t, http.StatusForbidden, statusCode, "expected forbidden response")
}

// TestUser represents a test user
type TestUser struct {
	ID       uint
	Email    string
	Password string
	Role     string
}

// DefaultTestUsers returns a set of default test users
func DefaultTestUsers() []TestUser {
	return []TestUser{
		{ID: 1, Email: "admin@test.com", Password: "admin123", Role: "admin"},
		{ID: 2, Email: "user1@test.com", Password: "user123", Role: "user"},
		{ID: 3, Email: "user2@test.com", Password: "user123", Role: "user"},
	}
}

// MockAuthMiddleware creates a mock authentication middleware for testing
func MockAuthMiddleware(userID uint, role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add mock auth context to request
			ctx := r.Context()
			// You would add your actual context values here
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateToken validates a JWT token for testing
func ValidateToken(t *testing.T, tokenString string, secret string) jwt.MapClaims {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)
	require.True(t, token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)

	return claims
}
