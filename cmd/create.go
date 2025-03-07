package cmd

import (
	"fmt"
	"harvest-cli/pkg/config"
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
	var date, projectName, taskName string
	var timeValue string

	// Initialize the command
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new time entry",
		Long: `Create a new time entry with date, project, task, and time.
Example: h create -d 2023-03-06 -p "Corporate Visions | vPlaybook" --task "Software Development" -t 7.5
If arguments are not provided, you will be prompted for input.`,
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
			var selectedProject *config.Project
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
				log.Println(taskName)
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

			// Output the final entry
			hours, minutes := convertDecimalToHoursMinutes(entry.Time)
			fmt.Println("\nTime Entry Created:")
			fmt.Printf("Date: %s\n", entry.Date)
			fmt.Printf("Project: %s (ID: %d)\n", projectName, entry.ProjectID)
			fmt.Printf("Task: %s (ID: %d)\n", taskName, entry.TaskID)
			fmt.Printf("Time: %.2f hours (%02d:%02d)\n", entry.Time, hours, minutes)
		},
	}

	// Define flags
	cmd.Flags().BoolVarP(&useDefault, "default", "", false, "Use default values for unspecified arguments")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Date in YYYY-MM-DD format (default: today)")
	cmd.Flags().StringVarP(&projectName, "project", "p", "", "Project")
	cmd.Flags().StringVarP(&taskName, "action", "a", "", "Action (Task)")
	cmd.Flags().StringVarP(&timeValue, "duration", "t", "", "Duration in the following format (e.g., HH:MM)")

	return cmd
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
