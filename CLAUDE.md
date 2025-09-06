# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Obtura is a modular web/CLI framework inspired by Laravel, built with:
- **Go** - Backend language
- **Templ** - Type-safe templating for HTMX rendering
- **HTMX** - Dynamic HTML without JavaScript frameworks
- **Tailwind CSS** - Utility-first CSS framework
- **Air** - Live reload for development

The framework aims to provide WordPress-like flexibility with Laravel-like developer experience, where every component can be swapped out while maintaining consistent patterns.

## Development Commands

### Using Make (Recommended)

```bash
# Core commands
make help             # Show all available commands
make setup            # Initial project setup
make dev              # Start development server with hot reload
make build            # Build the application
make run              # Build and run
make test             # Run tests
make clean            # Clean build artifacts

# Code quality
make lint             # Run linters (fmt, vet, golangci-lint)
make test-coverage    # Generate test coverage report

# Database
make db-migrate       # Run pending migrations
make db-rollback      # Rollback last migration
make db-fresh         # Drop all tables, migrate, and seed
make db-seed          # Run database seeders

# Code generation
make new-migration name=create_users_table
make new-model name=User
make new-controller name=UserController
make new-plugin name=MyPlugin

# Production
make build-prod       # Build optimized binary
make build-all        # Build for all platforms

# Utilities
make deps             # Update Go dependencies
make tools            # Check required development tools
```

### Manual Commands

```bash
# Run development server with hot reload
air

# Build the project
go build -o ./tmp/main ./cmd/obtura

# Generate Templ files
templ generate

# Install dependencies
go mod tidy
```

## Architecture Philosophy

### Plugin-Based Architecture
Everything is a plugin, including core functionality:
- **Themes** - Controls visual presentation and layout constraints
- Pages
- Settings
- Blog functionality
- File uploads
- Admin interface components

### Page System
Pages should:
- Auto-register in navigation when created
- Support dynamic routing
- Be swappable at runtime
- Follow a consistent interface pattern

### Layout System
Layouts must be:
- Swappable (like WordPress themes)
- Consistent in their required components (header, footer, content areas)
- Built using Templ components
- HTMX-aware for partial updates

### Admin Interface
- Separate admin area (like WordPress)
- Each plugin provides its own admin section
- Consistent navigation pattern for plugin settings
- HTMX-powered for seamless interactions

## Key Interfaces and Patterns

### Core Interfaces (to be implemented)

```go
// Plugin interface - all functionality extends from this
type Plugin interface {
    Name() string
    Version() string
    Register(registry *PluginRegistry) error
    Routes() []Route
    AdminRoutes() []Route
    Assets() []Asset
    Middleware() []Middleware
}

// Layout interface - for swappable layouts
type Layout interface {
    Name() string
    Render(content templ.Component) templ.Component
    Slots() []string
}

// Page interface - for auto-registering pages
type Page interface {
    Path() string
    Title() string
    NavGroup() string
    NavOrder() int
    Layout() string
    Component() templ.Component
}

// Theme interface - for managing visual presentation
type Theme interface {
    Plugin // Themes are plugins too
    
    // Theme metadata
    Screenshot() string
    Description() string
    Author() string
    
    // Layout management
    Layouts() map[string]Layout
    DefaultLayout() string
    
    // Asset management
    Styles() []string
    Scripts() []string
    
    // Theme configuration
    Settings() map[string]ThemeSetting
    Constraints() ThemeConstraints
}

// ThemeConstraints defines what a theme requires/supports
type ThemeConstraints interface {
    RequiredSlots() []string     // e.g., ["header", "content", "footer"]
    OptionalSlots() []string     // e.g., ["sidebar", "hero"]
    ColorSchemes() []string      // e.g., ["light", "dark", "auto"]
    RequiredComponents() []string // e.g., ["navigation", "search"]
}

// Middleware interface - for request/response processing
type Middleware interface {
    Name() string
    Priority() int // Lower numbers run first
    Process(next http.Handler) http.Handler
}

// Model interface - base for all database models
type Model interface {
    TableName() string
    PrimaryKey() string
    Timestamps() bool
    FillableFields() []string
    GuardedFields() []string
    Validate() error
}

// Migration interface - for database schema changes
type Migration interface {
    Version() string
    Up(schema SchemaBuilder) error
    Down(schema SchemaBuilder) error
}

// Seeder interface - for populating database with test data
type Seeder interface {
    Run(db DatabaseConnection) error
    Priority() int
}
```

## Key Development Patterns

### Creating a New Page
Pages should follow this pattern:
1. Define the page as a plugin
2. Register routes automatically
3. Provide navigation metadata
4. Support both full page and HTMX partial rendering

### Creating a New Plugin
Plugins must:
1. Implement a standard plugin interface
2. Self-register their routes
3. Provide admin interface if needed
4. Be completely self-contained
5. Register middleware for their specific needs

### Plugin Examples
Complete working examples are available in `/examples/plugins/`:
- `basic-plugin.go` - Simple plugin with lifecycle
- `blog-plugin.go` - Routable plugin with pages
- `cache-service-plugin.go` - Service plugin example
- `seo-plugin.go` - Hookable plugin with system hooks
- `analytics-admin-plugin.go` - Admin panel integration
- `main_example.go` - How to use plugins together

See [Plugin Development Guide](examples/PLUGIN_DEVELOPMENT.md) for detailed documentation.

### Themes Plugin (Core Plugin)
The Themes plugin is a special core plugin that:
1. Manages all visual presentation aspects
2. Defines constraints that other plugins must follow
3. Provides theme switching functionality
4. Handles asset loading (CSS/JS)
5. Manages color schemes and visual settings
6. Ensures consistent slot usage across layouts

Theme features:
- **Multiple layouts** per theme (e.g., full-width, sidebar, grid)
- **Slot system** for consistent component placement
- **Asset management** for styles and scripts
- **Settings API** for customization (colors, fonts, spacing)
- **Constraints system** to ensure plugin compatibility
- **Live preview** in admin panel

### Templ + HTMX Integration
- Use Templ components for all UI
- Support both full page loads and HTMX partials
- Maintain progressive enhancement
- Keep state management server-side

### Middleware System
Middleware provides request/response processing at three levels:
1. **Global Middleware** - Applied to all requests (auth, logging, CORS)
2. **Plugin Middleware** - Applied to plugin-specific routes
3. **Route Middleware** - Applied to individual routes

Middleware execution order:
- Sorted by priority (lower numbers run first)
- Global → Plugin → Route → Handler → Route → Plugin → Global
- Each middleware can modify request/response or short-circuit the chain

Common middleware patterns:
- **Authentication** - Validate tokens, load user context
- **Authorization** - Check permissions for routes
- **Rate Limiting** - Prevent abuse
- **Request ID** - Track requests through the system
- **CORS** - Handle cross-origin requests
- **Compression** - Gzip responses
- **Caching** - Cache responses for performance

### Frontend Architecture

The frontend is built with progressive enhancement in mind:

#### Core Technologies
- **HTMX** - Dynamic HTML updates without full page refreshes
- **Alpine.js** - Lightweight reactive framework for UI interactions
- **Tailwind CSS** - Utility-first CSS framework
- **Templ** - Type-safe server-side templates

#### Frontend Patterns

**Component Structure:**
- Server-rendered Templ components
- HTMX for dynamic updates
- Alpine.js for client-side state
- Progressive enhancement fallbacks

**HTMX Usage:**
```html
<!-- Navigation with hx-boost -->
<nav hx-boost="true">
  <a href="/about">About</a>
</nav>

<!-- Form submission -->
<form hx-post="/api/contact" hx-target="#result">
  <input name="email" type="email" required>
  <button type="submit">Submit</button>
</form>

<!-- Live search -->
<input type="search" 
       hx-get="/api/search" 
       hx-trigger="keyup changed delay:300ms"
       hx-target="#search-results">

<!-- Server-sent events -->
<div hx-sse="connect:/events">
  <div hx-sse="swap:message"></div>
</div>
```

**Alpine.js Integration:**
```html
<!-- Form with validation -->
<form x-data="{ email: '', valid: false }">
  <input x-model="email" 
         @input="valid = $el.validity.valid">
  <button :disabled="!valid">Submit</button>
</form>

<!-- Modal component -->
<div x-data="{ open: false }">
  <button @click="open = true">Open Modal</button>
  <div x-show="open" x-trap="open">
    <!-- Modal content -->
  </div>
</div>

<!-- Global state -->
<script>
  Alpine.store('user', {
    name: '',
    authenticated: false
  })
</script>
```

**Progressive Web App Features:**
- Offline support with service workers
- Push notifications
- App manifest for installability
- Local storage for offline data

**Frontend Best Practices:**
1. Always provide no-JavaScript fallbacks
2. Use server-side rendering for initial page loads
3. Implement loading states for HTMX requests
4. Handle errors gracefully with user feedback
5. Optimize for mobile-first responsive design
6. Use semantic HTML for accessibility
7. Implement keyboard navigation support
8. Test with screen readers

### Database System

The database layer follows Laravel-inspired patterns with Go idioms:

#### Database Plugin (Core Plugin)
The Database plugin provides:
1. **Connection management** with pooling
2. **Query builder** for fluent database queries
3. **ORM** with active record pattern
4. **Migration system** for schema versioning
5. **Seeding system** for test data
6. **Multiple driver support** (PostgreSQL, MySQL, SQLite, SQL Server)

#### MVC Pattern Implementation

**Models:**
- Extend base model interface
- Define relationships (HasOne, HasMany, BelongsTo, BelongsToMany)
- Support query scopes for reusable queries
- Lifecycle events (creating, created, updating, updated, deleting, deleted)
- Built-in validation
- Mass assignment protection

```go
type User struct {
    obtura.Model
    ID        uint      `db:"id" primary:"true"`
    Name      string    `db:"name" validate:"required,min=3"`
    Email     string    `db:"email" validate:"required,email,unique"`
    Password  string    `db:"password" validate:"required,min=8" hidden:"true"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}

func (u *User) TableName() string {
    return "users"
}

func (u *User) Posts() HasMany {
    return u.HasMany(&Post{}, "user_id")
}
```

**Controllers:**
- RESTful resource controllers
- Action filters for common operations
- Dependency injection support
- Request validation
- Response helpers

```go
type UserController struct {
    db obtura.Database
}

func (c *UserController) Index(ctx obtura.Context) error {
    users := []User{}
    c.db.Model(&User{}).Paginate(ctx.Query("page", 1), 20).Find(&users)
    return ctx.Render(views.UserIndex(users))
}

func (c *UserController) Store(ctx obtura.Context) error {
    var input CreateUserRequest
    if err := ctx.Validate(&input); err != nil {
        return ctx.ValidationError(err)
    }
    
    user := User{
        Name:  input.Name,
        Email: input.Email,
    }
    
    if err := c.db.Create(&user); err != nil {
        return ctx.Error(err)
    }
    
    return ctx.Created(user)
}
```

**Views:**
- Templ components with view models
- Layouts with inheritance
- Partial views for reusability
- Form helpers with CSRF protection

#### Migration System

Migrations provide version control for your database schema:

```go
type CreateUsersTable struct{}

func (m *CreateUsersTable) Version() string {
    return "2024_01_01_000001"
}

func (m *CreateUsersTable) Up(schema obtura.Schema) error {
    return schema.Create("users", func(table *obtura.Blueprint) {
        table.ID()
        table.String("name")
        table.String("email").Unique()
        table.String("password")
        table.Timestamps()
        table.Index("email")
    })
}

func (m *CreateUsersTable) Down(schema obtura.Schema) error {
    return schema.Drop("users")
}
```

**Migration Commands:**
```bash
obtura make:migration create_users_table
obtura migrate:up
obtura migrate:down
obtura migrate:rollback --step=3
obtura migrate:fresh --seed
```

#### Seeding System

Seeders populate your database with test data:

```go
type UserSeeder struct{}

func (s *UserSeeder) Run(db obtura.Database) error {
    factory := db.Factory(&User{})
    
    // Create admin user
    admin := User{
        Name:  "Admin User",
        Email: "admin@example.com",
        Password: hash("password"),
    }
    db.Create(&admin)
    
    // Create 50 random users
    factory.Count(50).Create()
    
    return nil
}
```

**Model Factories:**
```go
func UserFactory() *obtura.Factory {
    return obtura.Define(&User{}, func(faker obtura.Faker) interface{} {
        return User{
            Name:     faker.Name(),
            Email:    faker.Email(),
            Password: hash("password"),
        }
    })
}
```

**Seeding Commands:**
```bash
obtura make:seeder UserSeeder
obtura db:seed
obtura db:seed --class=UserSeeder
```

#### Query Builder

Fluent interface for building database queries:

```go
// Simple queries
users := db.Table("users").Where("active", true).Get()

// Complex queries
posts := db.Table("posts").
    Join("users", "posts.user_id", "=", "users.id").
    Where("users.active", true).
    WhereIn("posts.status", []string{"published", "featured"}).
    OrderBy("posts.created_at", "desc").
    Limit(10).
    Get()

// Raw queries with bindings
results := db.Raw("SELECT * FROM users WHERE votes > ?", 100).Get()

// Transactions
db.Transaction(func(tx *obtura.Transaction) error {
    tx.Create(&order)
    tx.Update(&inventory)
    return nil
})
```

#### Model Relationships

Support for all common relationship types:

```go
// One to One
user.Profile()

// One to Many  
user.Posts()

// Many to Many
user.Roles()

// Has Many Through
country.Posts() // through users

// Polymorphic
post.Comments()
video.Comments()

// Eager loading
users := db.Model(&User{}).With("Posts", "Profile").Find(&users)

// Lazy eager loading
user.Load("Posts")
```

#### Model Events & Observers

React to model lifecycle events:

```go
type UserObserver struct{}

func (o *UserObserver) Creating(user *User) error {
    user.UUID = uuid.New()
    return nil
}

func (o *UserObserver) Created(user *User) error {
    go SendWelcomeEmail(user)
    return nil
}

func (o *UserObserver) Updating(user *User) error {
    if user.HasChanged("email") {
        user.EmailVerified = false
    }
    return nil
}

// Register observer
db.Model(&User{}).Observe(&UserObserver{})
```

## Project Structure (Following Go Standards)

```
/cmd/obtura/        - CLI application entry point
/internal/          - Private application code
  /core/            - Core framework functionality
    /plugin/        - Plugin system interfaces and registry
    /router/        - Dynamic routing system
    /layout/        - Layout management system
    /page/          - Page registration and management
  /admin/           - Admin interface implementation
  /server/          - HTTP server and middleware
  /config/          - Configuration management
/pkg/               - Public libraries that can be used by external apps
  /plugin/          - Plugin SDK for developers
  /templ/           - Templ utilities and helpers
/api/               - API definitions (OpenAPI/Swagger specs)
/web/               - Web assets and templates
  /templates/       - Templ components
  /static/          - CSS, JS, images
  /layouts/         - Layout templates
/plugins/           - Plugin implementations
  /core/            - Core plugins
    /themes/        - Themes management plugin
    /pages/         - Pages management plugin
    /settings/      - Settings plugin
  /community/       - Community plugins
/themes/            - Theme files
  /default/         - Default theme
  /admin/           - Admin theme
/build/             - Packaging and CI scripts
/scripts/           - Development scripts
/configs/           - Configuration file templates
/docs/              - Documentation
  /diagrams/        - Architecture diagrams
/examples/          - Example plugins and usage
/test/              - Integration tests
```

See the architecture diagram in [docs/diagrams/architecture.md](docs/diagrams/architecture.md)

## Current Status

### ✅ Completed Features

1. **Plugin System** - Fully functional plugin architecture with:
   - Basic plugins for background tasks
   - Routable plugins for adding pages/APIs
   - Service plugins for providing services
   - Hookable plugins for system events
   - Admin plugins for admin panel features
   - Plugin registry with lifecycle management
   - Delayed route registration to fix middleware ordering

2. **Admin Interface** - Working admin panel with:
   - JWT authentication system
   - Plugin management UI
   - Plugin configuration interface
   - User management
   - Responsive design with Tailwind

3. **Documentation Plugin** - Auto-generates docs from code comments:
   - Scans Go packages and extracts documentation
   - Provides searchable API reference
   - Admin interface for regeneration
   - Type, function, and method documentation

4. **Database Integration** - SQLite/PostgreSQL support with:
   - GORM ORM
   - Migration system
   - User model with authentication

5. **Development Environment**:
   - Hot reload with Air
   - Templ integration for type-safe templates
   - Make commands for common tasks

### 🚧 In Progress

1. **Theme System** - Swappable themes with constraints
2. **CLI Code Generation** - Commands to generate plugins, pages, etc.
3. **Media Management** - File upload and management plugin

### 📁 Recent Changes

1. **Fixed Import Cycles**: Created `internal/types` package for shared types
2. **Fixed Middleware Ordering**: Added `SetRouter` method to delay route registration
3. **Added Plugin Examples**: Comprehensive examples for all plugin types
4. **Added Documentation Plugin**: Generates API docs from code comments
5. **Updated Architecture Diagrams**: Added documentation and registry flows

## Development Guidelines

### Go Standards Compliance
- Follow [Go project layout standards](https://github.com/golang-standards/project-layout)
- `/internal/` contains private application code
- `/pkg/` contains public libraries for external use
- `/cmd/` contains application entry points
- `/api/` contains API specifications
- `/web/` contains web assets (templates, static files)

### Code Organization
- Use interfaces for all major components
- Keep packages small and focused
- Avoid circular dependencies
- Place shared types in dedicated packages
- Use dependency injection for testability

### When implementing new features:
- Start with the simplest working version
- Write interfaces before implementations
- Ensure everything can be generated via CLI commands
- Maintain flexibility for users to customize any component
- Keep the plugin interface minimal but sufficient
- Use HTMX for interactivity, avoid complex JavaScript
- Follow Go idioms and conventions
- Add appropriate error handling and logging