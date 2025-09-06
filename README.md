# Obtura

A modular web framework for Go inspired by Laravel and WordPress, built with modern technologies.

## Features

### ✅ Implemented
- **Plugin-based architecture** - Extensible plugin system with multiple plugin types
- **Hot reload development** - Powered by Air for instant feedback
- **Type-safe templates** - Using Templ for compile-time safety
- **Admin interface** - Built-in admin panel with plugin management
- **Database integration** - SQLite/PostgreSQL with migrations
- **Authentication system** - JWT-based auth with admin roles
- **Plugin registry** - Central management of all plugins
- **Service plugins** - Plugins can provide services to other plugins
- **Hook system** - Plugins can hook into system events
- **Middleware system** - Flexible middleware at global, plugin, and route levels

### 🚧 In Progress
- **Theme system** - Swappable themes with layout constraints
- **CLI tooling** - Generate plugins, pages, and components
- **Plugin configuration UI** - Admin interface for plugin settings

### 📋 Planned
- **Media management** - Upload and manage files
- **Page builder** - Visual page composition
- **Plugin marketplace** - Install community plugins

## Tech Stack

- **Go** - Fast, compiled backend
- **Chi Router** - Lightweight, idiomatic HTTP routing
- **Templ** - Type-safe HTML templating
- **HTMX** - HTML-over-the-wire interactivity
- **Alpine.js** - Lightweight reactive UI
- **Tailwind CSS** - Utility-first styling
- **Air** - Live reload for development
- **GORM** - Database ORM
- **JWT** - Token-based authentication

## Getting Started

### Prerequisites
```bash
# Install Go 1.21+
# Install Templ
go install github.com/a-h/templ/cmd/templ@latest
# Install Air (for hot reload)
go install github.com/air-verse/air@latest
```

### Quick Start
```bash
# Clone the repository
git clone https://github.com/btassone/obtura.git
cd obtura

# Install dependencies
go mod download

# Run with Make (recommended)
make setup    # Initial setup
make dev      # Start development server

# Or run manually
templ generate
air
```

Visit http://localhost:3000 (proxied through Air)

### Using Make Commands
```bash
make help          # Show all available commands
make dev           # Start dev server with hot reload
make build         # Build the application
make test          # Run tests
make lint          # Run linters
make db-migrate    # Run database migrations
make new-plugin name=MyPlugin  # Generate a new plugin
```

## Plugin Development

Obtura's plugin system is the core of its extensibility. Everything is a plugin!

### Plugin Types

1. **Basic Plugin** - Simple plugins for background tasks
2. **Routable Plugin** - Add HTTP routes and pages
3. **Service Plugin** - Provide services to other plugins
4. **Hookable Plugin** - Hook into system events
5. **Admin Plugin** - Add admin panel functionality

### Creating a Plugin

```go
package myplugin

import (
    "github.com/btassone/obtura/pkg/plugin"
)

type MyPlugin struct {
    plugin.BasePlugin
}

func New() *MyPlugin {
    return &MyPlugin{
        BasePlugin: plugin.BasePlugin{
            PluginID:      "com.example.myplugin",
            PluginName:    "My Plugin",
            PluginVersion: "1.0.0",
        },
    }
}

// Add routes (optional)
func (p *MyPlugin) Routes() []plugin.Route {
    return []plugin.Route{
        {
            Method:  http.MethodGet,
            Path:    "/my-page",
            Handler: http.HandlerFunc(p.handlePage),
        },
    }
}
```

See [Plugin Development Guide](examples/PLUGIN_DEVELOPMENT.md) for detailed documentation.

## Project Structure

```
obtura/
├── cmd/obtura/          # CLI application entry point
├── internal/            # Private application code
│   ├── admin/          # Admin interface
│   ├── config/         # Configuration
│   ├── database/       # Database layer
│   ├── middleware/     # HTTP middleware
│   ├── models/         # Data models
│   ├── server/         # HTTP server
│   └── types/          # Shared types
├── pkg/                # Public libraries
│   ├── database/       # Database utilities
│   └── plugin/         # Plugin SDK
├── plugins/            # Built-in plugins
│   ├── auth/          # Authentication plugin
│   └── hello/         # Example plugin
├── web/               # Web assets
│   ├── static/        # CSS, JS, images
│   └── templates/     # Templ components
├── examples/          # Example code
│   └── plugins/       # Plugin examples
└── docs/              # Documentation

## Current Status

### Working Features
- ✅ Basic plugin system with registration and lifecycle
- ✅ Admin panel with authentication
- ✅ Plugin management UI
- ✅ Database integration with migrations
- ✅ Hot reload development environment
- ✅ Templ + HTMX integration
- ✅ Multiple plugin types (routable, service, hookable, admin)
- ✅ Plugin configuration system

### Recent Updates
- Fixed import cycles by creating shared types package
- Fixed Chi router middleware ordering issues
- Added comprehensive plugin examples
- Created plugin development documentation
- Implemented plugin configuration UI in admin panel

```

## Documentation

- [Plugin Development Guide](examples/PLUGIN_DEVELOPMENT.md) - Comprehensive plugin creation guide
- [Architecture Overview](docs/diagrams/architecture.md) - System architecture and design
- [API Reference](pkg/plugin/) - Plugin SDK documentation
- [Examples](examples/plugins/) - Working plugin examples

## Contributing

This is currently a personal project in early development. Feel free to watch/star for updates.

### Development Guidelines
- Follow Go best practices and idioms
- Write tests for new functionality
- Update documentation for API changes
- Use meaningful commit messages

## License

TBD