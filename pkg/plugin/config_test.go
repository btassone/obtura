package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryConfigStorage(t *testing.T) {
	storage := NewMemoryConfigStorage()

	// Test Save and Load
	config := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	err := storage.Save("test-plugin", config)
	require.NoError(t, err)

	loaded, err := storage.Load("test-plugin")
	require.NoError(t, err)
	assert.Equal(t, config, loaded)

	// Test Load non-existent
	_, err = storage.Load("non-existent")
	assert.Error(t, err)

	// Test Delete
	err = storage.Delete("test-plugin")
	require.NoError(t, err)

	_, err = storage.Load("test-plugin")
	assert.Error(t, err)

	// Test List
	configs := map[string]interface{}{
		"plugin1": map[string]interface{}{"a": 1},
		"plugin2": map[string]interface{}{"b": 2},
		"plugin3": map[string]interface{}{"c": 3},
	}

	for id, cfg := range configs {
		err = storage.Save(id, cfg)
		require.NoError(t, err)
	}

	list, err := storage.List()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"plugin1", "plugin2", "plugin3"}, list)
}

func TestJSONFileConfigStorage(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	storage, err := NewJSONFileConfigStorage(tmpDir)
	require.NoError(t, err)

	// Test Save and Load
	config := map[string]interface{}{
		"host":  "localhost",
		"port":  8080,
		"debug": true,
		"nested": map[string]interface{}{
			"level":  "info",
			"format": "json",
		},
	}

	err = storage.Save("test-plugin", config)
	require.NoError(t, err)

	// Verify file exists
	configPath := filepath.Join(tmpDir, "test-plugin.json")
	assert.FileExists(t, configPath)

	// Load config
	loaded, err := storage.Load("test-plugin")
	require.NoError(t, err)
	assert.Equal(t, config, loaded)

	// Test invalid plugin ID
	err = storage.Save("../../../evil", config)
	assert.Error(t, err)

	// Test Delete
	err = storage.Delete("test-plugin")
	require.NoError(t, err)
	assert.NoFileExists(t, configPath)

	// Test List
	plugins := []string{"plugin-a", "plugin-b", "plugin-c"}
	for _, id := range plugins {
		err = storage.Save(id, map[string]interface{}{"id": id})
		require.NoError(t, err)
	}

	list, err := storage.List()
	require.NoError(t, err)
	assert.ElementsMatch(t, plugins, list)
}

func TestConfigManager(t *testing.T) {
	storage := NewMemoryConfigStorage()
	manager := NewConfigManagerWithStorage(storage)

	// Test Set and Get
	config := map[string]interface{}{
		"apiKey":  "secret123",
		"timeout": 30,
	}

	manager.SetConfig("test-plugin", config)

	got := manager.GetConfig("test-plugin")
	assert.Equal(t, config, got)

	// Test non-existent config
	assert.Nil(t, manager.GetConfig("non-existent"))

	// Test Save and Load
	err := manager.SaveConfig("test-plugin")
	require.NoError(t, err)

	// Create new manager with same storage
	manager2 := NewConfigManagerWithStorage(storage)
	err = manager2.LoadConfig("test-plugin")
	require.NoError(t, err)

	got = manager2.GetConfig("test-plugin")
	assert.Equal(t, config, got)

	// Test SaveAll and LoadAll
	configs := map[string]interface{}{
		"plugin1": map[string]interface{}{"a": 1},
		"plugin2": map[string]interface{}{"b": 2},
	}

	for id, cfg := range configs {
		manager.SetConfig(id, cfg)
	}

	err = manager.SaveAll()
	require.NoError(t, err)

	manager3 := NewConfigManagerWithStorage(storage)
	err = manager3.LoadAll()
	require.NoError(t, err)

	for id, expected := range configs {
		got := manager3.GetConfig(id)
		assert.Equal(t, expected, got)
	}
}

func TestConfigSchema(t *testing.T) {
	storage := NewMemoryConfigStorage()
	manager := NewConfigManagerWithStorage(storage)

	// Register schema
	schema := &ConfigSchema{
		Fields: []ConfigField{
			{
				Name:     "apiKey",
				Type:     ConfigFieldTypeString,
				Required: true,
			},
			{
				Name:     "port",
				Type:     ConfigFieldTypeInt,
				Required: true,
				Min:      json.Number("1"),
				Max:      json.Number("65535"),
			},
			{
				Name:    "debug",
				Type:    ConfigFieldTypeBool,
				Default: false,
			},
		},
	}

	manager.RegisterSchema("test-plugin", schema)

	// Test valid config
	validConfig := map[string]interface{}{
		"apiKey": "secret",
		"port":   8080,
		"debug":  true,
	}

	err := manager.ValidateConfig("test-plugin", validConfig)
	require.NoError(t, err)

	// Test missing required field
	invalidConfig := map[string]interface{}{
		"port": 8080,
	}

	err = manager.ValidateConfig("test-plugin", invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "apiKey")

	// Test invalid port range
	invalidPort := map[string]interface{}{
		"apiKey": "secret",
		"port":   70000,
	}

	err = manager.ValidateConfig("test-plugin", invalidPort)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port")

	// Test wrong type
	wrongType := map[string]interface{}{
		"apiKey": "secret",
		"port":   "not-a-number",
	}

	err = manager.ValidateConfig("test-plugin", wrongType)
	assert.Error(t, err)
}

func TestGenerateSchemaFromStruct(t *testing.T) {
	type TestConfig struct {
		Host     string                 `json:"host" required:"true"`
		Port     int                    `json:"port" min:"1" max:"65535" default:"8080"`
		Debug    bool                   `json:"debug" default:"false"`
		Timeout  float64                `json:"timeout" min:"0.1" max:"60"`
		Features []string               `json:"features"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	config := &TestConfig{}
	schema := GenerateSchemaFromStruct(config)
	require.NotNil(t, schema)

	assert.Len(t, schema.Fields, 6)

	// Check host field
	hostField := findField(schema.Fields, "host")
	require.NotNil(t, hostField)
	assert.Equal(t, ConfigFieldTypeString, hostField.Type)
	assert.True(t, hostField.Required)

	// Check port field
	portField := findField(schema.Fields, "port")
	require.NotNil(t, portField)
	assert.Equal(t, ConfigFieldTypeInt, portField.Type)
	assert.Equal(t, json.Number("1"), portField.Min)
	assert.Equal(t, json.Number("65535"), portField.Max)
	assert.Equal(t, 8080, portField.Default)

	// Check debug field
	debugField := findField(schema.Fields, "debug")
	require.NotNil(t, debugField)
	assert.Equal(t, ConfigFieldTypeBool, debugField.Type)
	assert.Equal(t, false, debugField.Default)

	// Check timeout field
	timeoutField := findField(schema.Fields, "timeout")
	require.NotNil(t, timeoutField)
	assert.Equal(t, ConfigFieldTypeFloat, timeoutField.Type)

	// Check features field
	featuresField := findField(schema.Fields, "features")
	require.NotNil(t, featuresField)
	assert.Equal(t, ConfigFieldTypeArray, featuresField.Type)

	// Check metadata field
	metadataField := findField(schema.Fields, "metadata")
	require.NotNil(t, metadataField)
	assert.Equal(t, ConfigFieldTypeObject, metadataField.Type)
}

func TestValidateFieldValue(t *testing.T) {
	tests := []struct {
		name    string
		field   ConfigField
		value   interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid string",
			field: ConfigField{
				Name: "test",
				Type: ConfigFieldTypeString,
			},
			value:   "hello",
			wantErr: false,
		},
		{
			name: "required string missing",
			field: ConfigField{
				Name:     "test",
				Type:     ConfigFieldTypeString,
				Required: true,
			},
			value:   nil,
			wantErr: true,
			errMsg:  "is required",
		},
		{
			name: "string with pattern",
			field: ConfigField{
				Name:    "email",
				Type:    ConfigFieldTypeString,
				Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			value:   "test@example.com",
			wantErr: false,
		},
		{
			name: "string with invalid pattern",
			field: ConfigField{
				Name:    "email",
				Type:    ConfigFieldTypeString,
				Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			},
			value:   "not-an-email",
			wantErr: true,
			errMsg:  "does not match pattern",
		},
		{
			name: "int within range",
			field: ConfigField{
				Name: "port",
				Type: ConfigFieldTypeInt,
				Min:  json.Number("1"),
				Max:  json.Number("65535"),
			},
			value:   8080,
			wantErr: false,
		},
		{
			name: "int below min",
			field: ConfigField{
				Name: "port",
				Type: ConfigFieldTypeInt,
				Min:  json.Number("1"),
			},
			value:   0,
			wantErr: true,
			errMsg:  "is less than minimum",
		},
		{
			name: "int above max",
			field: ConfigField{
				Name: "port",
				Type: ConfigFieldTypeInt,
				Max:  json.Number("100"),
			},
			value:   200,
			wantErr: true,
			errMsg:  "is greater than maximum",
		},
		{
			name: "bool valid",
			field: ConfigField{
				Name: "enabled",
				Type: ConfigFieldTypeBool,
			},
			value:   true,
			wantErr: false,
		},
		{
			name: "array with options",
			field: ConfigField{
				Name:    "roles",
				Type:    ConfigFieldTypeArray,
				Options: []string{"admin", "user", "guest"},
			},
			value:   []interface{}{"admin", "user"},
			wantErr: false,
		},
		{
			name: "array with invalid option",
			field: ConfigField{
				Name:    "roles",
				Type:    ConfigFieldTypeArray,
				Options: []string{"admin", "user", "guest"},
			},
			value:   []interface{}{"admin", "superuser"},
			wantErr: true,
			errMsg:  "invalid option",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldValue(tt.field, tt.value)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Helper function to find a field by name
func findField(fields []ConfigField, name string) *ConfigField {
	for _, field := range fields {
		if field.Name == name {
			return &field
		}
	}
	return nil
}

func TestEncryptedConfigStorage(t *testing.T) {
	// This test would require implementing encryption functionality
	// For now, we'll skip it but leave a placeholder
	t.Skip("Encryption not yet implemented")
}

func TestConfigWatching(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	storage, err := NewJSONFileConfigStorage(tmpDir)
	require.NoError(t, err)

	manager := NewConfigManagerWithStorage(storage)

	// Save initial config
	initialConfig := map[string]interface{}{
		"value": "initial",
	}
	manager.SetConfig("test-plugin", initialConfig)
	err = manager.SaveConfig("test-plugin")
	require.NoError(t, err)

	// Simulate external change
	newConfig := map[string]interface{}{
		"value": "changed",
	}
	configPath := filepath.Join(tmpDir, "test-plugin.json")
	data, _ := json.MarshalIndent(newConfig, "", "  ")
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	// Reload config
	err = manager.LoadConfig("test-plugin")
	require.NoError(t, err)

	got := manager.GetConfig("test-plugin")
	assert.Equal(t, newConfig, got)
}

func TestConfigDefaults(t *testing.T) {
	manager := NewConfigManager()

	schema := &ConfigSchema{
		Fields: []ConfigField{
			{
				Name:    "host",
				Type:    ConfigFieldTypeString,
				Default: "localhost",
			},
			{
				Name:    "port",
				Type:    ConfigFieldTypeInt,
				Default: 8080,
			},
			{
				Name:    "debug",
				Type:    ConfigFieldTypeBool,
				Default: false,
			},
		},
	}

	manager.RegisterSchema("test-plugin", schema)

	// Apply defaults to empty config
	config := map[string]interface{}{}
	manager.ApplyDefaults("test-plugin", config)

	assert.Equal(t, "localhost", config["host"])
	assert.Equal(t, 8080, config["port"])
	assert.Equal(t, false, config["debug"])

	// Don't override existing values
	config2 := map[string]interface{}{
		"host": "example.com",
		"port": 9000,
	}
	manager.ApplyDefaults("test-plugin", config2)

	assert.Equal(t, "example.com", config2["host"])
	assert.Equal(t, 9000, config2["port"])
	assert.Equal(t, false, config2["debug"]) // Only this should be set to default
}
