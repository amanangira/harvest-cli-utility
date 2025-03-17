package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// ConfigCmd returns the config command
func ConfigCmd() *cobra.Command {
	var showSensitive bool

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Display configuration information",
		Long: `Display information about the configuration file being used.
Shows the path to the configuration file and its contents.
By default, sensitive information like API tokens are masked.
Use --show-sensitive flag to display all information including sensitive data.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Get the loaded config file path
			configPath, err := findLoadedConfigPath()
			if err != nil {
				log.Fatalf("Failed to determine config file path: %v", err)
			}

			fmt.Printf("Configuration file: %s\n\n", configPath)

			// Load and display the config
			displayConfig(configPath, showSensitive)
		},
	}

	// Define flags
	cmd.Flags().BoolVarP(&showSensitive, "show-sensitive", "s", false, "Show sensitive information like API tokens")

	return cmd
}

// findLoadedConfigPath determines which config file is being loaded
func findLoadedConfigPath() (string, error) {
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Get the executable directory
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// Try to find config.json in different locations
	configPaths := []string{
		"config.json",                                  // Current directory
		filepath.Join(execDir, "config.json"),          // Executable directory
		filepath.Join(homeDir, ".harvest-config.json"), // User's home directory
		filepath.Join("..", "config.json"),             // Parent directory
	}

	for _, path := range configPaths {
		_, err := os.Stat(path)
		if err == nil {
			// File exists, check if it's readable
			file, err := os.Open(path)
			if err == nil {
				file.Close()
				// This is the config file being used
				absPath, err := filepath.Abs(path)
				if err != nil {
					return path, nil // Return relative path if absolute fails
				}
				return absPath, nil
			}
		}
	}

	return "", fmt.Errorf("config.json not found in any of the expected locations: %v", configPaths)
}

// displayConfig reads and displays the configuration file
func displayConfig(configPath string, showSensitive bool) {
	// Read the config file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	// Parse the JSON
	var configMap map[string]interface{}
	if err := json.Unmarshal(configData, &configMap); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	// Mask sensitive information if needed
	if !showSensitive && configMap["harvest_api"] != nil {
		if harvestAPI, ok := configMap["harvest_api"].(map[string]interface{}); ok {
			if _, exists := harvestAPI["token"]; exists {
				harvestAPI["token"] = "********" // Mask the token
			}
		}
	}

	// Pretty print the config
	prettyJSON, err := json.MarshalIndent(configMap, "", "  ")
	if err != nil {
		log.Fatalf("Failed to format config: %v", err)
	}

	fmt.Println("Configuration contents:")
	fmt.Println("------------------------")
	fmt.Println(string(prettyJSON))
}
