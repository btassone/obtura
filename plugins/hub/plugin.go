// Package hub provides a central dashboard for all plugins
package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/btassone/obtura/pkg/plugin"
)

// Plugin provides a central hub for plugin discovery and management
type Plugin struct {
	id          string
	name        string
	version     string
	description string
	author      string
	registry    *plugin.Registry
}

// NewPlugin creates a new plugin hub instance
func NewPlugin(registry *plugin.Registry) *Plugin {
	return &Plugin{
		id:          "com.obtura.hub",
		name:        "Plugin Hub",
		version:     "1.0.0",
		description: "Central dashboard for discovering and managing plugins",
		author:      "Obtura Team",
		registry:    registry,
	}
}

// ID returns the plugin ID
func (p *Plugin) ID() string {
	return p.id
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return p.name
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return p.version
}

// Description returns the plugin description
func (p *Plugin) Description() string {
	return p.description
}

// Author returns the plugin author
func (p *Plugin) Author() string {
	return p.author
}

// Dependencies returns required plugin IDs
func (p *Plugin) Dependencies() []string {
	return []string{}
}

// Init initializes the plugin
func (p *Plugin) Init(ctx context.Context) error {
	return nil
}

// Start begins the plugin operation
func (p *Plugin) Start(ctx context.Context) error {
	return nil
}

// Stop halts the plugin
func (p *Plugin) Stop(ctx context.Context) error {
	return nil
}

// Destroy cleans up plugin resources
func (p *Plugin) Destroy(ctx context.Context) error {
	return nil
}

// ValidateConfig validates the plugin configuration
func (p *Plugin) ValidateConfig() error {
	return nil
}

// Config returns plugin configuration
func (p *Plugin) Config() interface{} {
	return struct{}{}
}

// DefaultConfig returns default configuration
func (p *Plugin) DefaultConfig() interface{} {
	return p.Config()
}

// Routes returns frontend routes
func (p *Plugin) Routes() []plugin.Route {
	return []plugin.Route{
		{
			Method:  http.MethodGet,
			Path:    "/hub",
			Handler: p.handleHub,
		},
		{
			Method:  http.MethodGet,
			Path:    "/hub/plugins/{id}",
			Handler: p.handlePluginDetail,
		},
	}
}

// AdminRoutes returns admin routes
func (p *Plugin) AdminRoutes() []plugin.Route {
	return []plugin.Route{
		{
			Method:  http.MethodGet,
			Path:    "/hub",
			Handler: p.handleAdminHub,
		},
		{
			Method:  http.MethodGet,
			Path:    "/hub/plugins/{id}",
			Handler: p.handleAdminPluginDetail,
		},
		{
			Method:  http.MethodGet,
			Path:    "/hub/plugins/{id}/docs",
			Handler: p.handlePluginDocs,
		},
		{
			Method:  http.MethodGet,
			Path:    "/hub/plugins/{id}/settings",
			Handler: p.handlePluginSettings,
		},
		{
			Method:  http.MethodPost,
			Path:    "/hub/plugins/{id}/settings",
			Handler: p.handleSaveSettings,
		},
	}
}

// AdminNavigation returns admin navigation items
func (p *Plugin) AdminNavigation() []plugin.NavItem {
	return []plugin.NavItem{
		{
			Title: "Plugin Hub",
			Path:  "/admin/hub",
			Icon:  "puzzle",
			Order: 10,
		},
	}
}

// Pages returns pages provided by this plugin
func (p *Plugin) Pages() []plugin.Page {
	return []plugin.Page{
		{
			ID:          "plugin-hub",
			Title:       "Plugin Hub",
			Path:        "/hub",
			Description: "Discover all available plugins and their features",
			Handler:     p.handleHub,
			Access:      []string{"public"},
			NavGroup:    "main",
			NavOrder:    50,
			Icon:        "puzzle",
		},
	}
}

// Documentation returns plugin documentation
func (p *Plugin) Documentation() plugin.PluginDocumentation {
	return plugin.PluginDocumentation{
		Overview: `The Plugin Hub provides a central location for discovering and managing all plugins in your Obtura installation.`,
		Installation: `The Plugin Hub is included by default in Obtura installations.`,
		Configuration: `No configuration required.`,
		Usage: `Navigate to /hub to see all available plugins and their features.`,
		Pages: []plugin.PageDocumentation{
			{
				Title:       "Plugin Hub",
				Path:        "/hub",
				Description: "Main plugin discovery page",
				Access:      "Public",
			},
			{
				Title:       "Plugin Hub Admin",
				Path:        "/admin/hub",
				Description: "Admin interface for managing plugins",
				Access:      "Admin",
			},
		},
	}
}

// handleHub displays the main plugin hub
func (p *Plugin) handleHub(w http.ResponseWriter, r *http.Request) {
	plugins := p.gatherPluginInfo()
	component := hubPage(plugins)
	templ.Handler(component).ServeHTTP(w, r)
}

// handlePluginDetail displays detailed information about a specific plugin
func (p *Plugin) handlePluginDetail(w http.ResponseWriter, r *http.Request) {
	pluginID := r.PathValue("id")
	
	pluginInfo, err := p.getPluginInfo(pluginID)
	if err != nil {
		http.Error(w, "Plugin not found", http.StatusNotFound)
		return
	}
	
	component := pluginDetailPage(pluginInfo)
	templ.Handler(component).ServeHTTP(w, r)
}

// handleAdminHub displays the admin plugin hub
func (p *Plugin) handleAdminHub(w http.ResponseWriter, r *http.Request) {
	plugins := p.gatherPluginInfo()
	component := adminHubPage(plugins)
	templ.Handler(component).ServeHTTP(w, r)
}

// handleAdminPluginDetail displays admin view of plugin details
func (p *Plugin) handleAdminPluginDetail(w http.ResponseWriter, r *http.Request) {
	pluginID := r.PathValue("id")
	
	pluginInfo, err := p.getPluginInfo(pluginID)
	if err != nil {
		http.Error(w, "Plugin not found", http.StatusNotFound)
		return
	}
	
	component := adminPluginDetailPage(pluginInfo)
	templ.Handler(component).ServeHTTP(w, r)
}

// handlePluginDocs displays plugin documentation
func (p *Plugin) handlePluginDocs(w http.ResponseWriter, r *http.Request) {
	pluginID := r.PathValue("id")
	
	pluginInfo, err := p.getPluginInfo(pluginID)
	if err != nil {
		http.Error(w, "Plugin not found", http.StatusNotFound)
		return
	}
	
	component := pluginDocsPage(pluginInfo)
	templ.Handler(component).ServeHTTP(w, r)
}

// handlePluginSettings displays plugin settings
func (p *Plugin) handlePluginSettings(w http.ResponseWriter, r *http.Request) {
	pluginID := r.PathValue("id")
	
	pluginInfo, err := p.getPluginInfo(pluginID)
	if err != nil {
		http.Error(w, "Plugin not found", http.StatusNotFound)
		return
	}
	
	// Get current settings values
	settings := p.getPluginSettings(pluginID)
	
	// Check for success message
	if r.URL.Query().Get("success") == "true" {
		// Add success flag to context or template data
		// For now, we'll handle this in the template
	}
	
	component := pluginSettingsPage(pluginInfo, settings)
	templ.Handler(component).ServeHTTP(w, r)
}

// handleSaveSettings saves plugin settings
func (p *Plugin) handleSaveSettings(w http.ResponseWriter, r *http.Request) {
	pluginID := r.PathValue("id")
	
	// Get the plugin
	plg, err := p.registry.Get(pluginID)
	if err != nil {
		http.Error(w, "Plugin not found", http.StatusNotFound)
		return
	}
	
	// Check if plugin supports settings
	settingsPlugin, ok := plg.(plugin.SettingsPlugin)
	if !ok {
		http.Error(w, "Plugin does not support settings", http.StatusBadRequest)
		return
	}
	
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	
	// Get plugin settings
	settings := settingsPlugin.Settings()
	
	// Process each setting
	for _, setting := range settings {
		value := r.FormValue(setting.Key)
		
		// Convert value based on type
		var convertedValue interface{}
		switch setting.Type {
		case plugin.SettingTypeBool:
			convertedValue = value == "on" || value == "true"
		case plugin.SettingTypeInt:
			var intVal int
			fmt.Sscanf(value, "%d", &intVal)
			convertedValue = intVal
		default:
			convertedValue = value
		}
		
		// Get old value for comparison
		oldValue := p.getSettingValue(pluginID, setting.Key)
		
		// Notify plugin of change
		if err := settingsPlugin.OnSettingChange(setting.Key, oldValue, convertedValue); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update setting %s: %v", setting.Key, err), http.StatusInternalServerError)
			return
		}
		
		// Save to config manager
		p.saveSettingValue(pluginID, setting.Key, convertedValue)
	}
	
	// Redirect back to settings page with success message
	http.Redirect(w, r, fmt.Sprintf("/admin/hub/plugins/%s/settings?success=true", pluginID), http.StatusSeeOther)
}

// gatherPluginInfo collects information about all plugins
func (p *Plugin) gatherPluginInfo() []PluginInfo {
	var plugins []PluginInfo
	
	// Get all registered plugins
	allPlugins := p.registry.List()
	
	for _, plg := range allPlugins {
		info := PluginInfo{
			ID:          plg.ID(),
			Name:        plg.Name(),
			Version:     plg.Version(),
			Description: plg.Description(),
			Author:      plg.Author(),
			IsActive:    p.registry.IsEnabled(plg.ID()),
			Dependencies: plg.Dependencies(),
		}
		
		// Check for various interfaces
		if _, ok := plg.(plugin.RoutablePlugin); ok {
			info.ProvidesRoutes = true
			if routable, ok := plg.(plugin.RoutablePlugin); ok {
				info.Routes = routable.Routes()
			}
		}
		
		if _, ok := plg.(plugin.AdminPlugin); ok {
			info.ProvidesAdmin = true
			if admin, ok := plg.(plugin.AdminPlugin); ok {
				info.AdminRoutes = admin.AdminRoutes()
			}
		}
		
		if _, ok := plg.(plugin.PageProviderPlugin); ok {
			info.ProvidesPages = true
			if pages, ok := plg.(plugin.PageProviderPlugin); ok {
				info.Pages = pages.Pages()
			}
		}
		
		if _, ok := plg.(plugin.SettingsPlugin); ok {
			info.ProvidesSettings = true
			if settings, ok := plg.(plugin.SettingsPlugin); ok {
				info.Settings = settings.Settings()
			}
		}
		
		if _, ok := plg.(plugin.HookablePlugin); ok {
			info.ProvidesHooks = true
		}
		
		if _, ok := plg.(plugin.EventPlugin); ok {
			info.ProvidesEvents = true
		}
		
		if _, ok := plg.(plugin.MigrationPlugin); ok {
			info.ProvidesMigrations = true
		}
		
		if _, ok := plg.(plugin.AssetPlugin); ok {
			info.ProvidesAssets = true
		}
		
		if _, ok := plg.(plugin.ServicePlugin); ok {
			info.ProvidesServices = true
		}
		
		if _, ok := plg.(plugin.DocumentablePlugin); ok {
			if doc, ok := plg.(plugin.DocumentablePlugin); ok {
				docs := doc.Documentation()
				info.Documentation = &docs
			}
		}
		
		// Find plugins that depend on this one
		info.DependedBy = p.findDependents(plg.ID(), allPlugins)
		
		plugins = append(plugins, info)
	}
	
	return plugins
}

// getPluginInfo gets information about a specific plugin
func (p *Plugin) getPluginInfo(pluginID string) (*PluginInfo, error) {
	plugins := p.gatherPluginInfo()
	for _, plugin := range plugins {
		if plugin.ID == pluginID {
			return &plugin, nil
		}
	}
	return nil, fmt.Errorf("plugin not found: %s", pluginID)
}

// getPluginSettings gets current settings values for a plugin
func (p *Plugin) getPluginSettings(pluginID string) map[string]interface{} {
	configManager := p.registry.GetConfigManager()
	config, ok := configManager.GetConfig(pluginID)
	if !ok {
		// If no config exists, check if plugin provides default config
		if plg, err := p.registry.Get(pluginID); err == nil {
			if defaultConfig := plg.DefaultConfig(); defaultConfig != nil {
				// Save default config
				configManager.SetConfig(pluginID, defaultConfig)
				// Convert to map
				if configMap, ok := p.convertToMap(defaultConfig); ok {
					return configMap
				}
			}
		}
		return make(map[string]interface{})
	}
	
	// Convert to map
	if configMap, ok := p.convertToMap(config); ok {
		return configMap
	}
	return make(map[string]interface{})
}

// getSettingValue gets a specific setting value
func (p *Plugin) getSettingValue(pluginID, key string) interface{} {
	settings := p.getPluginSettings(pluginID)
	return settings[key]
}

// saveSettingValue saves a specific setting value
func (p *Plugin) saveSettingValue(pluginID, key string, value interface{}) {
	settings := p.getPluginSettings(pluginID)
	settings[key] = value
	p.registry.GetConfigManager().SetConfig(pluginID, settings)
}

// convertToMap converts a config interface to a map
func (p *Plugin) convertToMap(config interface{}) (map[string]interface{}, bool) {
	// If already a map, return it
	if m, ok := config.(map[string]interface{}); ok {
		return m, true
	}
	
	// Try to convert via JSON
	data, err := json.Marshal(config)
	if err != nil {
		return nil, false
	}
	
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, false
	}
	
	return m, true
}

// findDependents finds all plugins that depend on the given plugin
func (p *Plugin) findDependents(pluginID string, allPlugins []plugin.Plugin) []string {
	var dependents []string
	for _, plg := range allPlugins {
		for _, dep := range plg.Dependencies() {
			if dep == pluginID {
				dependents = append(dependents, plg.ID())
				break
			}
		}
	}
	return dependents
}

