package analytics

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/btassone/obtura/pkg/plugin"
)

// Config holds the plugin configuration
type Config struct {
	Enabled       bool   `json:"enabled" label:"Enable Analytics" description:"Enable tracking of page views"`
	TrackingCode  string `json:"tracking_code" label:"Tracking Code" description:"Your analytics tracking code"`
	ExcludeAdmin  bool   `json:"exclude_admin" label:"Exclude Admin" description:"Don't track admin page views" default:"true"`
}

// Plugin tracks simple page analytics
type Plugin struct {
	config    *Config
	pageViews map[string]int
	mu        sync.RWMutex
}

// NewPlugin creates a new analytics plugin
func NewPlugin() *Plugin {
	return &Plugin{
		config: &Config{
			Enabled:      true,
			ExcludeAdmin: true,
		},
		pageViews: make(map[string]int),
	}
}

// Plugin interface implementation

func (p *Plugin) ID() string          { return "com.example.analytics" }
func (p *Plugin) Name() string        { return "Simple Analytics" }
func (p *Plugin) Version() string     { return "1.0.0" }
func (p *Plugin) Description() string { return "Basic page view analytics" }
func (p *Plugin) Author() string      { return "Example Author" }
func (p *Plugin) Dependencies() []string { return []string{} }

func (p *Plugin) Init(ctx context.Context) error    { return nil }
func (p *Plugin) Start(ctx context.Context) error   { return nil }
func (p *Plugin) Stop(ctx context.Context) error    { return nil }
func (p *Plugin) Destroy(ctx context.Context) error { return nil }

func (p *Plugin) Config() interface{}        { return p.config }
func (p *Plugin) DefaultConfig() interface{} { return p.config }
func (p *Plugin) ValidateConfig() error      { return nil }

// MiddlewarePlugin implementation

func (p *Plugin) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Track page view if enabled
			if p.config.Enabled {
				path := r.URL.Path
				
				// Skip admin pages if configured
				if p.config.ExcludeAdmin && len(path) >= 6 && path[:6] == "/admin" {
					next.ServeHTTP(w, r)
					return
				}
				
				// Track the page view
				p.trackPageView(path)
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// AdminPlugin implementation

func (p *Plugin) AdminRoutes() []plugin.Route {
	return []plugin.Route{
		{
			Method:  http.MethodGet,
			Path:    "/analytics",
			Handler: p.handleAnalytics,
		},
		{
			Method:  http.MethodGet,
			Path:    "/analytics/api/stats",
			Handler: p.handleAPIStats,
		},
	}
}

func (p *Plugin) AdminNavigation() []plugin.NavItem {
	return []plugin.NavItem{
		{
			Title: "Analytics",
			Path:  "/admin/analytics",
			Icon:  "chart",
			Order: 150,
		},
	}
}

// Helper methods

func (p *Plugin) trackPageView(path string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pageViews[path]++
}

func (p *Plugin) getStats() map[string]int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	stats := make(map[string]int)
	for path, views := range p.pageViews {
		stats[path] = views
	}
	return stats
}

// HTTP handlers

func (p *Plugin) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Analytics</title>
			<script src="https://unpkg.com/htmx.org@1.9.10"></script>
			<style>
				body { font-family: Arial, sans-serif; margin: 20px; }
				.stats { background: #f5f5f5; padding: 20px; border-radius: 8px; }
				.stat-item { display: flex; justify-content: space-between; padding: 10px; border-bottom: 1px solid #ddd; }
				.stat-item:last-child { border-bottom: none; }
			</style>
		</head>
		<body>
			<h1>Simple Analytics</h1>
			<div class="stats" hx-get="/admin/analytics/api/stats" hx-trigger="load, every 5s">
				Loading stats...
			</div>
			<p><a href="/admin">Back to Admin</a></p>
		</body>
		</html>
	`
	
	w.Write([]byte(html))
}

func (p *Plugin) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	stats := p.getStats()
	
	if len(stats) == 0 {
		w.Write([]byte(`<p>No page views tracked yet.</p>`))
		return
	}
	
	html := ""
	for path, views := range stats {
		html += fmt.Sprintf(`<div class="stat-item"><span>%s</span><strong>%d views</strong></div>`, path, views)
	}
	
	w.Write([]byte(html))
}

// EventPlugin implementation

func (p *Plugin) EventHandlers() map[string]plugin.EventHandler {
	return map[string]plugin.EventHandler{
		"page.viewed": func(ctx context.Context, event plugin.Event) error {
			// Handle page view events from other plugins
			if data, ok := event.Data.(map[string]interface{}); ok {
				if path, ok := data["path"].(string); ok {
					p.trackPageView(path)
				}
			}
			return nil
		},
	}
}