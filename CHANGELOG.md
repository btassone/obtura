# Changelog

All notable changes to the Obtura project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Documentation Plugin** - Automatically generates API documentation from Go source code comments
  - Scans packages and extracts doc comments
  - Provides searchable API reference at `/docs`
  - Admin interface for regenerating docs
  - Support for types, functions, methods, and constants

- **Plugin Examples** - Comprehensive examples demonstrating all plugin types
  - Basic plugin with lifecycle methods
  - Blog plugin showing routable functionality
  - Cache service plugin for inter-plugin services
  - SEO plugin demonstrating the hook system
  - Analytics plugin showing admin panel integration
  - Complete usage example in `main_example.go`

- **Plugin Development Guide** - Detailed documentation for plugin developers
  - Architecture overview
  - Step-by-step plugin creation
  - Best practices and patterns
  - Testing strategies

- **Shared Types Package** - `internal/types` for breaking import cycles
  - Moved `PluginInfo` type to shared location
  - Cleaner separation of concerns

### Fixed
- **Import Cycle** - Resolved circular dependency between admin and template packages
  - Created shared types package
  - Refactored imports to use shared types

- **Chi Middleware Ordering** - Fixed "all middlewares must be defined before routes" error
  - Added `SetRouter` method to plugin registry
  - Delayed route registration until after middleware setup
  - Routes now queued until router is available

### Changed
- **Plugin Registry** - Enhanced to support delayed route registration
  - Routes stored until router is set
  - Cleaner initialization flow
  - Better separation of concerns

- **Architecture Diagrams** - Updated to reflect current implementation
  - Added Documentation Plugin architecture
  - Added Plugin Registry architecture
  - Updated plugin types and lifecycle
  - Added detailed flow diagrams

## [0.1.0] - 2024-01-XX (Initial Development)

### Added
- **Core Plugin System** - Flexible plugin architecture
  - Plugin interface with lifecycle (Initialize, Start, Stop)
  - Multiple plugin types (Basic, Routable, Service, Hookable, Admin)
  - Plugin registry for central management
  - Configuration system for plugins

- **Admin Interface** - Built-in administration panel
  - JWT-based authentication
  - Plugin management UI
  - User management
  - Responsive design with Tailwind CSS

- **Authentication Plugin** - Secure user authentication
  - JWT token generation and validation
  - Login/logout functionality
  - Admin role support
  - Session management

- **Database Integration** - Modern database support
  - GORM ORM integration
  - Support for SQLite (dev) and PostgreSQL (prod)
  - Migration system
  - User model with secure password hashing

- **Development Environment** - Modern development workflow
  - Hot reload with Air
  - Templ for type-safe templates
  - HTMX for dynamic interactions
  - Alpine.js for reactive UI
  - Tailwind CSS for styling
  - Make commands for common tasks

- **HTTP Server** - Chi-based web server
  - Middleware support (global, plugin, route)
  - Clean URL routing
  - Static file serving
  - Request logging

- **Project Structure** - Go standard layout
  - `/cmd` - Application entry points
  - `/internal` - Private application code
  - `/pkg` - Public libraries
  - `/plugins` - Plugin implementations
  - `/web` - Templates and static assets
  - `/examples` - Example code
  - `/docs` - Documentation

### Infrastructure
- Go 1.21+ with modules
- Chi router for HTTP
- Templ for templating
- HTMX for interactivity
- Alpine.js for client state
- Tailwind CSS for styling
- GORM for database ORM
- Air for hot reload
- Make for build automation