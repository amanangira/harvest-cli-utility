package cmd

import (
	"fmt"
	"harvest-cli/pkg/config"
	"harvest-cli/pkg/harvest"
	"log"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// UpdateCmd returns the update command
func UpdateCmd() *cobra.Command {
	var interactive bool
	var date string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a time entry",
		Long: `Update a time entry.
By default, shows all time entries for today and lets you select one to update.
Use -i flag for interactive mode to select a time entry to update.
Use -d flag to specify a date (YYYY-MM-DD format) for time entry selection.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Load configuration
			var err error
			appConfig, err = config.LoadConfig()
			if err != nil {
				log.Fatalf("Failed to load configuration: %v", err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Create Harvest API client
			client := harvest.NewClient(&appConfig.HarvestAPI)

			// Parse the date if provided, otherwise use today
			var targetDate string
			if date != "" {
				// Validate date format
				_, err := time.Parse("2006-01-02", date)
				if err != nil {
					log.Fatalf("Invalid date format. Please use YYYY-MM-DD format: %v", err)
				}
				targetDate = date
			} else {
				// Use today's date
				targetDate = time.Now().Format("2006-01-02")

				// Confirm date with more intuitive options
				fmt.Printf("Using default date: %s\n", targetDate)

				dateOptions := []string{"Use this date", "Enter a different date"}
				datePrompt := promptui.Select{
					Label: "Select an option",
					Items: dateOptions,
				}

				dateIndex, _, err := datePrompt.Run()
				if err != nil {
					log.Fatalf("Prompt failed: %v", err)
				}

				if dateIndex == 1 {
					// User wants to enter a different date
					customDatePrompt := promptui.Prompt{
						Label:     "Enter date (YYYY-MM-DD)",
						Default:   targetDate,
						AllowEdit: true,
						Validate: func(input string) error {
							_, err := time.Parse("2006-01-02", input)
							return err
						},
					}

					targetDate, err = customDatePrompt.Run()
					if err != nil {
						log.Fatalf("Prompt failed: %v", err)
					}
				}
			}

			// Always use the interactive selection flow now
			handleTimeEntrySelection(client, targetDate)
		},
	}

	// Define flags
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Use interactive mode to select a time entry to update (deprecated, now the default behavior)")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Date in YYYY-MM-DD format (default: today)")

	return cmd
}

// handleTimeEntrySelection handles selecting a time entry to update
func handleTimeEntrySelection(client *harvest.Client, date string) {
	// Get time entries for the specified date
	params := map[string]string{
		"from": date,
		"to":   date,
	}

	fmt.Printf("Fetching time entries for %s...\n", date)
	timeEntries, err := client.GetTimeEntries(params)
	if err != nil {
		log.Fatalf("Failed to get time entries: %v", err)
	}

	if len(timeEntries) == 0 {
		fmt.Printf("No time entries found for %s\n", date)
		return
	}

	// Create a list of time entries for selection
	timeEntryOptions := make([]string, len(timeEntries))
	for i, entry := range timeEntries {
		hours, minutes := convertDecimalToHoursMinutes(entry.Hours)
		timeEntryOptions[i] = fmt.Sprintf("[%d] %s - %s (%02d:%02d) - %s",
			entry.ID,
			entry.Project.Name,
			entry.Task.Name,
			hours,
			minutes,
			entry.Notes)
	}

	prompt := promptui.Select{
		Label: "Select a time entry to update",
		Items: timeEntryOptions,
		Size:  10, // Show more items at once if available
	}

	index, _, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}

	selectedEntry := timeEntries[index]

	// Display selected time entry details
	hours, minutes := convertDecimalToHoursMinutes(selectedEntry.Hours)
	fmt.Println("\nSelected Time Entry Details:")
	fmt.Printf("ID: %d\n", selectedEntry.ID)
	fmt.Printf("Date: %s\n", selectedEntry.SpentDate)
	fmt.Printf("Project: %s\n", selectedEntry.Project.Name)
	fmt.Printf("Task: %s\n", selectedEntry.Task.Name)
	fmt.Printf("Hours: %02d:%02d\n", hours, minutes)
	if selectedEntry.Notes != "" {
		fmt.Printf("Notes: %s\n", selectedEntry.Notes)
	}

	// Confirm update with more intuitive options
	updateOptions := []string{"Update this time entry", "Cancel update"}
	updatePrompt := promptui.Select{
		Label: "What would you like to do?",
		Items: updateOptions,
	}

	updateIndex, _, err := updatePrompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}

	if updateIndex == 1 {
		fmt.Println("Update cancelled")
		return
	}

	// Update the time entry
	updateTimeEntry(client, &selectedEntry)
}

// updateTimeEntry updates a time entry with user input
func updateTimeEntry(client *harvest.Client, entry *harvest.TimeEntry) {
	// Create update request
	updateRequest := &harvest.TimeEntry{}

	// Prompt for date
	datePrompt := promptui.Prompt{
		Label:     "Date (YYYY-MM-DD)",
		Default:   entry.SpentDate,
		AllowEdit: true,
		Validate: func(input string) error {
			_, err := time.Parse("2006-01-02", input)
			return err
		},
	}

	dateResult, err := datePrompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}
	updateRequest.SpentDate = dateResult

	// Prompt for project
	projectNames := make([]string, len(appConfig.Projects))
	projectMap := make(map[string]int)

	for i, project := range appConfig.Projects {
		projectNames[i] = project.Name
		projectMap[project.Name] = project.ID
	}

	// Find current project name
	currentProjectName := entry.Project.Name

	projectPrompt := promptui.Select{
		Label: "Select Project",
		Items: projectNames,
		Size:  10,
	}

	// Try to set the default to the current project
	for i, name := range projectNames {
		if name == currentProjectName {
			projectPrompt.CursorPos = i
			break
		}
	}

	_, projectResult, err := projectPrompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}

	selectedProject := appConfig.GetProjectByName(projectResult)
	if selectedProject == nil {
		log.Fatalf("Project not found: %s", projectResult)
	}
	updateRequest.ProjectID = selectedProject.ID

	// Prompt for task
	taskNames := make([]string, len(selectedProject.Tasks))
	taskMap := make(map[string]int)

	for i, task := range selectedProject.Tasks {
		taskNames[i] = task.Name
		taskMap[task.Name] = task.ID
	}

	// Find current task name
	currentTaskName := entry.Task.Name

	taskPrompt := promptui.Select{
		Label: "Select Task",
		Items: taskNames,
	}

	// Try to set the default to the current task
	for i, name := range taskNames {
		if name == currentTaskName {
			taskPrompt.CursorPos = i
			break
		}
	}

	_, taskResult, err := taskPrompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}

	selectedTask := selectedProject.GetTaskByName(taskResult)
	if selectedTask == nil {
		log.Fatalf("Task not found: %s", taskResult)
	}
	updateRequest.TaskID = selectedTask.ID

	// Prompt for hours
	hours, minutes := convertDecimalToHoursMinutes(entry.Hours)
	currentTime := fmt.Sprintf("%02d:%02d", hours, minutes)

	timePrompt := promptui.Prompt{
		Label:     "Time (HH:MM)",
		Default:   currentTime,
		AllowEdit: true,
		Validate: func(input string) error {
			_, err := parseDuration(input)
			return err
		},
	}

	timeResult, err := timePrompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}

	timeValue, _ := parseDuration(timeResult)
	updateRequest.Hours = timeValue

	// Prompt for notes
	notesPrompt := promptui.Prompt{
		Label:     "Notes",
		Default:   entry.Notes,
		AllowEdit: true,
	}

	notesResult, err := notesPrompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}
	updateRequest.Notes = notesResult

	// Show summary of changes and confirm with more intuitive options
	fmt.Println("\nSummary of Changes:")
	fmt.Printf("Date: %s -> %s\n", entry.SpentDate, updateRequest.SpentDate)
	fmt.Printf("Project: %s -> %s\n", entry.Project.Name, projectResult)
	fmt.Printf("Task: %s -> %s\n", entry.Task.Name, taskResult)

	oldHours, oldMinutes := convertDecimalToHoursMinutes(entry.Hours)
	newHours, newMinutes := convertDecimalToHoursMinutes(updateRequest.Hours)
	fmt.Printf("Time: %02d:%02d -> %02d:%02d\n", oldHours, oldMinutes, newHours, newMinutes)

	if entry.Notes != updateRequest.Notes {
		fmt.Printf("Notes: %s -> %s\n", entry.Notes, updateRequest.Notes)
	}

	// Confirm update with more intuitive options
	confirmOptions := []string{"Save changes", "Cancel update"}
	confirmPrompt := promptui.Select{
		Label: "What would you like to do?",
		Items: confirmOptions,
	}

	confirmIndex, _, err := confirmPrompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}

	if confirmIndex == 1 {
		fmt.Println("Update cancelled")
		return
	}

	// Update the time entry
	updatedEntry, err := client.UpdateTimeEntry(entry.ID, updateRequest)
	if err != nil {
		log.Fatalf("Failed to update time entry: %v", err)
	}

	// Display updated time entry details
	hours, minutes = convertDecimalToHoursMinutes(updatedEntry.Hours)
	fmt.Println("\nTime Entry Updated Successfully:")
	fmt.Printf("ID: %d\n", updatedEntry.ID)
	fmt.Printf("Date: %s\n", updatedEntry.SpentDate)
	fmt.Printf("Project: %s\n", updatedEntry.Project.Name)
	fmt.Printf("Task: %s\n", updatedEntry.Task.Name)
	fmt.Printf("Hours: %02d:%02d\n", hours, minutes)
	if updatedEntry.Notes != "" {
		fmt.Printf("Notes: %s\n", updatedEntry.Notes)
	}
}
