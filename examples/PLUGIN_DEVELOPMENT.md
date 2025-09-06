# Obtura Plugin Development Guide

This guide explains how to create plugins for the Obtura framework. Plugins are the primary way to extend Obtura's functionality.

## Table of Contents

1. [Plugin Architecture](#plugin-architecture)
2. [Plugin Types](#plugin-types)
3. [Creating Your First Plugin](#creating-your-first-plugin)
4. [Plugin Interfaces](#plugin-interfaces)
5. [Best Practices](#best-practices)
6. [Examples](#examples)

## Plugin Architecture

Obtura follows a plugin-based architecture where almost everything is a plugin. This includes core functionality like authentication, pages, themes, and media handling.

### Key Concepts

- **Plugin Registry**: Central manager for all plugins
- **Plugin Lifecycle**: Initialize → Start → Stop
- **Plugin Types**: Basic, Routable, Service, Hookable, Admin
- **Configuration**: Each plugin can have its own configuration

## Plugin Types

### 1. Basic Plugin

The simplest plugin type that implements the core `Plugin` interface.

```go
type Plugin interface {
    ID() string
    Name() string
    Version() string
    Description() string
    Author() string
    Initialize(ctx context.Context) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Config() interface{}
    DefaultConfig() interface{}
}
```

**Use Case**: Background tasks, system utilities, data processing

### 2. Routable Plugin

Adds HTTP routes to your application.

```go
type RoutablePlugin interface {
    Plugin
    Routes() []Route
}
```

**Use Case**: Web pages, APIs, content delivery

### 3. Service Plugin

Provides services that other plugins can use.

```go
type ServicePlugin interface {
    Plugin
    Service() interface{}
}
```

**Use Case**: Caching, database connections, external API clients

### 4. Hookable Plugin

Can hook into system events and modify behavior.

```go
type HookablePlugin interface {
    Plugin
    Hooks() map[string]HookHandler
}
```

**Use Case**: SEO optimization, content filtering, event logging

### 5. Admin Plugin

Adds functionality to the admin panel.

```go
type AdminPlugin interface {
    Plugin
    AdminRoutes() []Route
}
```

**Use Case**: Analytics dashboards, content management, settings pages

## Creating Your First Plugin

### Step 1: Basic Structure

```go
package myplugin

import (
    "context"
    "github.com/btassone/obtura/pkg/plugin"
)

type MyPlugin struct {
    plugin.BasePlugin
}

func New() *MyPlugin {
    return &MyPlugin{
        BasePlugin: plugin.BasePlugin{
            PluginID:          "com.example.myplugin",
            PluginName:        "My Plugin",
            PluginVersion:     "1.0.0",
            PluginDescription: "Description of what my plugin does",
            PluginAuthor:      "Your Name",
        },
    }
}
```

### Step 2: Implement Required Methods

```go
func (p *MyPlugin) Initialize(ctx context.Context) error {
    // Setup database tables, load resources, etc.
    return nil
}

func (p *MyPlugin) Start(ctx context.Context) error {
    // Start background workers, open connections, etc.
    return nil
}

func (p *MyPlugin) Stop(ctx context.Context) error {
    // Cleanup resources, close connections, etc.
    return nil
}
```

### Step 3: Add Configuration

```go
type Config struct {
    Enabled    bool   `json:"enabled"`
    ApiKey     string `json:"api_key"`
    MaxRetries int    `json:"max_retries"`
}

func (p *MyPlugin) Config() interface{} {
    return &Config{
        Enabled:    true,
        ApiKey:     "",
        MaxRetries: 3,
    }
}

func (p *MyPlugin) DefaultConfig() interface{} {
    return p.Config()
}
```

### Step 4: Register Your Plugin

```go
// In your main application
registry := plugin.NewRegistry(router)
myPlugin := myplugin.New()
if err := registry.Register(myPlugin); err != nil {
    log.Fatal(err)
}
```

## Plugin Interfaces

### Route Structure

```go
type Route struct {
    Method      string
    Path        string
    Handler     http.Handler
    Middlewares []Middleware
}
```

### Hook Handler

```go
type HookHandler func(data interface{}) (interface{}, error)
```

### Middleware

```go
type Middleware func(http.Handler) http.Handler
```

## Best Practices

### 1. Plugin ID Convention

Use reverse domain notation:
- ✅ `com.example.myplugin`
- ❌ `my-plugin`

### 2. Error Handling

Always return meaningful errors:

```go
func (p *MyPlugin) Initialize(ctx context.Context) error {
    if err := p.createTables(); err != nil {
        return fmt.Errorf("failed to create tables: %w", err)
    }
    return nil
}
```

### 3. Context Usage

Respect context cancellation:

```go
func (p *MyPlugin) Start(ctx context.Context) error {
    go func() {
        ticker := time.NewTicker(time.Minute)
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                p.doWork()
            }
        }
    }()
    return nil
}
```

### 4. Configuration Validation

Validate configuration in Initialize:

```go
func (p *MyPlugin) Initialize(ctx context.Context) error {
    config := p.Config().(*Config)
    
    if config.ApiKey == "" && config.Enabled {
        return errors.New("API key is required when plugin is enabled")
    }
    
    return nil
}
```

### 5. Logging

Use structured logging:

```go
log.Printf("[%s] Processing request: path=%s method=%s", 
    p.ID(), r.URL.Path, r.Method)
```

### 6. Database Access

Use the database manager:

```go
func (p *MyPlugin) Initialize(ctx context.Context) error {
    // Get database from registry
    db := p.registry.GetDatabase()
    
    // Create tables
    return db.AutoMigrate(&MyModel{})
}
```

### 7. Template Organization

Keep templates with your plugin:

```
plugins/
  myplugin/
    plugin.go
    models.go
    handlers.go
    templates.templ
    assets/
      styles.css
      script.js
```

## Examples

### Example 1: Simple Page Plugin

```go
type AboutPlugin struct {
    plugin.BasePlugin
}

func (p *AboutPlugin) Routes() []plugin.Route {
    return []plugin.Route{
        {
            Method: http.MethodGet,
            Path:   "/about",
            Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                component := aboutPage()
                templ.Handler(component).ServeHTTP(w, r)
            }),
        },
    }
}
```

### Example 2: API Plugin

```go
type APIPlugin struct {
    plugin.BasePlugin
}

func (p *APIPlugin) Routes() []plugin.Route {
    return []plugin.Route{
        {
            Method: http.MethodGet,
            Path:   "/api/v1/status",
            Handler: http.HandlerFunc(p.handleStatus),
        },
        {
            Method: http.MethodPost,
            Path:   "/api/v1/data",
            Handler: http.HandlerFunc(p.handleData),
            Middlewares: []plugin.Middleware{
                p.authMiddleware,
                p.rateLimitMiddleware,
            },
        },
    }
}
```

### Example 3: Background Worker Plugin

```go
type WorkerPlugin struct {
    plugin.BasePlugin
    queue chan Job
}

func (p *WorkerPlugin) Start(ctx context.Context) error {
    p.queue = make(chan Job, 100)
    
    // Start workers
    for i := 0; i < 5; i++ {
        go p.worker(ctx, i)
    }
    
    return nil
}

func (p *WorkerPlugin) worker(ctx context.Context, id int) {
    for {
        select {
        case <-ctx.Done():
            return
        case job := <-p.queue:
            p.processJob(job)
        }
    }
}
```

### Example 4: Hook Plugin

```go
type SecurityPlugin struct {
    plugin.BasePlugin
}

func (p *SecurityPlugin) Hooks() map[string]plugin.HookHandler {
    return map[string]plugin.HookHandler{
        "before_save": p.sanitizeContent,
        "after_login": p.logLoginAttempt,
        "http_headers": p.addSecurityHeaders,
    }
}

func (p *SecurityPlugin) addSecurityHeaders(data interface{}) (interface{}, error) {
    if headers, ok := data.(http.Header); ok {
        headers.Set("X-Frame-Options", "DENY")
        headers.Set("X-Content-Type-Options", "nosniff")
        headers.Set("X-XSS-Protection", "1; mode=block")
    }
    return data, nil
}
```

### Example 5: Service Plugin

```go
type EmailService interface {
    Send(to, subject, body string) error
}

type EmailPlugin struct {
    plugin.BasePlugin
    client *smtp.Client
}

func (p *EmailPlugin) Service() interface{} {
    return EmailService(p)
}

func (p *EmailPlugin) Send(to, subject, body string) error {
    // Implementation
    return nil
}
```

## Plugin Testing

### Unit Testing

```go
func TestPluginInitialize(t *testing.T) {
    plugin := New()
    ctx := context.Background()
    
    err := plugin.Initialize(ctx)
    assert.NoError(t, err)
    
    assert.Equal(t, "com.example.myplugin", plugin.ID())
    assert.Equal(t, "1.0.0", plugin.Version())
}
```

### Integration Testing

```go
func TestPluginRoutes(t *testing.T) {
    plugin := NewBlogPlugin()
    routes := plugin.Routes()
    
    assert.Len(t, routes, 3)
    assert.Equal(t, "/blog", routes[0].Path)
    
    // Test handler
    req := httptest.NewRequest("GET", "/blog", nil)
    w := httptest.NewRecorder()
    
    routes[0].Handler.ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)
}
```

## Deployment

### Plugin Directory Structure

```
obtura/
  plugins/
    core/           # Core plugins
      auth/
      pages/
      themes/
    community/      # Community plugins
      analytics/
      seo/
      cache/
    custom/         # Your custom plugins
      myplugin/
```

### Plugin Manifest (plugin.yaml)

```yaml
id: com.example.myplugin
name: My Plugin
version: 1.0.0
description: Description of my plugin
author: Your Name
website: https://example.com
license: MIT
requires:
  obtura: ">=1.0.0"
  plugins:
    - com.obtura.auth: ">=1.0.0"
```

## Troubleshooting

### Common Issues

1. **Import Cycles**: Keep plugin dependencies minimal and use interfaces
2. **Context Leaks**: Always respect context cancellation
3. **Resource Leaks**: Clean up in Stop() method
4. **Race Conditions**: Use proper synchronization for shared state

### Debug Tips

1. Enable debug logging in plugin config
2. Use the plugin inspector in admin panel
3. Check plugin initialization order
4. Monitor resource usage

## Advanced Topics

### Plugin Dependencies

```go
func (p *MyPlugin) Initialize(ctx context.Context) error {
    // Check for required plugins
    if !p.registry.Has("com.obtura.auth") {
        return errors.New("auth plugin is required")
    }
    
    // Get service from another plugin
    if cache, ok := p.registry.GetService("com.example.cache"); ok {
        p.cache = cache.(CacheService)
    }
    
    return nil
}
```

### Dynamic Route Registration

```go
func (p *MyPlugin) Initialize(ctx context.Context) error {
    // Register routes based on configuration
    config := p.Config().(*Config)
    
    for _, endpoint := range config.Endpoints {
        route := plugin.Route{
            Method:  endpoint.Method,
            Path:    endpoint.Path,
            Handler: p.createHandler(endpoint),
        }
        p.registry.RegisterRoute(route)
    }
    
    return nil
}
```

### Plugin Communication

```go
// Using events
p.registry.PublishEvent("user.created", userData)

// Using hooks
result, err := p.registry.TriggerHook("before_save", content)

// Using services
service, _ := p.registry.GetService("com.example.email")
emailer := service.(EmailService)
emailer.Send(to, subject, body)
```

## Resources

- [Plugin Examples](/examples/plugins/) - Full working examples
- [API Reference](/docs/api/plugin/) - Complete API documentation
- [Plugin Registry](https://github.com/btassone/obtura-plugins) - Community plugins
- [Support Forum](https://forum.obtura.dev) - Get help and share ideas

## Contributing

We welcome plugin contributions! Please:

1. Follow the coding standards
2. Include tests
3. Document your plugin
4. Submit to the plugin registry

For more information, see [CONTRIBUTING.md](/CONTRIBUTING.md)