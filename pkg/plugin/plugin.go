package plugin

import (
	"context"
	"net/http"
)

// Plugin is the base interface that all plugins must implement
type Plugin interface {
	// Metadata
	ID() string          // Unique identifier (e.g., "com.obtura.auth")
	Name() string        // Human-readable name
	Version() string     // Semantic version
	Description() string // Brief description
	Author() string      // Plugin author
	
	// Lifecycle
	Init(ctx context.Context) error    // Initialize plugin
	Start(ctx context.Context) error   // Start plugin services
	Stop(ctx context.Context) error    // Stop plugin services
	Destroy(ctx context.Context) error // Cleanup resources
	
	// Dependencies
	Dependencies() []string // List of required plugin IDs
	
	// Configuration
	Config() interface{}           // Plugin configuration struct
	ValidateConfig() error         // Validate configuration
	DefaultConfig() interface{}    // Default configuration
}

// RoutablePlugin provides HTTP routes
type RoutablePlugin interface {
	Plugin
	Routes() []Route
}

// Route represents an HTTP route
type Route struct {
	Method      string
	Path        string
	Handler     http.HandlerFunc
	Middlewares []func(http.Handler) http.Handler
}

// AdminPlugin provides admin interface
type AdminPlugin interface {
	Plugin
	AdminRoutes() []Route
	AdminNavigation() []NavItem
}

// NavItem represents a navigation menu item
type NavItem struct {
	Title    string
	Path     string
	Icon     string // Icon identifier
	Order    int
	Children []NavItem
}

// HookablePlugin can register and respond to hooks
type HookablePlugin interface {
	Plugin
	Hooks() map[string]HookHandler
}

// HookHandler handles a specific hook
type HookHandler func(ctx context.Context, data interface{}) (interface{}, error)

// AssetPlugin provides static assets
type AssetPlugin interface {
	Plugin
	Assets() map[string][]byte // Path -> content
	AssetPaths() []string       // Paths to serve
}

// TemplatePlugin provides templates
type TemplatePlugin interface {
	Plugin
	Templates() map[string]string // Name -> template content
}

// ServicePlugin provides a service that other plugins can use
type ServicePlugin interface {
	Plugin
	Service() interface{} // The service instance
}

// MigrationPlugin provides database migrations
type MigrationPlugin interface {
	Plugin
	Migrations() []Migration
}

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	Up          func() error
	Down        func() error
}

// MiddlewarePlugin provides middleware
type MiddlewarePlugin interface {
	Plugin
	Middleware() func(http.Handler) http.Handler
}

// EventPlugin can emit and listen to events
type EventPlugin interface {
	Plugin
	EventHandlers() map[string]EventHandler
}

// EventHandler handles an event
type EventHandler func(ctx context.Context, event Event) error

// Event represents a system event
type Event struct {
	Name    string
	Source  string      // Plugin ID that emitted the event
	Data    interface{}
	Context context.Context
}

// DocumentablePlugin provides self-documentation
type DocumentablePlugin interface {
	Plugin
	Documentation() PluginDocumentation
}

// PluginDocumentation contains all documentation for a plugin
type PluginDocumentation struct {
	Overview     string                    // Main plugin overview
	Installation string                    // Installation instructions
	Configuration string                   // Configuration guide
	Usage        string                    // Usage examples
	API          []APIEndpoint            // API endpoints documentation
	Hooks        []HookDocumentation      // Available hooks
	Events       []EventDocumentation     // Events emitted/consumed
	Pages        []PageDocumentation      // Pages provided
	Settings     []SettingDocumentation   // Available settings
	FAQ          []FAQItem                // Frequently asked questions
}

// APIEndpoint documents a REST API endpoint
type APIEndpoint struct {
	Method      string
	Path        string
	Description string
	Parameters  []Parameter
	Response    ResponseDoc
	Example     string
}

// Parameter documents an API parameter
type Parameter struct {
	Name        string
	Type        string
	Required    bool
	Description string
	Default     interface{}
}

// ResponseDoc documents an API response
type ResponseDoc struct {
	StatusCodes map[int]string // Status code -> description
	Schema      string         // JSON schema or description
	Example     string
}

// HookDocumentation documents a hook
type HookDocumentation struct {
	Name        string
	Description string
	Parameters  string // Description of data passed
	Returns     string // Description of expected return
	Example     string
}

// EventDocumentation documents an event
type EventDocumentation struct {
	Name        string
	Type        string // "emitted" or "consumed"
	Description string
	Payload     string // Description of event data
	Example     string
}

// PageDocumentation documents a page provided by the plugin
type PageDocumentation struct {
	Title       string
	Path        string
	Description string
	Screenshot  string // Path to screenshot
	Access      string // Public, authenticated, admin
}

// SettingDocumentation documents a plugin setting
type SettingDocumentation struct {
	Key         string
	Type        string // string, int, bool, etc.
	Default     interface{}
	Description string
	Validation  string // Validation rules description
}

// FAQItem represents a frequently asked question
type FAQItem struct {
	Question string
	Answer   string
}

// PageProviderPlugin provides pages
type PageProviderPlugin interface {
	Plugin
	Pages() []Page
}

// Page represents a page provided by a plugin
type Page struct {
	ID          string
	Title       string
	Path        string
	Description string
	Handler     http.HandlerFunc
	Layout      string   // Layout to use
	Access      []string // Required permissions
	NavGroup    string   // Navigation group
	NavOrder    int      // Order in navigation
	Icon        string   // Icon identifier
}

// SettingsPlugin provides configurable settings
type SettingsPlugin interface {
	Plugin
	Settings() []Setting
	OnSettingChange(key string, oldValue, newValue interface{}) error
}

// Setting represents a configurable setting
type Setting struct {
	Key         string
	Name        string
	Description string
	Type        SettingType
	Default     interface{}
	Options     []SettingOption // For select/radio types
	Validation  SettingValidation
	Group       string // Settings group
	Order       int    // Display order
}

// SettingType defines the type of setting
type SettingType string

const (
	SettingTypeString   SettingType = "string"
	SettingTypeInt      SettingType = "int"
	SettingTypeBool     SettingType = "bool"
	SettingTypeSelect   SettingType = "select"
	SettingTypeMulti    SettingType = "multi"
	SettingTypeJSON     SettingType = "json"
	SettingTypeFile     SettingType = "file"
	SettingTypeColor    SettingType = "color"
	SettingTypeDateTime SettingType = "datetime"
)

// SettingOption represents an option for select/radio settings
type SettingOption struct {
	Value string
	Label string
}

// SettingValidation defines validation rules for a setting
type SettingValidation struct {
	Required bool
	Min      interface{} // Min value/length
	Max      interface{} // Max value/length
	Pattern  string      // Regex pattern
	Custom   func(value interface{}) error
}