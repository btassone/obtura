# Obtura Plugin Examples

This directory contains example plugins demonstrating different aspects of the Obtura plugin system.

## Plugin Examples

### 1. Basic Plugin (`basic-plugin.go`)
- Simplest plugin implementation
- Shows lifecycle methods (Initialize, Start, Stop)
- Demonstrates configuration structure

### 2. Blog Plugin (`blog-plugin.go`)
- Routable plugin with HTTP endpoints
- Shows how to add pages and APIs
- Includes Templ templates for rendering

### 3. Cache Service Plugin (`cache-service-plugin.go`)
- Service plugin providing functionality to other plugins
- In-memory cache with TTL support
- Background cleanup worker

### 4. SEO Plugin (`seo-plugin.go`)
- Hookable plugin modifying system behavior
- Demonstrates various hook types
- Shows content filtering and enhancement

### 5. Analytics Plugin (`analytics-admin-plugin.go`)
- Admin panel integration
- Dashboard and settings pages
- Public tracking endpoints
- API endpoints for data access

## Running the Examples

### Prerequisites
```bash
# Ensure templ is installed
go install github.com/a-h/templ/cmd/templ@latest

# Generate templ files
templ generate
```

### Option 1: Standalone Example
```bash
# Run the example main file
go run main_example.go

# Visit http://localhost:8080
```

### Option 2: Integration with Obtura
```go
// In your main Obtura application
import "github.com/btassone/obtura/examples"

// Register example plugins
registry.Register(examples.NewBlogPlugin())
registry.Register(examples.NewCacheServicePlugin())
// etc...
```

## Plugin Development Workflow

1. **Create Plugin Structure**
   ```go
   type MyPlugin struct {
       plugin.BasePlugin
   }
   ```

2. **Define Configuration**
   ```go
   func (p *MyPlugin) Config() interface{} {
       return &MyConfig{
           // Default values
       }
   }
   ```

3. **Implement Interfaces**
   - `RoutablePlugin` for HTTP routes
   - `ServicePlugin` for providing services
   - `HookablePlugin` for system hooks
   - `AdminPlugin` for admin panel

4. **Register and Test**
   ```go
   registry.Register(NewMyPlugin())
   ```

## Common Patterns

### Dependency Injection
```go
func NewMyPlugin(db *sql.DB, cache CacheService) *MyPlugin {
    return &MyPlugin{
        db:    db,
        cache: cache,
    }
}
```

### Service Discovery
```go
func (p *MyPlugin) Initialize(ctx context.Context) error {
    if service, ok := p.registry.GetService("com.example.cache"); ok {
        p.cache = service.(CacheService)
    }
    return nil
}
```

### Configuration Validation
```go
func (p *MyPlugin) Initialize(ctx context.Context) error {
    config := p.Config().(*MyConfig)
    if err := config.Validate(); err != nil {
        return fmt.Errorf("invalid configuration: %w", err)
    }
    return nil
}
```

### Background Workers
```go
func (p *MyPlugin) Start(ctx context.Context) error {
    go p.backgroundWorker(ctx)
    return nil
}

func (p *MyPlugin) backgroundWorker(ctx context.Context) {
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
}
```

## Testing Plugins

### Unit Testing
```go
func TestPlugin(t *testing.T) {
    plugin := NewMyPlugin()
    ctx := context.Background()
    
    // Test initialization
    err := plugin.Initialize(ctx)
    assert.NoError(t, err)
    
    // Test functionality
    // ...
}
```

### Integration Testing
```go
func TestPluginIntegration(t *testing.T) {
    registry := plugin.NewRegistry(chi.NewRouter())
    registry.Register(NewMyPlugin())
    
    ctx := context.Background()
    registry.Initialize(ctx)
    registry.Start(ctx)
    
    // Test plugin behavior
    // ...
    
    registry.Stop(ctx)
}
```

## Best Practices

1. **Use BasePlugin**: Inherit from `plugin.BasePlugin` for standard fields
2. **Context Handling**: Always respect context cancellation
3. **Error Handling**: Return meaningful errors with context
4. **Resource Cleanup**: Implement proper cleanup in `Stop()`
5. **Configuration**: Provide sensible defaults and validation
6. **Logging**: Use structured logging with plugin ID prefix
7. **Testing**: Write both unit and integration tests

## Debugging

Enable debug logging:
```go
func (p *MyPlugin) Initialize(ctx context.Context) error {
    if p.Config().(*MyConfig).Debug {
        p.logger = log.New(os.Stdout, "["+p.ID()+"] ", log.LstdFlags)
    }
    return nil
}
```

## Resources

- [Plugin Development Guide](../PLUGIN_DEVELOPMENT.md)
- [Obtura Documentation](https://docs.obtura.dev)
- [API Reference](https://pkg.go.dev/github.com/btassone/obtura/pkg/plugin)