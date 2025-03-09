package cmd

import (
	"bufio"
	"fmt"
	"harvest-cli/pkg/config"
	"harvest-cli/pkg/harvest"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// DeleteCmd returns the delete command
func DeleteCmd() *cobra.Command {
	var nonInteractive bool
	var date string

	cmd := &cobra.Command{
		Use:   "delete [timeEntryID]",
		Short: "Delete a time entry",
		Long: `Delete one or more time entries.
Example: h delete 123456789

By default, uses interactive mode to select time entries to delete.
Use --non-interactive flag with a time entry ID to delete directly.`,
		Args: cobra.MaximumNArgs(1),
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

			if len(args) > 0 && nonInteractive {
				// Direct delete by ID
				id, err := strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					log.Fatalf("Invalid time entry ID: %v", err)
				}
				handleDirectDelete(client, id)
			} else {
				// Interactive mode - first confirm or modify the date
				targetDate := date
				if date == time.Now().Format("2006-01-02") {
					// If using today's date (default), confirm with user
					fmt.Printf("Using default date: %s\n", date)

					// Ask if user wants to use a different date
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
							Default:   date,
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

				// Now proceed with the interactive delete using the confirmed/modified date
				handleInteractiveDelete(client, targetDate)
			}
		},
	}

	// Define flags
	cmd.Flags().BoolVarP(&nonInteractive, "non-interactive", "n", false, "Use non-interactive mode with a time entry ID")
	cmd.Flags().StringVarP(&date, "date", "d", time.Now().Format("2006-01-02"), "Date in YYYY-MM-DD format (default: today)")

	return cmd
}

// handleInteractiveDelete handles the interactive deletion of time entries
func handleInteractiveDelete(client *harvest.Client, date string) {
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

	// Create a custom multi-select interface
	reader := bufio.NewReader(os.Stdin)

	// Track selected entries
	selectedIndices := make(map[int]bool)

	// Display instructions
	fmt.Println("\nMulti-select Time Entries to Delete:")
	fmt.Println("-----------------------------------")
	fmt.Println("Instructions:")
	fmt.Println("- Enter the number of an entry to select/deselect it")
	fmt.Println("- Enter 'a' to select all entries")
	fmt.Println("- Enter 'n' to deselect all entries")
	fmt.Println("- Enter 'd' when done to proceed with deletion")
	fmt.Println("- Enter 'q' to quit without deleting")
	fmt.Println("-----------------------------------")

	// Main selection loop
	for {
		// Display the current list with selection status
		fmt.Println("\nTime entries for", date, ":")
		fmt.Println("-----------------------------------")

		for i, entry := range timeEntries {
			hours, minutes := convertDecimalToHoursMinutes(entry.Hours)
			selected := " "
			if selectedIndices[i] {
				selected = "X"
			}

			fmt.Printf("[%d] [%s] %s - %s (%02d:%02d) - %s\n",
				i+1,
				selected,
				entry.Project.Name,
				entry.Task.Name,
				hours,
				minutes,
				entry.Notes)
		}

		fmt.Println("-----------------------------------")
		fmt.Printf("%d of %d entries selected\n", len(selectedIndices), len(timeEntries))
		fmt.Print("Enter selection (number, a, n, d, q): ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch strings.ToLower(input) {
		case "a": // Select all
			for i := range timeEntries {
				selectedIndices[i] = true
			}
		case "n": // Deselect all
			selectedIndices = make(map[int]bool)
		case "d": // Done, proceed with deletion
			goto processSelection
		case "q": // Quit
			fmt.Println("Operation cancelled")
			return
		default: // Try to parse as a number
			num, err := strconv.Atoi(input)
			if err == nil && num > 0 && num <= len(timeEntries) {
				// Toggle selection
				idx := num - 1
				selectedIndices[idx] = !selectedIndices[idx]
			}
		}
	}

processSelection:
	// Collect selected entries
	var selectedEntries []harvest.TimeEntry
	for i, entry := range timeEntries {
		if selectedIndices[i] {
			selectedEntries = append(selectedEntries, entry)
		}
	}

	if len(selectedEntries) == 0 {
		fmt.Println("No entries selected. Operation cancelled.")
		return
	}

	// Display selected entries
	fmt.Println("\nSelected Time Entries:")
	fmt.Println("-----------------------------------")
	for i, entry := range selectedEntries {
		hours, minutes := convertDecimalToHoursMinutes(entry.Hours)
		fmt.Printf("%d. ID: %d - %s - %s - %s (%02d:%02d)\n",
			i+1,
			entry.ID,
			entry.SpentDate,
			entry.Project.Name,
			entry.Task.Name,
			hours,
			minutes)
	}
	fmt.Println("-----------------------------------")

	// Confirm deletion with options
	fmt.Print("Are you sure you want to delete these entries? (y/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Deletion cancelled")
		return
	}

	// Delete the selected time entries
	var successCount, failCount int
	for _, entry := range selectedEntries {
		err = client.DeleteTimeEntry(entry.ID)
		if err != nil {
			fmt.Printf("Failed to delete time entry %d: %v\n", entry.ID, err)
			failCount++
		} else {
			fmt.Printf("Time entry %d deleted successfully\n", entry.ID)
			successCount++
		}
	}

	// Summary
	fmt.Println("\nDeletion Summary:")
	fmt.Println("-----------------------------------")
	fmt.Printf("Total: %d entries\n", len(selectedEntries))
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Failed: %d\n", failCount)
	fmt.Println("-----------------------------------")
}

// handleDirectDelete handles the direct deletion of a time entry by ID
func handleDirectDelete(client *harvest.Client, id int64) {
	// Get the time entry to confirm details
	entry, err := client.GetTimeEntry(id)
	if err != nil {
		log.Fatalf("Failed to get time entry: %v", err)
	}

	// Display time entry details
	hours, minutes := convertDecimalToHoursMinutes(entry.Hours)
	fmt.Println("Time Entry Details:")
	fmt.Printf("ID: %d\n", entry.ID)
	fmt.Printf("Date: %s\n", entry.SpentDate)
	fmt.Printf("Project: %s\n", entry.Project.Name)
	fmt.Printf("Task: %s\n", entry.Task.Name)
	fmt.Printf("Hours: %02d:%02d\n", hours, minutes)
	if entry.Notes != "" {
		fmt.Printf("Notes: %s\n", entry.Notes)
	}

	// Confirm deletion with options
	deleteOptions := []string{"Delete this time entry", "Cancel deletion"}
	deletePrompt := promptui.Select{
		Label: "What would you like to do?",
		Items: deleteOptions,
	}

	deleteIndex, _, err := deletePrompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}

	if deleteIndex == 1 {
		fmt.Println("Deletion cancelled")
		return
	}

	// Delete the time entry
	err = client.DeleteTimeEntry(id)
	if err != nil {
		log.Fatalf("Failed to delete time entry: %v", err)
	}

	fmt.Printf("Time entry %d deleted successfully\n", id)
}
