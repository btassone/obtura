package hub

import "github.com/btassone/obtura/pkg/plugin"

// PluginInfo contains comprehensive information about a plugin
type PluginInfo struct {
	// Basic info
	ID          string
	Name        string
	Version     string
	Description string
	Author      string
	IsActive    bool
	
	// Features
	ProvidesRoutes      bool
	ProvidesAdmin       bool
	ProvidesPages       bool
	ProvidesSettings    bool
	ProvidesHooks       bool
	ProvidesEvents      bool
	ProvidesMigrations  bool
	ProvidesAssets      bool
	ProvidesServices    bool
	
	// Documentation
	Documentation *plugin.PluginDocumentation
	
	// Pages
	Pages []plugin.Page
	
	// Settings
	Settings []plugin.Setting
	
	// Routes
	Routes      []plugin.Route
	AdminRoutes []plugin.Route
	
	// Dependencies
	Dependencies []string
	DependedBy   []string
	
	// Additional fields for template usage
	Type     string
	Active   bool
	Priority int
}