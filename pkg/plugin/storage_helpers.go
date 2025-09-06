package plugin

import (
	"os"
	"path/filepath"
	"strings"
)

// defaultFilepath provides path operations
type defaultFilepath struct{}

func (defaultFilepath) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// defaultEnsureDir creates a directory if it doesn't exist
func defaultEnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// defaultReadFile reads the entire file
func defaultReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// defaultWriteFile writes data to a file
func defaultWriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// defaultRemoveFile removes a file
func defaultRemoveFile(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil // File already doesn't exist
	}
	return err
}

// defaultListFiles lists files with a specific extension
func defaultListFiles(dir string, ext string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if strings.HasSuffix(name, ext) {
			files = append(files, name)
		}
	}
	
	return files, nil
}