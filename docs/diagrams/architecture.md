# Obtura Architecture Diagrams

## Overall System Architecture

```mermaid
graph TB
    subgraph "Frontend Layer"
        Browser[Client Browser]
        HTMX[HTMX]
        Alpine[Alpine.js]
        PWA[PWA Features]
        
        Browser --> HTMX
        Browser --> Alpine
        Browser --> PWA
    end

    subgraph "CLI Tool"
        CLI[obtura CLI]
        CLI --> Generate[Generate Commands]
        CLI --> Serve[Dev Server]
        CLI --> Build[Build Commands]
    end

    subgraph "Core Framework"
        Core[Core System]
        PM[Plugin Registry]
        Router[Chi Router]
        MW[Middleware Manager]
        Types[Shared Types]
        
        Core --> PM
        Core --> Router
        Core --> MW
        Core --> Types
    end

    subgraph "Web Layer"
        Server[HTTP Server]
        Middleware[Middleware Stack]
        Templates[Templ Components]
        Static[Static Assets]
        
        Server --> Middleware
        Middleware --> Templates
        Templates --> Static
    end

    subgraph "Database Layer"
        DB[Database Core]
        ORM[ORM/Models]
        Migrations[Migrations]
        Seeders[Seeders]
        
        DB --> ORM
        DB --> Migrations
        DB --> Seeders
    end

    subgraph "Plugin System"
        PI[Plugin Interface]
        CorePlugins[Core Plugins]
        UserPlugins[User Plugins]
        
        subgraph "Plugin Types"
            Basic[Basic Plugin]
            Routable[Routable Plugin]
            Service[Service Plugin]
            Hookable[Hookable Plugin]
            AdminPlugin[Admin Plugin]
        end
        
        PI --> CorePlugins
        PI --> UserPlugins
        PI --> Basic
        PI --> Routable
        PI --> Service
        PI --> Hookable
        PI --> AdminPlugin
    end

    subgraph "Admin Interface"
        Admin[Admin Dashboard]
        Settings[Settings Manager]
        PluginUI[Plugin UIs]
        
        Admin --> Settings
        Admin --> PluginUI
    end

    Browser --> Server
    HTMX --> Server
    Alpine --> HTMX
    CLI --> Core
    Core --> Server
    Core --> DB
    PM --> PI
    Router --> Middleware
    Middleware --> Router
    Layout --> Templates
    CorePlugins --> Admin
    UserPlugins --> Admin
    CorePlugins --> Middleware
    UserPlugins --> Middleware
    DBPlugin --> DB
    DBPlugin --> ORM
    Templates --> Browser
    Static --> Browser
    ORM --> CorePlugins
    ORM --> UserPlugins
```

## Plugin Architecture

```mermaid
graph LR
    subgraph "Plugin Interface"
        Interface[Plugin]
        Interface --> ID[ID()]
        Interface --> Name[Name()]
        Interface --> Version[Version()]
        Interface --> Initialize[Initialize()]
        Interface --> Start[Start()]
        Interface --> Stop[Stop()]
        Interface --> Config[Config()]
    end

    subgraph "Plugin Types"
        RoutablePlugin[RoutablePlugin]
        RoutablePlugin --> Routes[Routes()]
        
        ServicePlugin[ServicePlugin]
        ServicePlugin --> Service[Service()]
        
        HookablePlugin[HookablePlugin]
        HookablePlugin --> Hooks[Hooks()]
        
        AdminPlugin[AdminPlugin]
        AdminPlugin --> AdminRoutes[AdminRoutes()]
    end

    subgraph "Core Plugins"
        Auth[Auth Plugin]
        Docs[Documentation Plugin]
        Hello[Hello Plugin]
        Analytics[Analytics Plugin]
        SEO[SEO Plugin]
        Cache[Cache Plugin]
    end

    subgraph "Plugin Lifecycle"
        Register[Register Plugin]
        Init[Initialize]
        StartP[Start Plugin]
        Running[Running State]
        StopP[Stop Plugin]
        
        Register --> Init
        Init --> StartP
        StartP --> Running
        Running --> StopP
    end

    Interface --> RoutablePlugin
    Interface --> ServicePlugin
    Interface --> HookablePlugin
    Interface --> AdminPlugin
    
    Interface --> Auth
    Interface --> Docs
    Interface --> Hello
```

## Request Flow

```mermaid
sequenceDiagram
    participant Browser
    participant Server
    participant MW as Middleware Stack
    participant Router
    participant Plugin
    participant Layout
    participant Templ
    participant HTMX

    Browser->>Server: HTTP Request
    Server->>MW: Process Request
    MW->>MW: Global Middleware
    MW->>MW: Plugin Middleware
    MW->>MW: Route Middleware
    MW->>Router: Route Request
    Router->>Plugin: Find Handler
    Plugin->>Templ: Render Component
    
    alt Full Page Request
        Templ->>Layout: Wrap in Layout
        Layout->>MW: Full HTML Response
        MW->>MW: Route Middleware
        MW->>MW: Plugin Middleware
        MW->>MW: Global Middleware
        MW->>Browser: HTTP Response
    else HTMX Partial Request
        Templ->>HTMX: Partial HTML
        HTMX->>MW: Partial Response
        MW->>MW: Route Middleware
        MW->>MW: Plugin Middleware
        MW->>MW: Global Middleware
        MW->>Browser: Update DOM
    end
```

## Development Workflow

```mermaid
graph TD
    subgraph "Development Mode"
        Edit[Edit Code]
        Air[Air Detects Change]
        Rebuild[Rebuild Binary]
        Templ[Generate Templ]
        Reload[Hot Reload]
        
        Edit --> Air
        Air --> Rebuild
        Air --> Templ
        Rebuild --> Reload
        Templ --> Reload
    end

    subgraph "CLI Generation"
        CLICmd[obtura generate]
        CLICmd --> GenPage[Generate Page]
        CLICmd --> GenPlugin[Generate Plugin]
        CLICmd --> GenLayout[Generate Layout]
        CLICmd --> GenComp[Generate Component]
    end

    subgraph "Build Process"
        BuildCmd[obtura build]
        BuildCmd --> CompileGo[Compile Go]
        BuildCmd --> CompileTempl[Compile Templ]
        BuildCmd --> BuildCSS[Build Tailwind]
        BuildCmd --> Bundle[Bundle Assets]
    end
```

## Layout System

```mermaid
graph TB
    subgraph "Layout Components"
        Base[Base Layout Interface]
        Base --> Header[Header Component]
        Base --> Nav[Navigation Component]
        Base --> Content[Content Slot]
        Base --> Sidebar[Sidebar Component]
        Base --> Footer[Footer Component]
    end

    subgraph "Layout Registry"
        Registry[Layout Registry]
        Registry --> Default[Default Layout]
        Registry --> Admin[Admin Layout]
        Registry --> Custom[Custom Layouts]
    end

    subgraph "Page Rendering"
        Page[Page Component]
        Page --> SelectLayout[Select Layout]
        SelectLayout --> Registry
        Registry --> Render[Render with Layout]
    end
```

## Navigation System

```mermaid
graph LR
    subgraph "Navigation Registry"
        NavReg[Navigation Registry]
        NavReg --> MenuItems[Menu Items]
        NavReg --> Groups[Menu Groups]
        NavReg --> Order[Sort Order]
    end

    subgraph "Auto Registration"
        Plugin[Plugin]
        Plugin --> RegisterNav[Register Nav Items]
        RegisterNav --> NavReg
    end

    subgraph "Rendering"
        NavComp[Nav Component]
        NavComp --> GetItems[Get Menu Items]
        GetItems --> NavReg
        NavComp --> RenderMenu[Render Menu]
    end
```

## Themes Plugin Architecture

```mermaid
graph TB
    subgraph "Theme Plugin Core"
        ThemePlugin[Themes Plugin]
        ThemeManager[Theme Manager]
        ActiveTheme[Active Theme]
        
        ThemePlugin --> ThemeManager
        ThemeManager --> ActiveTheme
    end

    subgraph "Theme Components"
        Theme[Theme]
        Theme --> Layouts[Layouts Map]
        Theme --> Assets[Assets]
        Theme --> Settings[Theme Settings]
        Theme --> Constraints[Constraints]
        
        Layouts --> DefaultLayout[Default Layout]
        Layouts --> CustomLayouts[Custom Layouts]
        
        Assets --> Styles[CSS Files]
        Assets --> Scripts[JS Files]
        
        Settings --> Colors[Color Settings]
        Settings --> Typography[Typography]
        Settings --> Spacing[Spacing]
    end

    subgraph "Constraint System"
        Constraints --> RequiredSlots[Required Slots]
        Constraints --> OptionalSlots[Optional Slots]
        Constraints --> ColorSchemes[Color Schemes]
        Constraints --> Components[Required Components]
    end

    subgraph "Theme Application"
        PageRender[Page Render]
        PageRender --> GetTheme[Get Active Theme]
        GetTheme --> ThemeManager
        PageRender --> SelectLayout[Select Layout]
        SelectLayout --> Layouts
        PageRender --> LoadAssets[Load Theme Assets]
        LoadAssets --> Assets
        PageRender --> ApplyConstraints[Apply Constraints]
        ApplyConstraints --> Constraints
    end

    ThemeManager --> Theme
```

## Middleware Architecture

```mermaid
graph TB
    subgraph "Middleware Stack"
        Request[HTTP Request]
        MW1[Logger Middleware]
        MW2[Auth Middleware]
        MW3[CORS Middleware]
        MW4[Rate Limiter]
        MW5[Plugin Middleware]
        Handler[Route Handler]
        Response[HTTP Response]
        
        Request --> MW1
        MW1 --> MW2
        MW2 --> MW3
        MW3 --> MW4
        MW4 --> MW5
        MW5 --> Handler
        Handler --> Response
    end

    subgraph "Middleware Registry"
        Registry[Middleware Registry]
        Global[Global Middleware]
        Route[Route Middleware]
        Plugin[Plugin Middleware]
        
        Registry --> Global
        Registry --> Route
        Registry --> Plugin
    end

    subgraph "Middleware Interface"
        Interface[IMiddleware]
        Process[Process(next Handler)]
        Priority[Priority()]
        Name[Name()]
        
        Interface --> Process
        Interface --> Priority
        Interface --> Name
    end

    subgraph "Plugin Integration"
        PluginReg[Plugin Registration]
        PluginReg --> RegisterMW[Register Middleware]
        RegisterMW --> Registry
    end
```

## Middleware Types and Flow

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant GlobalMW as Global Middleware
    participant PluginMW as Plugin Middleware
    participant RouteMW as Route Middleware
    participant Handler
    participant DB

    Client->>Server: HTTP Request
    Server->>GlobalMW: Logger, CORS, Security
    
    alt Authenticated Route
        GlobalMW->>GlobalMW: Check Auth Token
        GlobalMW->>DB: Validate Session
    end
    
    GlobalMW->>PluginMW: Plugin-specific MW
    PluginMW->>RouteMW: Route-specific MW
    
    alt Rate Limited
        RouteMW->>RouteMW: Check Rate Limits
        RouteMW-->>Client: 429 Too Many Requests
    else Allowed
        RouteMW->>Handler: Process Request
        Handler->>DB: Business Logic
        DB->>Handler: Data
        Handler->>RouteMW: Response
        RouteMW->>PluginMW: Response
        PluginMW->>GlobalMW: Response
        GlobalMW->>Client: HTTP Response
    end
```

## Frontend Architecture

```mermaid
graph TB
    subgraph "Client Browser"
        HTML[HTML Document]
        HTMX[HTMX Library]
        Alpine[Alpine.js]
        TW[Tailwind CSS]
        
        HTML --> HTMX
        HTML --> Alpine
        HTML --> TW
    end

    subgraph "Frontend Components"
        Pages[Page Components]
        Partials[HTMX Partials]
        Forms[Form Components]
        Navigation[Nav Components]
        Modals[Modal Components]
        
        Pages --> Partials
        Pages --> Forms
        Pages --> Navigation
        Pages --> Modals
    end

    subgraph "Client State"
        LocalStorage[Local Storage]
        SessionStorage[Session Storage]
        URLState[URL State]
        HTMXState[HTMX State]
        
        Alpine --> LocalStorage
        Alpine --> SessionStorage
        HTMX --> URLState
        HTMX --> HTMXState
    end

    subgraph "Frontend Features"
        PWA[Progressive Web App]
        Offline[Offline Support]
        Push[Push Notifications]
        Theme[Theme Switching]
        
        PWA --> Offline
        PWA --> Push
        Alpine --> Theme
    end

    HTMX --> Pages
    Alpine --> Forms
    TW --> Pages
```

## Frontend-Backend Interaction

```mermaid
sequenceDiagram
    participant User
    participant Browser
    participant HTMX
    participant Alpine
    participant Server
    participant Templ
    participant DB

    User->>Browser: Interact with UI
    
    alt Form Submission
        Browser->>Alpine: Validate Form
        Alpine->>HTMX: Submit via HTMX
        HTMX->>Server: POST Request
        Server->>DB: Process Data
        DB->>Server: Result
        Server->>Templ: Render Response
        Templ->>HTMX: HTML Fragment
        HTMX->>Browser: Update DOM
        Browser->>Alpine: Trigger Events
    else Navigation
        Browser->>HTMX: Click Link
        HTMX->>Server: GET Request
        Server->>Templ: Render Page/Partial
        Templ->>HTMX: HTML Response
        HTMX->>Browser: Update DOM
        Browser->>Browser: Update URL
    else Real-time Updates
        Server->>Server: SSE Event
        Server->>HTMX: Server-Sent Event
        HTMX->>Browser: Update DOM
        Browser->>Alpine: Update State
    end
```

## Client-Side Component Architecture

```mermaid
graph LR
    subgraph "Templ Components"
        BaseLayout[Base Layout]
        PageComp[Page Components]
        SharedComp[Shared Components]
        
        BaseLayout --> PageComp
        PageComp --> SharedComp
    end

    subgraph "HTMX Features"
        Boost[hx-boost]
        Swap[hx-swap]
        Trigger[hx-trigger]
        Target[hx-target]
        SSE[hx-sse]
        WS[hx-ws]
        
        PageComp --> Boost
        PageComp --> Swap
        Forms --> Trigger
        Forms --> Target
        RealTime --> SSE
        Chat --> WS
    end

    subgraph "Alpine.js Features"
        Data[x-data]
        Show[x-show/x-if]
        Model[x-model]
        Store[Alpine.store]
        Magic[$refs/$el]
        
        Forms --> Data
        Modals --> Show
        Forms --> Model
        GlobalState --> Store
        Interactive --> Magic
    end

    subgraph "Progressive Enhancement"
        NoJS[No-JS Fallback]
        Loading[Loading States]
        Error[Error Handling]
        Offline[Offline Mode]
        
        HTMX --> Loading
        HTMX --> Error
        Alpine --> Offline
        BaseLayout --> NoJS
    end
```

## Database Architecture

```mermaid
graph TB
    subgraph "Database Layer"
        DBCore[Database Core]
        ConnPool[Connection Pool]
        QueryBuilder[Query Builder]
        Migrations[Migration System]
        Seeder[Database Seeder]
        
        DBCore --> ConnPool
        DBCore --> QueryBuilder
        DBCore --> Migrations
        DBCore --> Seeder
    end

    subgraph "ORM/Models"
        Model[Base Model]
        Relations[Relationships]
        Scopes[Query Scopes]
        Events[Model Events]
        Validation[Validation Rules]
        
        Model --> Relations
        Model --> Scopes
        Model --> Events
        Model --> Validation
    end

    subgraph "Migration Components"
        MigrationFiles[Migration Files]
        MigrationRunner[Migration Runner]
        SchemaBuilder[Schema Builder]
        Rollback[Rollback System]
        
        MigrationFiles --> MigrationRunner
        MigrationRunner --> SchemaBuilder
        MigrationRunner --> Rollback
    end

    subgraph "Seeding System"
        Seeders[Seeder Classes]
        Factories[Model Factories]
        Faker[Faker Data]
        SeedRunner[Seed Runner]
        
        Seeders --> Factories
        Factories --> Faker
        Seeders --> SeedRunner
    end

    subgraph "Database Drivers"
        PostgreSQL[PostgreSQL]
        MySQL[MySQL]
        SQLite[SQLite]
        MSSQL[SQL Server]
    end

    DBCore --> Model
    Migrations --> MigrationFiles
    Seeder --> Seeders
    ConnPool --> PostgreSQL
    ConnPool --> MySQL
    ConnPool --> SQLite
    ConnPool --> MSSQL
```

## MVC Architecture Pattern

```mermaid
graph LR
    subgraph "Model Layer"
        Models[Models]
        Repos[Repositories]
        Services[Services]
        DTOs[Data Transfer Objects]
        
        Models --> Repos
        Repos --> Services
        Services --> DTOs
    end

    subgraph "View Layer"
        Templates[Templ Components]
        Layouts[Layout System]
        Partials[Partial Views]
        ViewModels[View Models]
        
        Templates --> Layouts
        Templates --> Partials
        Templates --> ViewModels
    end

    subgraph "Controller Layer"
        Controllers[Controllers]
        Actions[Action Methods]
        Filters[Action Filters]
        Validators[Input Validators]
        
        Controllers --> Actions
        Actions --> Filters
        Actions --> Validators
    end

    subgraph "Request Flow"
        Request[HTTP Request]
        Router[Router]
        Middleware[Middleware]
        Response[HTTP Response]
        
        Request --> Router
        Router --> Middleware
        Middleware --> Controllers
    end

    Controllers --> Services
    Services --> Models
    Controllers --> ViewModels
    ViewModels --> Templates
    Templates --> Response
```

## Database Migration Flow

```mermaid
sequenceDiagram
    participant CLI
    participant MigrationCmd
    participant MigrationRunner
    participant SchemaBuilder
    participant Database
    participant MigrationTable

    CLI->>MigrationCmd: obtura migrate:up
    MigrationCmd->>MigrationRunner: Run pending migrations
    MigrationRunner->>MigrationTable: Check applied migrations
    MigrationTable->>MigrationRunner: List of applied
    
    loop For each pending migration
        MigrationRunner->>MigrationRunner: Load migration file
        MigrationRunner->>SchemaBuilder: Parse schema changes
        SchemaBuilder->>Database: Execute DDL statements
        Database->>SchemaBuilder: Success/Error
        
        alt Migration successful
            SchemaBuilder->>MigrationTable: Record migration
            MigrationTable->>MigrationRunner: Confirmed
        else Migration failed
            SchemaBuilder->>MigrationRunner: Rollback transaction
            MigrationRunner->>CLI: Report error
        end
    end
    
    MigrationRunner->>CLI: Migration complete
```

## Database Seeding Flow

```mermaid
sequenceDiagram
    participant CLI
    participant SeederCmd
    participant SeederRunner
    participant Factories
    participant Faker
    participant Database

    CLI->>SeederCmd: obtura db:seed
    SeederCmd->>SeederRunner: Run seeders
    
    loop For each seeder
        SeederRunner->>SeederRunner: Load seeder class
        SeederRunner->>Factories: Request model instances
        Factories->>Faker: Generate fake data
        Faker->>Factories: Random data
        Factories->>SeederRunner: Model instances
        SeederRunner->>Database: Insert records
        Database->>SeederRunner: Success/Error
    end
    
    SeederRunner->>CLI: Seeding complete
```

## Model Lifecycle and Events

```mermaid
graph TB
    subgraph "Model Lifecycle"
        Creating[Creating Event]
        Created[Created Event]
        Updating[Updating Event]
        Updated[Updated Event]
        Deleting[Deleting Event]
        Deleted[Deleted Event]
        
        Creating --> Created
        Updating --> Updated
        Deleting --> Deleted
    end

    subgraph "Event Handlers"
        Observers[Model Observers]
        Listeners[Event Listeners]
        Hooks[Lifecycle Hooks]
        
        Observers --> Listeners
        Listeners --> Hooks
    end

    subgraph "Common Use Cases"
        Validation[Pre-save Validation]
        Auditing[Audit Logging]
        Caching[Cache Invalidation]
        Notifications[Send Notifications]
        
        Creating --> Validation
        Updated --> Caching
        Created --> Notifications
        Deleted --> Auditing
    end
```

## Theme Switching Flow

```mermaid
sequenceDiagram
    participant Admin
    participant ThemePlugin
    participant ThemeManager
    participant Settings
    participant Cache
    participant Frontend

    Admin->>ThemePlugin: Select New Theme
    ThemePlugin->>ThemeManager: Validate Theme
    ThemeManager->>ThemeManager: Check Constraints
    
    alt Theme Valid
        ThemeManager->>Settings: Update Active Theme
        Settings->>Cache: Clear Theme Cache
        Cache->>Frontend: Trigger Reload
        Frontend->>ThemeManager: Load New Theme
        ThemeManager->>Frontend: Apply Theme Assets
    else Theme Invalid
        ThemeManager->>Admin: Show Constraint Errors
    end
```

## Documentation Plugin Architecture

```mermaid
graph TB
    subgraph "Documentation Plugin"
        DocsPlugin[Docs Plugin]
        Scanner[Package Scanner]
        Parser[Go AST Parser]
        Generator[Doc Generator]
        Search[Search Engine]
        
        DocsPlugin --> Scanner
        Scanner --> Parser
        Parser --> Generator
        Generator --> Search
    end

    subgraph "Documentation Sources"
        GoFiles[Go Source Files]
        Comments[Doc Comments]
        Packages[Package Structure]
        
        GoFiles --> Comments
        GoFiles --> Packages
    end

    subgraph "Documentation Types"
        PackageDocs[Package Docs]
        TypeDocs[Type Documentation]
        FunctionDocs[Function Docs]
        MethodDocs[Method Docs]
        
        PackageDocs --> TypeDocs
        TypeDocs --> MethodDocs
        PackageDocs --> FunctionDocs
    end

    subgraph "Output Routes"
        DocsIndex[/docs - Index Page]
        APIRef[/docs/api - API Reference]
        PackageView[/docs/api/{package}]
        SearchAPI[/docs/search]
        AdminDocs[/admin/docs]
    end

    Scanner --> GoFiles
    Parser --> Comments
    Generator --> PackageDocs
    DocsPlugin --> DocsIndex
    DocsPlugin --> APIRef
    DocsPlugin --> PackageView
    DocsPlugin --> SearchAPI
    DocsPlugin --> AdminDocs
```

## Documentation Generation Flow

```mermaid
sequenceDiagram
    participant Admin
    participant DocsPlugin
    participant Scanner
    participant Parser
    participant AST
    participant Storage
    participant UI

    Admin->>DocsPlugin: Regenerate Docs
    DocsPlugin->>Scanner: Scan Packages
    
    loop For each package
        Scanner->>Parser: Parse Go Files
        Parser->>AST: Build AST
        AST->>Parser: Extract Documentation
        Parser->>Storage: Store Package Docs
    end
    
    Storage->>DocsPlugin: Documentation Ready
    DocsPlugin->>UI: Render Documentation
    UI->>Admin: Display Success

    alt Search Request
        User->>DocsPlugin: Search Query
        DocsPlugin->>Storage: Query Docs
        Storage->>DocsPlugin: Results
        DocsPlugin->>UI: Render Results
    end
```

## Plugin Registry Architecture

```mermaid
graph TB
    subgraph "Plugin Registry"
        Registry[Registry Core]
        PluginMap[Plugin Map]
        ServiceMap[Service Map]
        HookMap[Hook Map]
        RouteQueue[Route Queue]
        
        Registry --> PluginMap
        Registry --> ServiceMap
        Registry --> HookMap
        Registry --> RouteQueue
    end

    subgraph "Registry Methods"
        Register[Register()]
        Get[Get()]
        List[List()]
        GetService[GetService()]
        TriggerHook[TriggerHook()]
        SetRouter[SetRouter()]
        
        Register --> PluginMap
        Get --> PluginMap
        GetService --> ServiceMap
        TriggerHook --> HookMap
        SetRouter --> RouteQueue
    end

    subgraph "Plugin States"
        Registered[Registered]
        Initialized[Initialized]
        Started[Started]
        Stopped[Stopped]
        
        Registered --> Initialized
        Initialized --> Started
        Started --> Stopped
    end

    subgraph "Route Registration"
        DelayedRoutes[Delayed Routes]
        RouterSet[Router Set]
        RoutesActive[Routes Active]
        
        DelayedRoutes --> RouterSet
        RouterSet --> RoutesActive
    end
```