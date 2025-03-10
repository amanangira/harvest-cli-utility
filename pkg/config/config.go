package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Projects       []Project `json:"projects"`
	DefaultProject string    `json:"default_project,omitempty"`
	DefaultTask    string    `json:"default_task,omitempty"`
	HarvestAPI     APIConfig `json:"harvest_api"`
}

// APIConfig represents the Harvest API configuration
type APIConfig struct {
	AccountID string `json:"account_id"`
	Token     string `json:"token"`
	BaseURL   string `json:"base_url,omitempty"`
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
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Get the executable directory
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// Try to find config.json in different locations
	configPaths := []string{
		"config.json",                                  // Current directory
		filepath.Join(execDir, "config.json"),          // Executable directory
		filepath.Join(homeDir, ".harvest-config.json"), // User's home directory
		filepath.Join("..", "config.json"),             // Parent directory
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
		return nil, fmt.Errorf("config.json not found in any of the expected locations: %v", configPaths)
	}
	defer configFile.Close()

	fmt.Printf("Using config file: %s\n", configPath)

	var config Config
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config.json: %w", err)
	}

	// Set default base URL if not provided
	if config.HarvestAPI.BaseURL == "" {
		config.HarvestAPI.BaseURL = "https://api.harvestapp.com/v2"
	}

	return &config, nil
}

// GetProjectByName returns a project by its name
func (c *Config) GetProjectByName(name string) *Project {
	for i, project := range c.Projects {
		if project.Name == name {
			return &c.Projects[i]
		}
	}
	return nil
}

// GetProjectByID returns a project by its ID
func (c *Config) GetProjectByID(id int) *Project {
	for i, project := range c.Projects {
		if project.ID == id {
			return &c.Projects[i]
		}
	}
	return nil
}

// GetTaskByName returns a task by its name within a project
func (p *Project) GetTaskByName(name string) *Task {
	for i, task := range p.Tasks {
		if task.Name == name {
			return &p.Tasks[i]
		}
	}
	return nil
}

// GetTaskByID returns a task by its ID within a project
func (p *Project) GetTaskByID(id int) *Task {
	for i, task := range p.Tasks {
		if task.ID == id {
			return &p.Tasks[i]
		}
	}
	return nil
}

// GetDefaultProject returns the default project
func (c *Config) GetDefaultProject() *Project {
	if c.DefaultProject == "" {
		return nil
	}
	return c.GetProjectByName(c.DefaultProject)
}

// GetDefaultTask returns the default task for a project
func (c *Config) GetDefaultTask(project *Project) *Task {
	if c.DefaultTask == "" {
		return nil
	}

	return project.GetTaskByName(c.DefaultTask)
}
