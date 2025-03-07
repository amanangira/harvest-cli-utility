package cmd

import (
	"fmt"
	"harvest-cli/pkg/config"
	"harvest-cli/pkg/harvest"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// TimeEntry represents a time entry
type TimeEntry struct {
	Date      string
	ProjectID int
	TaskID    int
	Time      float64
}

// appConfig holds the application configuration
var appConfig *config.Config

// CreateCmd returns the create command
func CreateCmd() *cobra.Command {
	var useDefault bool
	var useDefaultMode bool
	var date, projectName, taskName string
	var timeValue string

	// Initialize the command
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new time entry",
		Long: `Create a new time entry with date, project, task, and time.
Example: h create -d 2023-03-06 -p "Corporate Visions | vPlaybook" --task "Software Development" -t 7.5
If arguments are not provided, you will be prompted for input.

Use -D flag for default mode, which uses default project and task from config.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Load configuration
			var err error
			appConfig, err = config.LoadConfig()
			if err != nil {
				log.Fatalf("Failed to load configuration: %v", err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			entry := TimeEntry{}

			// Handle default mode
			if useDefaultMode {
				// In default mode, we ignore other CLI arguments and use defaults from config
				handleDefaultMode(&entry)
			} else {
				// Regular mode - process arguments or prompt for input
				handleRegularMode(cmd, &entry, useDefault, date, projectName, taskName, timeValue)
			}

			// Create the time entry in Harvest
			createHarvestTimeEntry(&entry)
		},
	}

	// Define flags
	cmd.Flags().BoolVarP(&useDefaultMode, "default-mode", "D", false, "Use default mode (uses default project and task from config)")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Date in YYYY-MM-DD format (default: today)")
	cmd.Flags().StringVarP(&projectName, "project", "p", "", "Project")
	cmd.Flags().StringVarP(&taskName, "action", "a", "", "Action (Task)")
	cmd.Flags().StringVarP(&timeValue, "time", "t", "", "Duration in the following format (e.g., HH:MM)")

	return cmd
}

// handleDefaultMode handles the default mode for time entry creation
func handleDefaultMode(entry *TimeEntry) {
	// Set date to today
	entry.Date = time.Now().Format("2006-01-02")
	fmt.Printf("Using default date: %s\n", entry.Date)

	// Get default project from config
	defaultProject := appConfig.GetDefaultProject()
	if defaultProject == nil {
		log.Fatalf("No default project configured. Please set default_project in config.json")
	}
	entry.ProjectID = defaultProject.ID
	fmt.Printf("Using default project: %s (ID: %d)\n", defaultProject.Name, defaultProject.ID)

	// Get default task from config
	defaultTask := appConfig.GetDefaultTask(defaultProject)
	if defaultTask == nil {
		log.Fatalf("No default task configured. Please set default_task in config.json")
	}
	entry.TaskID = defaultTask.ID
	fmt.Printf("Using default task: %s (ID: %d)\n", defaultTask.Name, defaultTask.ID)

	// Prompt for time (always required)
	prompt := promptui.Prompt{
		Label: "Time (HH:MM)",
		Validate: func(input string) error {
			_, err := parseDuration(input)
			return err
		},
	}
	result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}
	entry.Time, _ = parseDuration(result)
}

// handleRegularMode handles the regular mode for time entry creation
func handleRegularMode(cmd *cobra.Command, entry *TimeEntry, useDefault bool, date, projectName, taskName, timeValue string) {
	var selectedProject *config.Project

	// Handle date
	if date != "" {
		entry.Date = date
	} else if useDefault {
		entry.Date = time.Now().Format("2006-01-02")
	} else {
		defaultDate := time.Now().Format("2006-01-02")
		prompt := promptui.Prompt{
			Label:     "Date (YYYY-MM-DD)",
			Default:   defaultDate,
			AllowEdit: true,
			Validate: func(input string) error {
				_, err := time.Parse("2006-01-02", input)
				return err
			},
		}
		result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed: %v\n", err)
			return
		}
		entry.Date = result
	}

	// Handle project selection
	if projectName != "" {
		// Find project by name
		selectedProject = appConfig.GetProjectByName(projectName)
		if selectedProject == nil {
			fmt.Printf("Project '%s' not found in configuration\n", projectName)
			return
		}
		entry.ProjectID = selectedProject.ID
	} else {
		// Create a list of project names for selection
		projectNames := make([]string, len(appConfig.Projects))
		for i, project := range appConfig.Projects {
			projectNames[i] = project.Name
		}

		prompt := promptui.Select{
			Label: "Select Project",
			Items: projectNames,
		}
		index, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed: %v\n", err)
			return
		}

		selectedProject = &appConfig.Projects[index]
		entry.ProjectID = selectedProject.ID
		projectName = result
	}

	// Handle task selection
	if taskName != "" {
		// Find task by name within the selected project
		task := selectedProject.GetTaskByName(taskName)
		if task == nil {
			fmt.Printf("Task '%s' not found in project '%s'\n", taskName, projectName)
			return
		}
		entry.TaskID = task.ID
	} else {
		// Create a list of task names for selection
		var taskNames []string
		for _, task := range selectedProject.Tasks {
			taskNames = append(taskNames, task.Name)
		}

		prompt := promptui.Select{
			Label: "Select Task",
			Items: taskNames,
		}
		index, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed: %v\n", err)
			return
		}

		entry.TaskID = selectedProject.Tasks[index].ID
		taskName = result
	}

	// Handle time
	if timeValue != "" {
		var err error
		entry.Time, err = parseDuration(timeValue)
		if err != nil {
			fmt.Printf("Invalid duration format: %v\n", err)
			return
		}
	} else {
		prompt := promptui.Prompt{
			Label: "Time (HH:MM)",
			Validate: func(input string) error {
				_, err := parseDuration(input)
				return err
			},
		}
		result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed: %v\n", err)
			return
		}
		entry.Time, _ = parseDuration(result)
	}

	// Output the final entry details
	hours, minutes := convertDecimalToHoursMinutes(entry.Time)
	fmt.Println("\nTime Entry Details:")
	fmt.Printf("Date: %s\n", entry.Date)
	fmt.Printf("Project ID: %d\n", entry.ProjectID)
	fmt.Printf("Task ID: %d\n", entry.TaskID)
	fmt.Printf("Time: %.2f hours (%02d:%02d)\n", entry.Time, hours, minutes)
}

// createHarvestTimeEntry creates a time entry in Harvest
func createHarvestTimeEntry(entry *TimeEntry) {
	// Create Harvest API client
	client := harvest.NewClient(&appConfig.HarvestAPI)

	// Create time entry request
	timeEntry := &harvest.TimeEntry{
		SpentDate: entry.Date,
		ProjectID: entry.ProjectID,
		TaskID:    entry.TaskID,
		Hours:     entry.Time,
	}

	// Send request to Harvest API
	fmt.Println("\nSending time entry to Harvest...")
	createdEntry, err := client.CreateTimeEntry(timeEntry)
	if err != nil {
		log.Fatalf("Failed to create time entry: %v", err)
	}

	// Output success message
	fmt.Println("\nTime Entry Created Successfully in Harvest!")
	fmt.Printf("Entry ID: %d\n", createdEntry.ID)
	fmt.Printf("Date: %s\n", createdEntry.SpentDate)
	fmt.Printf("Project ID: %d\n", createdEntry.ProjectID)
	fmt.Printf("Task ID: %d\n", createdEntry.TaskID)
	fmt.Printf("Hours: %.2f\n", createdEntry.Hours)
}

// convertDecimalToHoursMinutes converts decimal hours to hours and minutes
func convertDecimalToHoursMinutes(decimalHours float64) (int, int) {
	hours := int(decimalHours)
	minutes := int((decimalHours - float64(hours)) * 60)
	return hours, minutes
}

func parseDuration(duration string) (float64, error) {
	parts := strings.Split(duration, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid format, expected HH:MM")
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours value")
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes value")
	}

	if minutes < 0 || minutes >= 60 {
		return 0, fmt.Errorf("minutes must be between 0 and 59")
	}

	return float64(hours) + float64(minutes)/60, nil
}
