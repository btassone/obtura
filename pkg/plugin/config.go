package plugin

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// ConfigManager manages plugin configurations
type ConfigManager struct {
	storage ConfigStorage
	schemas map[string]*ConfigSchema
	cache   map[string]interface{}
}

// NewConfigManager creates a new config manager
func NewConfigManager() *ConfigManager {
	return NewConfigManagerWithStorage(NewMemoryConfigStorage())
}

// NewConfigManagerWithStorage creates a new config manager with a specific storage backend
func NewConfigManagerWithStorage(storage ConfigStorage) *ConfigManager {
	return &ConfigManager{
		storage: storage,
		schemas: make(map[string]*ConfigSchema),
		cache:   make(map[string]interface{}),
	}
}

// ConfigSchema defines the structure of a plugin's configuration
type ConfigSchema struct {
	Fields []ConfigField `json:"fields"`
}

// ConfigField represents a single configuration field
type ConfigField struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // string, number, boolean, select, multiselect
	Label       string      `json:"label"`
	Description string      `json:"description"`
	Default     interface{} `json:"default"`
	Required    bool        `json:"required"`
	Options     []Option    `json:"options,omitempty"` // For select/multiselect
	Validation  *Validation `json:"validation,omitempty"`
}

// Option represents a select option
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Validation rules for a field
type Validation struct {
	Min    *float64 `json:"min,omitempty"`
	Max    *float64 `json:"max,omitempty"`
	MinLen *int     `json:"minLen,omitempty"`
	MaxLen *int     `json:"maxLen,omitempty"`
	Regex  string   `json:"regex,omitempty"`
}

// RegisterSchema registers a configuration schema for a plugin
func (cm *ConfigManager) RegisterSchema(pluginID string, schema *ConfigSchema) {
	cm.schemas[pluginID] = schema
}

// GetSchema returns the schema for a plugin
func (cm *ConfigManager) GetSchema(pluginID string) (*ConfigSchema, bool) {
	schema, ok := cm.schemas[pluginID]
	return schema, ok
}

// SetConfig sets configuration for a plugin
func (cm *ConfigManager) SetConfig(pluginID string, config interface{}) error {
	// Validate against schema if available
	if schema, ok := cm.schemas[pluginID]; ok {
		if err := cm.validateConfig(config, schema); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}
	}
	
	// Convert to map for storage
	configMap, err := cm.toMap(config)
	if err != nil {
		return fmt.Errorf("failed to convert config to map: %w", err)
	}
	
	// Save to storage
	if err := cm.storage.Save(pluginID, configMap); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	
	// Update cache
	cm.cache[pluginID] = config
	
	return nil
}

// GetConfig returns configuration for a plugin
func (cm *ConfigManager) GetConfig(pluginID string) (interface{}, bool) {
	// Check cache first
	if config, ok := cm.cache[pluginID]; ok {
		return config, true
	}
	
	// Load from storage
	configMap, err := cm.storage.Load(pluginID)
	if err != nil {
		return nil, false
	}
	
	// Cache and return
	cm.cache[pluginID] = configMap
	return configMap, true
}

// LoadConfig loads configuration into a struct
func (cm *ConfigManager) LoadConfig(pluginID string, target interface{}) error {
	config, ok := cm.GetConfig(pluginID)
	if !ok {
		return fmt.Errorf("no configuration found for plugin %s", pluginID)
	}
	
	// Convert to JSON and back to handle type conversions
	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(jsonData, target)
}

// SaveConfig saves configuration from a struct
func (cm *ConfigManager) SaveConfig(pluginID string, config interface{}) error {
	return cm.SetConfig(pluginID, config)
}

// validateConfig validates configuration against schema
func (cm *ConfigManager) validateConfig(config interface{}, schema *ConfigSchema) error {
	// Convert config to map for easier validation
	configMap, ok := config.(map[string]interface{})
	if !ok {
		// Try to convert struct to map
		jsonData, err := json.Marshal(config)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(jsonData, &configMap); err != nil {
			return err
		}
	}
	
	// Validate each field
	for _, field := range schema.Fields {
		value, exists := configMap[field.Name]
		
		// Check required fields
		if field.Required && (!exists || value == nil) {
			return fmt.Errorf("field %s is required", field.Name)
		}
		
		// Skip validation if field doesn't exist and isn't required
		if !exists {
			continue
		}
		
		// Validate field type and constraints
		if err := cm.validateField(value, field); err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}
	}
	
	return nil
}

// validateField validates a single field value
func (cm *ConfigManager) validateField(value interface{}, field ConfigField) error {
	// Type validation
	switch field.Type {
	case "string":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		if field.Validation != nil {
			if field.Validation.MinLen != nil && len(str) < *field.Validation.MinLen {
				return fmt.Errorf("minimum length is %d", *field.Validation.MinLen)
			}
			if field.Validation.MaxLen != nil && len(str) > *field.Validation.MaxLen {
				return fmt.Errorf("maximum length is %d", *field.Validation.MaxLen)
			}
			// TODO: Add regex validation
		}
		
	case "number":
		num, ok := toFloat64(value)
		if !ok {
			return fmt.Errorf("expected number, got %T", value)
		}
		if field.Validation != nil {
			if field.Validation.Min != nil && num < *field.Validation.Min {
				return fmt.Errorf("minimum value is %f", *field.Validation.Min)
			}
			if field.Validation.Max != nil && num > *field.Validation.Max {
				return fmt.Errorf("maximum value is %f", *field.Validation.Max)
			}
		}
		
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
		
	case "select":
		// Validate against options
		if !cm.isValidOption(value, field.Options) {
			return fmt.Errorf("invalid option")
		}
		
	case "multiselect":
		// Validate each selected option
		values, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array for multiselect")
		}
		for _, v := range values {
			if !cm.isValidOption(v, field.Options) {
				return fmt.Errorf("invalid option: %v", v)
			}
		}
	}
	
	return nil
}

// isValidOption checks if a value is in the options list
func (cm *ConfigManager) isValidOption(value interface{}, options []Option) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	
	for _, opt := range options {
		if opt.Value == str {
			return true
		}
	}
	return false
}

// toFloat64 converts various numeric types to float64
func toFloat64(value interface{}) (float64, bool) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), true
	}
	return 0, false
}

// GenerateSchemaFromStruct generates a config schema from a struct
func GenerateSchemaFromStruct(v interface{}) *ConfigSchema {
	schema := &ConfigSchema{
		Fields: []ConfigField{},
	}
	
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Get field metadata from tags
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		
		configField := ConfigField{
			Name:        jsonTag,
			Type:        getFieldType(field.Type),
			Label:       field.Tag.Get("label"),
			Description: field.Tag.Get("description"),
			Required:    field.Tag.Get("required") == "true",
		}
		
		// Set default from tag
		if defaultTag := field.Tag.Get("default"); defaultTag != "" {
			configField.Default = defaultTag
		}
		
		schema.Fields = append(schema.Fields, configField)
	}
	
	return schema
}

// getFieldType returns the config field type for a Go type
func getFieldType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice:
		return "multiselect"
	default:
		return "string"
	}
}

// toMap converts an interface to a map
func (cm *ConfigManager) toMap(v interface{}) (map[string]interface{}, error) {
	// If already a map, return it
	if m, ok := v.(map[string]interface{}); ok {
		return m, nil
	}
	
	// Convert via JSON
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	
	return m, nil
}