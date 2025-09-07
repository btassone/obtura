package admin

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/btassone/obtura/internal/types"
	"github.com/btassone/obtura/pkg/plugin"
	adminpages "github.com/btassone/obtura/web/templates/admin/pages"
)

// PluginListData represents data for the plugins list page
type PluginListData struct {
	User    interface{} // User data
	Plugins []types.PluginInfo
}

// handlePluginsListWithRegistry handles the plugins list page
func handlePluginsListWithRegistry(registry *plugin.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		plugins := registry.List()
		
		// Convert plugins to display format
		var pluginInfos []types.PluginInfo
		for _, p := range plugins {
			info := types.PluginInfo{
				ID:          p.ID(),
				Name:        p.Name(),
				Description: p.Description(),
				Version:     p.Version(),
				Author:      p.Author(),
				IsActive:    true, // All registered plugins are active
				IsCore:      isCore(p.ID()),
				HasConfig:   hasConfig(p),
			}
			pluginInfos = append(pluginInfos, info)
		}
		
		// TODO: Create a new template that accepts plugin data
		component := adminpages.PluginsListWithData(user, pluginInfos)
		templ.Handler(component).ServeHTTP(w, r)
	}
}

// handlePluginToggle toggles a plugin on/off
func handlePluginToggleWithRegistry(registry *plugin.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement plugin enable/disable
		// For now, just redirect back
		http.Redirect(w, r, "/admin/plugins", http.StatusSeeOther)
	}
}

// handlePluginConfig handles plugin configuration page
func handlePluginConfigWithRegistry(registry *plugin.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pluginID := r.PathValue("id")
		user := getUser(r)
		
		p, err := registry.Get(pluginID)
		if err != nil {
			http.Error(w, "Plugin not found", http.StatusNotFound)
			return
		}
		
		// Get config schema and current config
		schema, _ := registry.GetConfigManager().GetSchema(pluginID)
		config, _ := registry.GetConfig(pluginID)
		
		// Convert config to map for template
		configMap, ok := config.(map[string]interface{})
		if !ok {
			configMap = make(map[string]interface{})
		}
		
		component := adminpages.PluginConfig(user, p, schema, configMap)
		templ.Handler(component).ServeHTTP(w, r)
	}
}

// handlePluginConfigUpdate handles plugin configuration updates
func handlePluginConfigUpdateWithRegistry(registry *plugin.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pluginID := r.PathValue("id")
		
		// TODO: Parse form data and update config
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}
		
		// Convert form data to config
		config := make(map[string]interface{})
		for key, values := range r.Form {
			if len(values) > 0 {
				config[key] = values[0]
			}
		}
		
		// Update config
		if err := registry.SetConfig(pluginID, config); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		http.Redirect(w, r, "/admin/plugins/"+pluginID+"/config", http.StatusSeeOther)
	}
}

// isCore checks if a plugin is a core plugin
func isCore(pluginID string) bool {
	corePlugins := []string{
		"com.obtura.auth",
		"com.obtura.pages",
		"com.obtura.themes",
		"com.obtura.media",
	}
	
	for _, core := range corePlugins {
		if pluginID == core {
			return true
		}
	}
	return false
}

// hasConfig checks if a plugin has configuration
func hasConfig(p plugin.Plugin) bool {
	// Check if plugin config is not nil and not empty
	config := p.Config()
	return config != nil
}