package docs

// PackageDoc represents documentation for a package
type PackageDoc struct {
	Name        string
	ImportPath  string
	Path        string  // For templates
	Doc         string
	Types       []TypeDoc
	Functions   []FunctionDoc
	Methods     []MethodDoc  // For templates
	Constants   []ValueDoc
	Variables   []ValueDoc
}

// TypeDoc represents documentation for a type
type TypeDoc struct {
	Name        string
	Doc         string
	Methods     []FunctionDoc
	Fields      []FieldDoc
	Declaration string  // For templates
}

// FunctionDoc represents documentation for a function or method
type FunctionDoc struct {
	Name       string
	Doc        string
	Signature  string
	Parameters []ParamDoc
	Returns    []string
}

// MethodDoc represents documentation for a method
type MethodDoc struct {
	Name      string
	Receiver  string
	Doc       string
	Signature string
}

// FieldDoc represents documentation for a struct field
type FieldDoc struct {
	Name string
	Type string
	Doc  string
	Tag  string
}

// ParamDoc represents a function parameter
type ParamDoc struct {
	Name string
	Type string
}

// ValueDoc represents documentation for a constant or variable
type ValueDoc struct {
	Name  string
	Type  string
	Value string
	Doc   string
}