package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Projects []Project `json:"projects"`
}

// Project represents a project in the configuration
type Project struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Tasks []Task `json:"tasks"`
}

// Task represents a task in the configuration
type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// LoadConfig loads the configuration from the config.json file
func LoadConfig() (*Config, error) {
	// Get the executable directory
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// Try to find config.json in different locations
	configPaths := []string{
		filepath.Join(execDir, "config.json"),
		"config.json",
		filepath.Join("..", "config.json"),
	}

	var configFile *os.File
	var configPath string

	for _, path := range configPaths {
		file, err := os.Open(path)
		if err == nil {
			configFile = file
			configPath = path
			break
		}
	}

	if configFile == nil {
		return nil, fmt.Errorf("config.json not found in any of the expected locations")
	}
	defer configFile.Close()

	fmt.Printf("Using config file: %s\n", configPath)

	var config Config
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config.json: %w", err)
	}

	return &config, nil
}

// GetProjectByName returns a project by its name
func (c *Config) GetProjectByName(name string) *Project {
	for _, project := range c.Projects {
		if project.Name == name {
			return &project
		}
	}
	return nil
}

// GetProjectByID returns a project by its ID
func (c *Config) GetProjectByID(id int) *Project {
	for _, project := range c.Projects {
		if project.ID == id {
			return &project
		}
	}
	return nil
}

// GetTaskByName returns a task by its name within a project
func (p *Project) GetTaskByName(name string) *Task {
	for _, task := range p.Tasks {
		if task.Name == name {
			return &task
		}
	}
	return nil
}

// GetTaskByID returns a task by its ID within a project
func (p *Project) GetTaskByID(id int) *Task {
	for _, task := range p.Tasks {
		if task.ID == id {
			return &task
		}
	}
	return nil
}
