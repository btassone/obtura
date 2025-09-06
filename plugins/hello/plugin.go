package hello

import (
	"context"
	"fmt"
	"net/http"

	"github.com/btassone/obtura/pkg/plugin"
)

// Config holds the plugin configuration
type Config struct {
	Greeting      string `json:"greeting" label:"Greeting Message" description:"The message to display" default:"Hello from Obtura!" required:"true"`
	ShowTimestamp bool   `json:"show_timestamp" label:"Show Timestamp" description:"Display current time with greeting"`
	TextColor     string `json:"text_color" label:"Text Color" description:"Color of the greeting text" default:"blue"`
}

// Plugin is a simple example plugin
type Plugin struct {
	config *Config
}

// NewPlugin creates a new hello plugin
func NewPlugin() *Plugin {
	return &Plugin{
		config: &Config{
			Greeting:      "Hello from Obtura!",
			ShowTimestamp: false,
			TextColor:     "blue",
		},
	}
}

// Plugin interface implementation

func (p *Plugin) ID() string          { return "com.example.hello" }
func (p *Plugin) Name() string        { return "Hello World" }
func (p *Plugin) Version() string     { return "1.0.0" }
func (p *Plugin) Description() string { return "A simple hello world plugin example" }
func (p *Plugin) Author() string      { return "Example Author" }
func (p *Plugin) Dependencies() []string { return []string{} }

func (p *Plugin) Init(ctx context.Context) error {
	// Plugin initialization
	return nil
}

func (p *Plugin) Start(ctx context.Context) error {
	// Plugin startup
	return nil
}

func (p *Plugin) Stop(ctx context.Context) error  { return nil }
func (p *Plugin) Destroy(ctx context.Context) error { return nil }

func (p *Plugin) Config() interface{}        { return p.config }
func (p *Plugin) DefaultConfig() interface{} { return p.config }
func (p *Plugin) ValidateConfig() error {
	if p.config.Greeting == "" {
		return fmt.Errorf("greeting cannot be empty")
	}
	return nil
}

// RoutablePlugin implementation

func (p *Plugin) Routes() []plugin.Route {
	return []plugin.Route{
		{
			Method:  http.MethodGet,
			Path:    "/hello",
			Handler: p.handleHello,
		},
	}
}

// AdminPlugin implementation

func (p *Plugin) AdminRoutes() []plugin.Route {
	return []plugin.Route{
		{
			Method:  http.MethodGet,
			Path:    "/hello/stats",
			Handler: p.handleStats,
		},
	}
}

func (p *Plugin) AdminNavigation() []plugin.NavItem {
	return []plugin.NavItem{
		{
			Title: "Hello Plugin",
			Path:  "/admin/hello/stats",
			Icon:  "chat",
			Order: 200,
		},
	}
}

// HTTP handlers

func (p *Plugin) handleHello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	html := fmt.Sprintf(`
		<html>
		<head>
			<title>Hello Plugin</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					display: flex;
					justify-content: center;
					align-items: center;
					height: 100vh;
					margin: 0;
					background-color: #f0f0f0;
				}
				.greeting {
					text-align: center;
					padding: 2rem;
					background: white;
					border-radius: 8px;
					box-shadow: 0 2px 8px rgba(0,0,0,0.1);
				}
				.greeting h1 {
					color: %s;
					margin: 0 0 1rem 0;
				}
				.timestamp {
					color: #666;
					font-size: 0.9rem;
				}
			</style>
		</head>
		<body>
			<div class="greeting">
				<h1>%s</h1>
				%s
				<p><a href="/">Back to Home</a></p>
			</div>
		</body>
		</html>
	`, p.config.TextColor, p.config.Greeting, p.getTimestamp())
	
	w.Write([]byte(html))
}

func (p *Plugin) handleStats(w http.ResponseWriter, r *http.Request) {
	// Simple stats page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`
		<h1>Hello Plugin Stats</h1>
		<p>This is the admin stats page for the Hello plugin.</p>
		<p>Current greeting: ` + p.config.Greeting + `</p>
	`))
}

func (p *Plugin) getTimestamp() string {
	if p.config.ShowTimestamp {
		return `<p class="timestamp">Current time: ` + context.Background().Value("time").(string) + `</p>`
	}
	return ""
}

// DocumentablePlugin implementation

func (p *Plugin) Documentation() plugin.PluginDocumentation {
	return plugin.PluginDocumentation{
		Overview: `The Hello World plugin is a simple example that demonstrates the basic plugin structure in Obtura. It provides a greeting page and shows how to implement configuration, routes, and admin interfaces.`,
		Installation: `No special installation required. The plugin is included by default.`,
		Configuration: `
The plugin supports the following configuration options:
- greeting: The message to display (default: "Hello from Obtura!")
- show_timestamp: Whether to show the current time (default: false)
- text_color: Color of the greeting text (default: "blue")
`,
		Usage: `Navigate to /hello to see the greeting page. Admins can view stats at /admin/hello/stats.`,
		API: []plugin.APIEndpoint{
			{
				Method:      "GET",
				Path:        "/hello",
				Description: "Display the greeting page",
				Response: plugin.ResponseDoc{
					StatusCodes: map[int]string{
						200: "Success - returns HTML page with greeting",
					},
				},
			},
		},
		Pages: []plugin.PageDocumentation{
			{
				Title:       "Hello Page",
				Path:        "/hello",
				Description: "Displays a configurable greeting message",
				Access:      "Public",
			},
			{
				Title:       "Hello Stats",
				Path:        "/admin/hello/stats",
				Description: "Admin statistics for the Hello plugin",
				Access:      "Admin",
			},
		},
		Settings: []plugin.SettingDocumentation{
			{
				Key:         "greeting",
				Type:        "string",
				Default:     "Hello from Obtura!",
				Description: "The greeting message to display",
				Validation:  "Required, non-empty string",
			},
			{
				Key:         "show_timestamp",
				Type:        "bool",
				Default:     false,
				Description: "Whether to show the current timestamp",
			},
			{
				Key:         "text_color",
				Type:        "string",
				Default:     "blue",
				Description: "CSS color for the greeting text",
			},
		},
		FAQ: []plugin.FAQItem{
			{
				Question: "How do I change the greeting message?",
				Answer:   "You can change the greeting message in the plugin settings under Admin > Plugin Hub > Hello World > Settings.",
			},
			{
				Question: "Can I use HTML in the greeting?",
				Answer:   "No, for security reasons the greeting is displayed as plain text only.",
			},
		},
	}
}

// PageProviderPlugin implementation

func (p *Plugin) Pages() []plugin.Page {
	return []plugin.Page{
		{
			ID:          "hello-main",
			Title:       "Hello",
			Path:        "/hello",
			Description: "A simple greeting page",
			Handler:     p.handleHello,
			Access:      []string{"public"},
			NavGroup:    "main",
			NavOrder:    100,
			Icon:        "chat",
		},
	}
}

// SettingsPlugin implementation

func (p *Plugin) Settings() []plugin.Setting {
	return []plugin.Setting{
		{
			Key:         "greeting",
			Name:        "Greeting Message",
			Description: "The message to display on the hello page",
			Type:        plugin.SettingTypeString,
			Default:     "Hello from Obtura!",
			Validation: plugin.SettingValidation{
				Required: true,
				Min:      1,
				Max:      200,
			},
			Group: "Display",
			Order: 1,
		},
		{
			Key:         "show_timestamp",
			Name:        "Show Timestamp",
			Description: "Display the current time with the greeting",
			Type:        plugin.SettingTypeBool,
			Default:     false,
			Group:       "Display",
			Order:       2,
		},
		{
			Key:         "text_color",
			Name:        "Text Color",
			Description: "Color of the greeting text",
			Type:        plugin.SettingTypeColor,
			Default:     "#0000FF",
			Group:       "Display",
			Order:       3,
		},
	}
}

func (p *Plugin) OnSettingChange(key string, oldValue, newValue interface{}) error {
	// Update the config when settings change
	switch key {
	case "greeting":
		if str, ok := newValue.(string); ok {
			p.config.Greeting = str
		}
	case "show_timestamp":
		if b, ok := newValue.(bool); ok {
			p.config.ShowTimestamp = b
		}
	case "text_color":
		if str, ok := newValue.(string); ok {
			p.config.TextColor = str
		}
	}
	return nil
}