// Package docs provides automatic documentation generation from code comments
package docs

import (
	"context"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/a-h/templ"
	"github.com/btassone/obtura/pkg/plugin"
)

// Plugin generates documentation from Go source code comments
type Plugin struct {
	id          string
	name        string
	version     string
	description string
	author      string
	packages    map[string]*PackageDoc
}


// NewPlugin creates a new documentation plugin
func NewPlugin() *Plugin {
	return &Plugin{
		id:          "com.obtura.docs",
		name:        "Documentation Generator",
		version:     "1.0.0",
		description: "Automatically generates documentation from code comments",
		author:      "Obtura Team",
		packages:    make(map[string]*PackageDoc),
	}
}

// ID returns the plugin ID
func (p *Plugin) ID() string {
	return p.id
}

// Name returns the plugin name
func (p *Plugin) Name() string {
	return p.name
}

// Version returns the plugin version
func (p *Plugin) Version() string {
	return p.version
}

// Description returns the plugin description
func (p *Plugin) Description() string {
	return p.description
}

// Author returns the plugin author
func (p *Plugin) Author() string {
	return p.author
}

// Dependencies returns required plugin IDs
func (p *Plugin) Dependencies() []string {
	return []string{}
}

// ValidateConfig validates the plugin configuration
func (p *Plugin) ValidateConfig() error {
	return nil
}

// Destroy cleans up plugin resources
func (p *Plugin) Destroy(ctx context.Context) error {
	return nil
}

// Init sets up the documentation plugin
func (p *Plugin) Init(ctx context.Context) error {
	// Scan for Go packages
	if err := p.scanPackages(); err != nil {
		return fmt.Errorf("failed to scan packages: %w", err)
	}
	
	return nil
}

// PathToURLSegment converts a package path to a URL segment
func PathToURLSegment(path string) string {
	// Remove leading "./"
	path = strings.TrimPrefix(path, "./")
	// Replace "/" with "-"
	return strings.ReplaceAll(path, "/", "-")
}

// Routes returns the documentation routes
func (p *Plugin) Routes() []plugin.Route {
	return []plugin.Route{
		{
			Method:  http.MethodGet,
			Path:    "/docs",
			Handler: p.handleDocsIndex,
		},
		{
			Method:  http.MethodGet,
			Path:    "/docs/api",
			Handler: p.handleAPIReference,
		},
		{
			Method:  http.MethodGet,
			Path:    "/docs/api/{package}",
			Handler: p.handlePackageDoc,
		},
		{
			Method:  http.MethodGet,
			Path:    "/docs/search",
			Handler: p.handleSearch,
		},
	}
}

// AdminRoutes returns admin panel routes
func (p *Plugin) AdminRoutes() []plugin.Route {
	return []plugin.Route{
		{
			Method:  http.MethodGet,
			Path:    "/docs",
			Handler: p.handleAdminDocs,
		},
		{
			Method:  http.MethodPost,
			Path:    "/docs/regenerate",
			Handler: p.handleRegenerateDocs,
		},
	}
}

// AdminNavigation returns admin navigation items
func (p *Plugin) AdminNavigation() []plugin.NavItem {
	return []plugin.NavItem{
		{
			Title: "Documentation",
			Path:  "/admin/docs",
			Icon:  "document-text",
			Order: 100,
		},
	}
}

// scanPackages scans the codebase for Go packages
func (p *Plugin) scanPackages() error {
	// Define packages to scan
	packagesToScan := []string{
		"./pkg/plugin",
		"./pkg/database",
		"./internal/admin",
		"./internal/server",
		"./internal/middleware",
		"./plugins/auth",
		"./plugins/hello",
		"./plugins/docs",
	}
	
	for _, pkgPath := range packagesToScan {
		if err := p.parsePackage(pkgPath); err != nil {
			// Log error but continue scanning other packages
			fmt.Printf("Error parsing package %s: %v\n", pkgPath, err)
		}
	}
	
	return nil
}

// parsePackage parses a Go package and extracts documentation
func (p *Plugin) parsePackage(pkgPath string) error {
	absPath, err := filepath.Abs(pkgPath)
	if err != nil {
		return err
	}
	
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, absPath, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	
	for pkgName, pkg := range packages {
		// Skip test packages
		if strings.HasSuffix(pkgName, "_test") {
			continue
		}
		
		// Create package documentation
		docPkg := doc.New(pkg, pkgPath, 0)
		
		pkgDoc := &PackageDoc{
			Name:       pkgName,
			ImportPath: pkgPath,
			Doc:        docPkg.Doc,
		}
		
		// Extract types
		for _, t := range docPkg.Types {
			typeDoc := TypeDoc{
				Name: t.Name,
				Doc:  t.Doc,
			}
			
			// Extract methods
			for _, m := range t.Methods {
				methodDoc := FunctionDoc{
					Name:      m.Name,
					Doc:       m.Doc,
					Signature: p.getFunctionSignature(m.Decl),
				}
				typeDoc.Methods = append(typeDoc.Methods, methodDoc)
			}
			
			// Extract fields for structs
			if t.Decl != nil && t.Decl.Specs != nil {
				for _, spec := range t.Decl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							typeDoc.Fields = p.extractFields(structType)
						}
					}
				}
			}
			
			pkgDoc.Types = append(pkgDoc.Types, typeDoc)
		}
		
		// Extract functions
		for _, f := range docPkg.Funcs {
			funcDoc := FunctionDoc{
				Name:      f.Name,
				Doc:       f.Doc,
				Signature: p.getFunctionSignature(f.Decl),
			}
			pkgDoc.Functions = append(pkgDoc.Functions, funcDoc)
		}
		
		// Extract constants
		for _, c := range docPkg.Consts {
			for _, name := range c.Names {
				constDoc := ValueDoc{
					Name: name,
					Doc:  c.Doc,
				}
				pkgDoc.Constants = append(pkgDoc.Constants, constDoc)
			}
		}
		
		// Extract variables
		for _, v := range docPkg.Vars {
			for _, name := range v.Names {
				varDoc := ValueDoc{
					Name: name,
					Doc:  v.Doc,
				}
				pkgDoc.Variables = append(pkgDoc.Variables, varDoc)
			}
		}
		
		// Store package documentation
		p.packages[pkgPath] = pkgDoc
	}
	
	return nil
}

// extractFields extracts field documentation from a struct
func (p *Plugin) extractFields(structType *ast.StructType) []FieldDoc {
	var fields []FieldDoc
	
	for _, field := range structType.Fields.List {
		// Skip unexported fields
		if len(field.Names) > 0 && !ast.IsExported(field.Names[0].Name) {
			continue
		}
		
		var fieldDoc FieldDoc
		if len(field.Names) > 0 {
			fieldDoc.Name = field.Names[0].Name
		}
		
		// Get field type
		if field.Type != nil {
			fieldDoc.Type = fmt.Sprintf("%v", field.Type)
		}
		
		// Get field tag
		if field.Tag != nil {
			fieldDoc.Tag = field.Tag.Value
		}
		
		// Get field comment
		if field.Comment != nil {
			fieldDoc.Doc = field.Comment.Text()
		}
		
		fields = append(fields, fieldDoc)
	}
	
	return fields
}

// getFunctionSignature extracts function signature
func (p *Plugin) getFunctionSignature(fn *ast.FuncDecl) string {
	if fn == nil {
		return ""
	}
	
	var sig strings.Builder
	sig.WriteString("func ")
	
	// Add receiver if method
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		sig.WriteString("(")
		// Simplified receiver representation
		sig.WriteString("...)")
		sig.WriteString(" ")
	}
	
	sig.WriteString(fn.Name.Name)
	sig.WriteString("(")
	
	// Add parameters
	if fn.Type.Params != nil {
		params := []string{}
		for _, param := range fn.Type.Params.List {
			paramStr := fmt.Sprintf("%v", param.Type)
			params = append(params, paramStr)
		}
		sig.WriteString(strings.Join(params, ", "))
	}
	
	sig.WriteString(")")
	
	// Add return types
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		sig.WriteString(" ")
		if len(fn.Type.Results.List) > 1 {
			sig.WriteString("(")
		}
		
		results := []string{}
		for _, result := range fn.Type.Results.List {
			resultStr := fmt.Sprintf("%v", result.Type)
			results = append(results, resultStr)
		}
		sig.WriteString(strings.Join(results, ", "))
		
		if len(fn.Type.Results.List) > 1 {
			sig.WriteString(")")
		}
	}
	
	return sig.String()
}

// handleDocsIndex shows the documentation index
func (p *Plugin) handleDocsIndex(w http.ResponseWriter, r *http.Request) {
	component := docsIndexPage(p.packages)
	templ.Handler(component).ServeHTTP(w, r)
}

// handleAPIReference shows the API reference
func (p *Plugin) handleAPIReference(w http.ResponseWriter, r *http.Request) {
	// Sort packages by name
	var packages []*PackageDoc
	for _, pkg := range p.packages {
		packages = append(packages, pkg)
	}
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})
	
	component := apiReferencePage(packages)
	templ.Handler(component).ServeHTTP(w, r)
}

// handlePackageDoc shows documentation for a specific package
func (p *Plugin) handlePackageDoc(w http.ResponseWriter, r *http.Request) {
	pkgPath := r.PathValue("package")
	pkgPath = "./" + strings.ReplaceAll(pkgPath, "-", "/")
	
	pkg, ok := p.packages[pkgPath]
	if !ok {
		http.Error(w, "Package not found", http.StatusNotFound)
		return
	}
	
	component := packageDocPage(pkg)
	templ.Handler(component).ServeHTTP(w, r)
}

// handleSearch handles documentation search
func (p *Plugin) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
		return
	}
	
	// Simple search implementation
	var results []map[string]string
	query = strings.ToLower(query)
	
	for _, pkg := range p.packages {
		// Search in package name and doc
		if strings.Contains(strings.ToLower(pkg.Name), query) || 
		   strings.Contains(strings.ToLower(pkg.Doc), query) {
			results = append(results, map[string]string{
				"type":        "package",
				"name":        pkg.Name,
				"description": pkg.Doc,
				"url":         fmt.Sprintf("/docs/api/%s", strings.ReplaceAll(pkg.ImportPath[2:], "/", "-")),
			})
		}
		
		// Search in types
		for _, t := range pkg.Types {
			if strings.Contains(strings.ToLower(t.Name), query) ||
			   strings.Contains(strings.ToLower(t.Doc), query) {
				results = append(results, map[string]string{
					"type":        "type",
					"name":        fmt.Sprintf("%s.%s", pkg.Name, t.Name),
					"description": t.Doc,
					"url":         fmt.Sprintf("/docs/api/%s#%s", strings.ReplaceAll(pkg.ImportPath[2:], "/", "-"), t.Name),
				})
			}
		}
		
		// Search in functions
		for _, f := range pkg.Functions {
			if strings.Contains(strings.ToLower(f.Name), query) ||
			   strings.Contains(strings.ToLower(f.Doc), query) {
				results = append(results, map[string]string{
					"type":        "function",
					"name":        fmt.Sprintf("%s.%s", pkg.Name, f.Name),
					"description": f.Doc,
					"url":         fmt.Sprintf("/docs/api/%s#%s", strings.ReplaceAll(pkg.ImportPath[2:], "/", "-"), f.Name),
				})
			}
		}
	}
	
	// Return JSON results
	w.Header().Set("Content-Type", "application/json")
	// Simple JSON encoding
	var jsonResults []string
	for _, r := range results {
		jsonResults = append(jsonResults, fmt.Sprintf(`{"type":"%s","name":"%s","description":"%s","url":"%s"}`,
			r["type"], r["name"], strings.ReplaceAll(r["description"], "\"", "'"), r["url"]))
	}
	w.Write([]byte("[" + strings.Join(jsonResults, ",") + "]"))
}

// handleAdminDocs shows documentation management in admin
func (p *Plugin) handleAdminDocs(w http.ResponseWriter, r *http.Request) {
	stats := map[string]int{
		"packages":  len(p.packages),
		"types":     0,
		"functions": 0,
	}
	
	for _, pkg := range p.packages {
		stats["types"] += len(pkg.Types)
		stats["functions"] += len(pkg.Functions)
	}
	
	component := docsAdminPage(stats)
	templ.Handler(component).ServeHTTP(w, r)
}

// handleRegenerateDocs regenerates all documentation
func (p *Plugin) handleRegenerateDocs(w http.ResponseWriter, r *http.Request) {
	// Clear existing docs
	p.packages = make(map[string]*PackageDoc)
	
	// Rescan packages
	if err := p.scanPackages(); err != nil {
		http.Error(w, "Failed to regenerate documentation", http.StatusInternalServerError)
		return
	}
	
	// Redirect back to admin
	http.Redirect(w, r, "/admin/docs", http.StatusSeeOther)
}

// Start begins the documentation plugin
func (p *Plugin) Start(ctx context.Context) error {
	return nil
}

// Stop halts the documentation plugin
func (p *Plugin) Stop(ctx context.Context) error {
	return nil
}

// Config returns plugin configuration
func (p *Plugin) Config() interface{} {
	return struct {
		AutoRegenerate bool     `json:"auto_regenerate" schema:"title:Auto Regenerate,description:Automatically regenerate docs on file changes"`
		IncludePrivate bool     `json:"include_private" schema:"title:Include Private,description:Include unexported types and functions"`
		PackagePaths   []string `json:"package_paths" schema:"title:Package Paths,description:Additional package paths to scan"`
	}{
		AutoRegenerate: false,
		IncludePrivate: false,
		PackagePaths:   []string{},
	}
}

// DefaultConfig returns default configuration
func (p *Plugin) DefaultConfig() interface{} {
	return p.Config()
}

