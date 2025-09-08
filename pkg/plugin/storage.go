package plugin

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// ConfigStorage defines the interface for storing plugin configurations
type ConfigStorage interface {
	// Load retrieves configuration for a plugin
	Load(pluginID string) (map[string]interface{}, error)
	
	// Save stores configuration for a plugin
	Save(pluginID string, config map[string]interface{}) error
	
	// Delete removes configuration for a plugin
	Delete(pluginID string) error
	
	// List returns all stored plugin IDs
	List() ([]string, error)
}

// MemoryConfigStorage is an in-memory implementation of ConfigStorage
type MemoryConfigStorage struct {
	mu      sync.RWMutex
	configs map[string]map[string]interface{}
}

// NewMemoryConfigStorage creates a new memory-based config storage
func NewMemoryConfigStorage() ConfigStorage {
	return &MemoryConfigStorage{
		configs: make(map[string]map[string]interface{}),
	}
}

// Load retrieves configuration for a plugin
func (s *MemoryConfigStorage) Load(pluginID string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	config, exists := s.configs[pluginID]
	if !exists {
		return nil, fmt.Errorf("no configuration found for plugin: %s", pluginID)
	}
	
	// Return a copy to prevent external modifications
	configCopy := make(map[string]interface{})
	for k, v := range config {
		configCopy[k] = v
	}
	
	return configCopy, nil
}

// Save stores configuration for a plugin
func (s *MemoryConfigStorage) Save(pluginID string, config map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Store a copy to prevent external modifications
	configCopy := make(map[string]interface{})
	for k, v := range config {
		configCopy[k] = v
	}
	
	s.configs[pluginID] = configCopy
	return nil
}

// Delete removes configuration for a plugin
func (s *MemoryConfigStorage) Delete(pluginID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.configs, pluginID)
	return nil
}

// List returns all stored plugin IDs
func (s *MemoryConfigStorage) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	ids := make([]string, 0, len(s.configs))
	for id := range s.configs {
		ids = append(ids, id)
	}
	
	return ids, nil
}

// JSONFileConfigStorage stores configurations in JSON files
type JSONFileConfigStorage struct {
	basePath string
	mu       sync.RWMutex
}

// NewJSONFileConfigStorage creates a new JSON file-based config storage
func NewJSONFileConfigStorage(basePath string) (ConfigStorage, error) {
	// Ensure the base path exists
	if err := ensureDir(basePath); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	
	return &JSONFileConfigStorage{
		basePath: basePath,
	}, nil
}

// Load retrieves configuration from a JSON file
func (s *JSONFileConfigStorage) Load(pluginID string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	filePath := s.configPath(pluginID)
	data, err := readFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}
	
	return config, nil
}

// Save stores configuration to a JSON file
func (s *JSONFileConfigStorage) Save(pluginID string, config map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Validate plugin ID to prevent directory traversal
	if err := validatePluginID(pluginID); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	filePath := s.configPath(pluginID)
	if err := writeFile(filePath, data); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// Delete removes a configuration file
func (s *JSONFileConfigStorage) Delete(pluginID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	filePath := s.configPath(pluginID)
	if err := removeFile(filePath); err != nil {
		return fmt.Errorf("failed to delete config file: %w", err)
	}
	
	return nil
}

// List returns all stored plugin IDs
func (s *JSONFileConfigStorage) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	files, err := listFiles(s.basePath, ".json")
	if err != nil {
		return nil, fmt.Errorf("failed to list config files: %w", err)
	}
	
	ids := make([]string, 0, len(files))
	for _, file := range files {
		// Remove .json extension
		id := file[:len(file)-5]
		ids = append(ids, id)
	}
	
	return ids, nil
}

// configPath returns the file path for a plugin's configuration
func (s *JSONFileConfigStorage) configPath(pluginID string) string {
	return filepath.Join(s.basePath, pluginID+".json")
}

// validatePluginID validates a plugin ID to prevent directory traversal
func validatePluginID(pluginID string) error {
	if pluginID == "" {
		return fmt.Errorf("plugin ID cannot be empty")
	}
	
	// Check for directory traversal patterns
	if strings.Contains(pluginID, "..") || strings.Contains(pluginID, "/") || strings.Contains(pluginID, "\\") {
		return fmt.Errorf("invalid plugin ID: %s", pluginID)
	}
	
	return nil
}

// File system helpers (to be implemented based on OS)
var (
	ensureDir  = defaultEnsureDir
	readFile   = defaultReadFile
	writeFile  = defaultWriteFile
	removeFile = defaultRemoveFile
	listFiles  = defaultListFiles
	filepathHelper   = defaultFilepath{}
)