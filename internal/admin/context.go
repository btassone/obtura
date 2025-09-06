package admin

import (
	"context"
	"net/http"

	"github.com/btassone/obtura/internal/models"
	"github.com/btassone/obtura/pkg/plugin"
	authPlugin "github.com/btassone/obtura/plugins/auth"
)

// AdminContext holds data for admin templates
type AdminContext struct {
	User *models.User
}

// GetAdminContext extracts the admin context from the request
func GetAdminContext(r *http.Request) *AdminContext {
	// Try to get auth user from context
	if authUser, ok := r.Context().Value(authPlugin.UserContextKey).(plugin.AuthUser); ok {
		// Convert to models.User for templates
		user := &models.User{
			Name:  authUser.Name(),
			Email: authUser.Email(),
			Role:  authUser.Role(),
		}
		return &AdminContext{
			User: user,
		}
	}
	return &AdminContext{}
}

// WithAdminContext adds admin context to the request context
func WithAdminContext(r *http.Request, ctx *AdminContext) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), "adminContext", ctx))
}