package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config represents the application configuration
type Config struct {
	Projects             []Project `json:"projects"`
	DefaultProject       string    `json:"default_project,omitempty"`
	DefaultTask          string    `json:"default_task,omitempty"`
	YearStartDate        string    `json:"year_start_date,omitempty"`        // Format: "MM-DD", defaults to "01-01" if not specified
	MonthlyCapacityHours float64   `json:"monthly_capacity_hours,omitempty"` // Default: 160 hours
	BillableTaskIDs      []int     `json:"billable_task_ids,omitempty"`      // IDs of tasks considered billable for utilization calculation
	HarvestAPI           APIConfig `json:"harvest_api"`
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

// GetYearStartDate returns the configured year start date or January 1st if not configured
func (c *Config) GetYearStartDate() (int, int, error) {
	if c.YearStartDate == "" {
		return 1, 1, nil // Default to January 1st
	}

	// Parse the MM-DD format
	parts := strings.Split(c.YearStartDate, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid year_start_date format: %s, expected MM-DD", c.YearStartDate)
	}

	month, err := strconv.Atoi(parts[0])
	if err != nil || month < 1 || month > 12 {
		return 0, 0, fmt.Errorf("invalid month in year_start_date: %s", parts[0])
	}

	day, err := strconv.Atoi(parts[1])
	if err != nil || day < 1 || day > 31 {
		return 0, 0, fmt.Errorf("invalid day in year_start_date: %s", parts[1])
	}

	return month, day, nil
}

// GetMonthlyCapacityHours returns the configured monthly capacity hours or default value of 160
func (c *Config) GetMonthlyCapacityHours() float64 {
	if c.MonthlyCapacityHours <= 0 {
		return 160.0 // Default monthly capacity is 160 hours
	}
	return c.MonthlyCapacityHours
}

// IsBillableTask checks if a task ID is in the list of billable task IDs
func (c *Config) IsBillableTask(taskID int) bool {
	// If no billable tasks are defined, consider all tasks billable
	if len(c.BillableTaskIDs) == 0 {
		return true
	}

	for _, id := range c.BillableTaskIDs {
		if id == taskID {
			return true
		}
	}
	return false
}
