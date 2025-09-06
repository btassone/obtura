package types

// PluginInfo represents plugin information for display
type PluginInfo struct {
	ID          string
	Name        string
	Description string
	Version     string
	Author      string
	IsActive    bool
	IsCore      bool
	HasConfig   bool
}