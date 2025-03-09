package cmd

import (
	"fmt"
	"harvest-cli/pkg/config"
	"harvest-cli/pkg/harvest"
	"log"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// ProjectSummary represents a summary of time entries for a project
type ProjectSummary struct {
	ProjectName   string
	TaskSummaries map[string]float64
	TotalHours    float64
}

// ListCmd returns the list command
func ListCmd() *cobra.Command {
	var monthly, weekly bool
	var date string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time entries",
		Long: `List time entries for a specific day, week, or month.
By default, lists all time entries for the current day.
Use -d flag to specify a date (YYYY-MM-DD format).
Use -w flag for weekly summary.
Use -m flag for monthly summary.`,
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

			// Parse the date if provided
			var targetDate time.Time
			var err error
			if date != "" {
				targetDate, err = time.Parse("2006-01-02", date)
				if err != nil {
					log.Fatalf("Invalid date format. Please use YYYY-MM-DD format: %v", err)
				}
			} else {
				targetDate = time.Now()
			}

			if monthly {
				// Monthly summary
				handleMonthlySummary(client, targetDate)
			} else if weekly {
				// Weekly summary
				handleWeeklySummary(client, targetDate)
			} else {
				// Daily list
				handleDailyList(client, targetDate.Format("2006-01-02"))
			}
		},
	}

	// Define flags
	cmd.Flags().BoolVarP(&monthly, "monthly", "m", false, "Show monthly summary")
	cmd.Flags().BoolVarP(&weekly, "weekly", "w", false, "Show weekly summary")
	cmd.Flags().StringVarP(&date, "date", "d", "", "Date in YYYY-MM-DD format (default: today)")

	return cmd
}

// handleDailyList handles listing time entries for a specific day
func handleDailyList(client *harvest.Client, date string) {
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

	// Display time entries in a table format
	fmt.Printf("\nTime Entries for %s:\n", date)

	// Create a new tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print table header
	fmt.Fprintln(w, "ID\tProject (ID) | Task (ID)\tNotes\tDuration")
	fmt.Fprintln(w, "----\t------------------------\t--------------------\t--------")

	var totalHours float64

	for _, entry := range timeEntries {
		hours, minutes := convertDecimalToHoursMinutes(entry.Hours)
		projectTaskInfo := fmt.Sprintf("%s (%d) | %s (%d)",
			entry.Project.Name,
			entry.Project.ID,
			entry.Task.Name,
			entry.Task.ID)

		// Truncate notes if too long
		notes := entry.Notes
		if len(notes) > 30 {
			notes = notes[:27] + "..."
		}

		// Format duration
		duration := fmt.Sprintf("%02d:%02d", hours, minutes)

		// Print table row
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
			entry.ID,
			projectTaskInfo,
			notes,
			duration)

		totalHours += entry.Hours
	}

	// Flush the tabwriter
	w.Flush()

	// Print total
	totalHoursInt, totalMinutes := convertDecimalToHoursMinutes(totalHours)
	fmt.Printf("\nTotal: %02d:%02d hours\n", totalHoursInt, totalMinutes)
}

// handleWeeklySummary handles showing a weekly summary of time entries
func handleWeeklySummary(client *harvest.Client, targetDate time.Time) {
	// Calculate the start of the week (Monday)
	weekday := targetDate.Weekday()
	if weekday == 0 { // Sunday
		weekday = 7
	}
	startOfWeek := targetDate.AddDate(0, 0, -int(weekday-1))

	// Initialize with the specified week
	showWeeklySummary(client, startOfWeek)
}

// showWeeklySummary shows a summary for a specific week
func showWeeklySummary(client *harvest.Client, startDate time.Time) {
	// Calculate the end of the week (Sunday)
	endDate := startDate.AddDate(0, 0, 6)

	// Format dates for display and API
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")
	displayDateRange := fmt.Sprintf("%s to %s", startDate.Format("Jan 2"), endDate.Format("Jan 2, 2006"))

	// Get time entries for the week
	params := map[string]string{
		"from": startDateStr,
		"to":   endDateStr,
	}

	fmt.Printf("Fetching time entries for week of %s...\n", displayDateRange)
	timeEntries, err := client.GetTimeEntries(params)
	if err != nil {
		log.Fatalf("Failed to get time entries: %v", err)
	}

	if len(timeEntries) == 0 {
		fmt.Printf("No time entries found for week of %s\n", displayDateRange)

		// Offer navigation options
		handleSummaryNavigation(client, startDate, "week")
		return
	}

	// Group time entries by project and task
	projectSummaries := groupTimeEntriesByProject(timeEntries)

	// Display weekly summary
	fmt.Printf("\nWeekly Summary (%s):\n", displayDateRange)

	// Create a new tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print table header
	fmt.Fprintln(w, "Project\tTask\tDuration")
	fmt.Fprintln(w, "-------\t----\t--------")

	var totalHours float64

	// Sort projects by name for consistent display
	var projectNames []string
	for projectName := range projectSummaries {
		projectNames = append(projectNames, projectName)
	}
	sort.Strings(projectNames)

	for _, projectName := range projectNames {
		summary := projectSummaries[projectName]

		// Sort tasks by name
		var taskNames []string
		for taskName := range summary.TaskSummaries {
			taskNames = append(taskNames, taskName)
		}
		sort.Strings(taskNames)

		for i, taskName := range taskNames {
			hours := summary.TaskSummaries[taskName]
			hoursInt, minutes := convertDecimalToHoursMinutes(hours)

			// For the first task, include the project name
			if i == 0 {
				fmt.Fprintf(w, "%s\t%s\t%02d:%02d\n",
					projectName,
					taskName,
					hoursInt,
					minutes)
			} else {
				// For subsequent tasks, leave the project column empty
				fmt.Fprintf(w, "\t%s\t%02d:%02d\n",
					taskName,
					hoursInt,
					minutes)
			}
		}

		totalHours += summary.TotalHours
	}

	// Flush the tabwriter
	w.Flush()

	// Print total
	totalHoursInt, totalMinutes := convertDecimalToHoursMinutes(totalHours)
	fmt.Printf("\nTotal: %02d:%02d hours\n", totalHoursInt, totalMinutes)

	// Offer navigation options
	handleSummaryNavigation(client, startDate, "week")
}

// handleMonthlySummary handles showing a monthly summary of time entries
func handleMonthlySummary(client *harvest.Client, targetDate time.Time) {
	// Calculate the start of the month
	startOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())

	// Initialize with the specified month
	showMonthlySummary(client, startOfMonth)
}

// showMonthlySummary shows a summary for a specific month
func showMonthlySummary(client *harvest.Client, startDate time.Time) {
	// Calculate the end of the month
	endDate := startDate.AddDate(0, 1, -1)

	// Format dates for display and API
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")
	displayMonth := startDate.Format("January 2006")

	// Get time entries for the month
	params := map[string]string{
		"from": startDateStr,
		"to":   endDateStr,
	}

	fmt.Printf("Fetching time entries for %s...\n", displayMonth)
	timeEntries, err := client.GetTimeEntries(params)
	if err != nil {
		log.Fatalf("Failed to get time entries: %v", err)
	}

	if len(timeEntries) == 0 {
		fmt.Printf("No time entries found for %s\n", displayMonth)

		// Offer navigation options
		handleSummaryNavigation(client, startDate, "month")
		return
	}

	// Group time entries by project and task
	projectSummaries := groupTimeEntriesByProject(timeEntries)

	// Display monthly summary
	fmt.Printf("\nMonthly Summary (%s):\n", displayMonth)

	// Create a new tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print table header
	fmt.Fprintln(w, "Project\tTask\tDuration")
	fmt.Fprintln(w, "-------\t----\t--------")

	var totalHours float64

	// Sort projects by name for consistent display
	var projectNames []string
	for projectName := range projectSummaries {
		projectNames = append(projectNames, projectName)
	}
	sort.Strings(projectNames)

	for _, projectName := range projectNames {
		summary := projectSummaries[projectName]

		// Sort tasks by name
		var taskNames []string
		for taskName := range summary.TaskSummaries {
			taskNames = append(taskNames, taskName)
		}
		sort.Strings(taskNames)

		for i, taskName := range taskNames {
			hours := summary.TaskSummaries[taskName]
			hoursInt, minutes := convertDecimalToHoursMinutes(hours)

			// For the first task, include the project name
			if i == 0 {
				fmt.Fprintf(w, "%s\t%s\t%02d:%02d\n",
					projectName,
					taskName,
					hoursInt,
					minutes)
			} else {
				// For subsequent tasks, leave the project column empty
				fmt.Fprintf(w, "\t%s\t%02d:%02d\n",
					taskName,
					hoursInt,
					minutes)
			}
		}

		totalHours += summary.TotalHours
	}

	// Flush the tabwriter
	w.Flush()

	// Print total
	totalHoursInt, totalMinutes := convertDecimalToHoursMinutes(totalHours)
	fmt.Printf("\nTotal: %02d:%02d hours\n", totalHoursInt, totalMinutes)

	// Offer navigation options
	handleSummaryNavigation(client, startDate, "month")
}

// handleSummaryNavigation handles navigation between different time periods
func handleSummaryNavigation(client *harvest.Client, currentDate time.Time, periodType string) {
	options := []string{"Previous " + periodType, "Next " + periodType, "Exit"}

	prompt := promptui.Select{
		Label: "Navigation",
		Items: options,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return // Exit on error
	}

	switch index {
	case 0: // Previous period
		var newDate time.Time
		if periodType == "week" {
			newDate = currentDate.AddDate(0, 0, -7)
			showWeeklySummary(client, newDate)
		} else {
			newDate = currentDate.AddDate(0, -1, 0)
			showMonthlySummary(client, newDate)
		}
	case 1: // Next period
		var newDate time.Time
		if periodType == "week" {
			newDate = currentDate.AddDate(0, 0, 7)
			showWeeklySummary(client, newDate)
		} else {
			newDate = currentDate.AddDate(0, 1, 0)
			showMonthlySummary(client, newDate)
		}
	case 2: // Exit
		return
	}
}

// groupTimeEntriesByProject groups time entries by project and task
func groupTimeEntriesByProject(timeEntries []harvest.TimeEntry) map[string]ProjectSummary {
	projectSummaries := make(map[string]ProjectSummary)

	for _, entry := range timeEntries {
		projectName := entry.Project.Name
		taskName := entry.Task.Name

		// Get or create project summary
		summary, exists := projectSummaries[projectName]
		if !exists {
			summary = ProjectSummary{
				ProjectName:   projectName,
				TaskSummaries: make(map[string]float64),
				TotalHours:    0,
			}
		}

		// Update task hours
		summary.TaskSummaries[taskName] += entry.Hours
		summary.TotalHours += entry.Hours

		// Update the map
		projectSummaries[projectName] = summary
	}

	return projectSummaries
}
