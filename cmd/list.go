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

// TaskSummary represents a summary of time entries for a task across projects
type TaskSummary struct {
	TaskName   string
	TotalHours float64
}

// ListCmd returns the list command
func ListCmd() *cobra.Command {
	var monthly, weekly, yearly bool
	var date string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List time entries",
		Long: `List time entries for a specific day, week, month, or year.
By default, lists all time entries for the current day.
Use -d flag to specify a date (YYYY-MM-DD format).
Use -w flag for weekly summary.
Use -m flag for monthly summary.
Use -y flag for yearly summary (based on year_start_date in config, defaults to January 1st).`,
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

			if yearly {
				// Yearly summary
				handleYearlySummary(client, targetDate)
			} else if monthly {
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
	cmd.Flags().BoolVarP(&yearly, "yearly", "y", false, "Show yearly summary")
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
	taskHours := make(map[string]float64)

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

		// Aggregate hours by task
		taskHours[entry.Task.Name] += entry.Hours
	}

	// Flush the tabwriter
	w.Flush()

	// Print total
	totalHoursInt, totalMinutes := convertDecimalToHoursMinutes(totalHours)
	fmt.Printf("\nTotal: %02d:%02d hours\n", totalHoursInt, totalMinutes)

	// Print task-based aggregation
	fmt.Println("\nTime by Task:")
	fmt.Println("------------------------------------")

	// Sort tasks by name
	var taskNames []string
	for taskName := range taskHours {
		taskNames = append(taskNames, taskName)
	}
	sort.Strings(taskNames)

	// Create a new tabwriter for task summary
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Task\tDuration\t% of Total")
	fmt.Fprintln(tw, "----\t--------\t----------")

	for _, taskName := range taskNames {
		hours := taskHours[taskName]
		hoursInt, minutes := convertDecimalToHoursMinutes(hours)
		percentage := (hours / totalHours) * 100

		fmt.Fprintf(tw, "%s\t%02d:%02d\t%.1f%%\n",
			taskName,
			hoursInt,
			minutes,
			percentage)
	}

	tw.Flush()
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
	taskHours := make(map[string]float64)

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

			// Aggregate hours by task across all projects
			taskHours[taskName] += hours
		}

		totalHours += summary.TotalHours
	}

	// Flush the tabwriter
	w.Flush()

	// Print total
	totalHoursInt, totalMinutes := convertDecimalToHoursMinutes(totalHours)
	fmt.Printf("\nTotal: %02d:%02d hours\n", totalHoursInt, totalMinutes)

	// Print task-based aggregation
	fmt.Println("\nTime by Task (across all projects):")
	fmt.Println("------------------------------------")

	// Sort tasks by name
	var taskNames []string
	for taskName := range taskHours {
		taskNames = append(taskNames, taskName)
	}
	sort.Strings(taskNames)

	// Create a new tabwriter for task summary
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Task\tDuration\t% of Total")
	fmt.Fprintln(tw, "----\t--------\t----------")

	for _, taskName := range taskNames {
		hours := taskHours[taskName]
		hoursInt, minutes := convertDecimalToHoursMinutes(hours)
		percentage := (hours / totalHours) * 100

		fmt.Fprintf(tw, "%s\t%02d:%02d\t%.1f%%\n",
			taskName,
			hoursInt,
			minutes,
			percentage)
	}

	tw.Flush()

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

	// Display monthly summary
	fmt.Printf("\nMonthly Summary (%s):\n", displayMonth)

	// Calculate capacity and utilization metrics
	monthlyCapacity := appConfig.GetMonthlyCapacityHours()

	// Calculate period length in months (should be 1.0 for a complete month)
	periodLength := calculateMonthsBetween(startDate, endDate.AddDate(0, 0, 1))
	periodCapacity := monthlyCapacity * periodLength

	var totalHours float64
	var billableHours float64

	// Create maps for task summaries
	taskSummaries := make(map[string]float64)
	billableTaskSummaries := make(map[string]float64)
	nonBillableTaskSummaries := make(map[string]float64)

	// Track project-wise hours
	projectHours := make(map[string]float64)

	// Process time entries
	for _, entry := range timeEntries {
		projectName := entry.Project.Name
		taskName := entry.Task.Name
		hours := entry.Hours

		// Add to task and project totals
		taskSummaries[taskName] += hours
		projectHours[projectName] += hours
		totalHours += hours

		// Check if task is billable
		if appConfig.IsBillableTask(int(entry.Task.ID)) {
			billableHours += hours
			billableTaskSummaries[taskName] += hours
		} else {
			nonBillableTaskSummaries[taskName] += hours
		}
	}

	// Calculate overtime or remaining capacity hours
	leaveHours := billableHours - periodCapacity
	leaveDays := leaveHours / 8.0 // Converting hours to days based on 8-hour workdays

	// Display simplified capacity metrics
	fmt.Printf("\nCapacity Metrics:\n")
	fmt.Printf("- Period Length: %.2f months\n", periodLength)
	fmt.Printf("- Period Capacity: %.2f hours\n", periodCapacity)
	fmt.Printf("- Total Hours: %.2f hours\n", totalHours)
	fmt.Printf("- Billable Hours: %.2f hours\n", billableHours)

	// Display overtime or remaining capacity
	if leaveDays >= 0 {
		fmt.Printf("- Overtime (in days): %.2f days (%.2f hours)\n",
			leaveDays, leaveHours)
	} else {
		fmt.Printf("- Capacity Remaining (in days): %.2f days (%.2f hours)\n",
			-leaveDays, -leaveHours)
	}

	fmt.Println("\nBillable Tasks Summary:")
	fmt.Println("------------------------")

	// Create a tabwriter for tasks
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Task\tHours\t% of Total\t% of Capacity")
	fmt.Fprintln(w, "----\t-----\t-----------\t------------")

	// Sort and display billable tasks first
	var billableTaskNames []string
	for taskName := range billableTaskSummaries {
		billableTaskNames = append(billableTaskNames, taskName)
	}
	sort.Strings(billableTaskNames)

	for _, taskName := range billableTaskNames {
		hours := billableTaskSummaries[taskName]
		percentOfTotal := (hours / totalHours) * 100
		percentOfCapacity := (hours / periodCapacity) * 100

		fmt.Fprintf(w, "%s\t%.2f\t%.1f%%\t%.1f%%\n",
			taskName,
			hours,
			percentOfTotal,
			percentOfCapacity)
	}

	// Calculate billable totals
	fmt.Fprintf(w, "TOTAL BILLABLE\t%.2f\t%.1f%%\t%.1f%%\n",
		billableHours,
		(billableHours/totalHours)*100,
		(billableHours/periodCapacity)*100)

	w.Flush()

	// Display project summary
	fmt.Println("\nProject Summary:")
	fmt.Println("---------------")

	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Project\tHours\t% of Total")
	fmt.Fprintln(w, "-------\t-----\t----------")

	// Sort projects
	var projectNames []string
	for projectName := range projectHours {
		projectNames = append(projectNames, projectName)
	}
	sort.Strings(projectNames)

	for _, projectName := range projectNames {
		hours := projectHours[projectName]
		percentOfTotal := (hours / totalHours) * 100

		fmt.Fprintf(w, "%s\t%.2f\t%.1f%%\n",
			projectName,
			hours,
			percentOfTotal)
	}

	fmt.Fprintf(w, "TOTAL\t%.2f\t100.0%%\n", totalHours)

	w.Flush()

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

// handleYearlySummary handles the yearly summary view
func handleYearlySummary(client *harvest.Client, targetDate time.Time) {
	// Get year start date from config
	startMonth, startDay, err := appConfig.GetYearStartDate()
	if err != nil {
		log.Fatalf("Failed to get year start date: %v", err)
	}

	// Determine the year boundaries
	var yearStart time.Time
	currentYear := targetDate.Year()

	// Create the start date for the current year cycle
	yearStart = time.Date(currentYear, time.Month(startMonth), startDay, 0, 0, 0, 0, targetDate.Location())

	// If the target date is before the year start in the current calendar year,
	// we need to use the previous calendar year's start date
	if targetDate.Before(yearStart) {
		yearStart = time.Date(currentYear-1, time.Month(startMonth), startDay, 0, 0, 0, 0, targetDate.Location())
	}

	// Calculate the end date (next year's start - 1 day)
	yearEnd := time.Date(yearStart.Year()+1, time.Month(startMonth), startDay, 0, 0, 0, 0, targetDate.Location()).AddDate(0, 0, -1)

	// If the target date is in the future, use today as the end date
	now := time.Now()
	if yearEnd.After(now) {
		yearEnd = now
	}

	// Format for API call and display
	from := yearStart.Format("2006-01-02")
	to := yearEnd.Format("2006-01-02")
	yearLabel := fmt.Sprintf("%d/%d", yearStart.Year(), yearStart.Year()+1)

	// If using standard calendar year, just show the year
	if startMonth == 1 && startDay == 1 {
		yearLabel = fmt.Sprintf("%d", yearStart.Year())
	}

	fmt.Printf("Yearly Summary (%s)\n", yearLabel)
	fmt.Printf("Period: %s to %s\n\n", from, to)

	// Get time entries for the period
	params := map[string]string{
		"from": from,
		"to":   to,
	}

	entries, err := client.GetTimeEntries(params)
	if err != nil {
		log.Fatalf("Failed to get time entries: %v", err)
	}

	if len(entries) == 0 {
		fmt.Println("No time entries found for this period.")
		return
	}

	// Calculate period length in months
	periodLength := calculateMonthsBetween(yearStart, yearEnd.AddDate(0, 0, 1))

	// Calculate capacity based on monthly capacity
	monthlyCapacity := appConfig.GetMonthlyCapacityHours()
	yearlyCapacity := monthlyCapacity * periodLength

	// Initialize counters
	var totalHours float64
	var billableHours float64

	// Create maps for task summaries
	taskSummaries := make(map[string]float64)
	billableTaskSummaries := make(map[string]float64)
	nonBillableTaskSummaries := make(map[string]float64)

	// Track project-wise hours
	projectHours := make(map[string]float64)

	// Process all entries
	for _, entry := range entries {
		projectName := entry.Project.Name
		taskName := entry.Task.Name
		hours := entry.Hours

		// Add to task and project totals
		taskSummaries[taskName] += hours
		projectHours[projectName] += hours
		totalHours += hours

		// Check if task is billable
		if appConfig.IsBillableTask(int(entry.Task.ID)) {
			billableHours += hours
			billableTaskSummaries[taskName] += hours
		} else {
			nonBillableTaskSummaries[taskName] += hours
		}
	}

	// Calculate overtime or remaining capacity hours
	leaveHours := billableHours - yearlyCapacity
	leaveDays := leaveHours / 8.0 // Converting hours to days based on 8-hour workdays

	// Display simplified capacity metrics
	fmt.Printf("Capacity Metrics:\n")
	fmt.Printf("- Period Length: %.2f months\n", periodLength)
	fmt.Printf("- Period Capacity: %.2f hours\n", yearlyCapacity)
	fmt.Printf("- Total Hours: %.2f hours\n", totalHours)
	fmt.Printf("- Billable Hours: %.2f hours\n", billableHours)

	// Display overtime or remaining capacity
	if leaveDays >= 0 {
		fmt.Printf("- Overtime (in days): %.2f days (%.2f hours)\n",
			leaveDays, leaveHours)
	} else {
		fmt.Printf("- Capacity Remaining (in days): %.2f days (%.2f hours)\n",
			-leaveDays, -leaveHours)
	}

	fmt.Println("\nBillable Tasks Summary:")
	fmt.Println("------------------------")

	// Create a tabwriter for tasks
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Task\tHours\t% of Total\t% of Capacity\t")
	fmt.Fprintln(w, "----\t-----\t-----------\t------------\t")

	// Sort and display billable tasks first
	var billableTaskNames []string
	for taskName := range billableTaskSummaries {
		billableTaskNames = append(billableTaskNames, taskName)
	}
	sort.Strings(billableTaskNames)

	for _, taskName := range billableTaskNames {
		hours := billableTaskSummaries[taskName]
		percentOfTotal := (hours / totalHours) * 100
		percentOfCapacity := (hours / yearlyCapacity) * 100

		fmt.Fprintf(w, "%s\t%.2f\t%.1f%%\t%.1f%%\t\n",
			taskName,
			hours,
			percentOfTotal,
			percentOfCapacity)
	}

	// Calculate billable totals
	fmt.Fprintf(w, "TOTAL BILLABLE\t%.2f\t%.1f%%\t%.1f%%\t\n",
		billableHours,
		(billableHours/totalHours)*100,
		(billableHours/yearlyCapacity)*100)

	w.Flush()

	// Display project summary
	fmt.Println("\nProject Summary:")
	fmt.Println("---------------")

	w = tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Project\tHours\t% of Total\t")
	fmt.Fprintln(w, "-------\t-----\t----------\t")

	// Sort projects
	var projectNames []string
	for projectName := range projectHours {
		projectNames = append(projectNames, projectName)
	}
	sort.Strings(projectNames)

	for _, projectName := range projectNames {
		hours := projectHours[projectName]
		percentOfTotal := (hours / totalHours) * 100

		fmt.Fprintf(w, "%s\t%.2f\t%.1f%%\t\n",
			projectName,
			hours,
			percentOfTotal)
	}

	fmt.Fprintf(w, "TOTAL\t%.2f\t100.0%%\t\n", totalHours)

	w.Flush()
}

// calculateMonthsBetween calculates the number of months between two dates
func calculateMonthsBetween(start, end time.Time) float64 {
	// Handle same day or invalid range
	if end.Before(start) || end.Equal(start) {
		return 0.0
	}

	// Calculate complete months
	completeMonths := 0

	// Start with the current month
	currentMonth := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())

	// Calculate months between start and end
	for currentMonth.Before(end) {
		nextMonth := currentMonth.AddDate(0, 1, 0)
		// Only count if the entire month is within our range
		if nextMonth.Before(end) || nextMonth.Equal(end) {
			completeMonths++
		}
		currentMonth = nextMonth
	}

	// Calculate days in the final partial month
	var partialMonth float64 = 0.0
	if start.AddDate(0, completeMonths, 0).Before(end) {
		// Days in the final month
		daysInMonth := time.Date(end.Year(), end.Month()+1, 0, 0, 0, 0, 0, end.Location()).Day()
		// Calculate the fraction of the month
		partialMonth = float64(end.Day()) / float64(daysInMonth)
	}

	// Return complete months plus any partial month
	return float64(completeMonths) + partialMonth
}
