package cmd

import (
	"fmt"
	"harvest-cli/pkg/config"
	"harvest-cli/pkg/harvest"
	"log"
	"strconv"
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

	// Create a list of time entries for selection with promptui
	prompt := promptui.Select{
		Label: "Select time entries to delete (press enter to select, then confirm)",
		Items: timeEntries,
		Size:  10,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ .Project.Name }} - {{ .Task.Name }}",
			Active:   "→ {{ .Project.Name }} - {{ .Task.Name }} ({{ .Hours }} hours) - {{ .Notes }}",
			Inactive: "  {{ .Project.Name }} - {{ .Task.Name }} ({{ .Hours }} hours) - {{ .Notes }}",
			Selected: "Selected: {{ .Project.Name }} - {{ .Task.Name }}",
		},
	}

	// Track selected entries
	var selectedEntries []harvest.TimeEntry

	// Allow multiple selection
	fmt.Println("\nSelect time entries to delete:")
	fmt.Println("Select entries one by one, you'll be asked if you want to select more after each selection")

	for {
		index, _, err := prompt.Run()
		if err != nil {
			if err.Error() == "^C" {
				fmt.Println("Operation cancelled")
				return
			}
			log.Fatalf("Prompt failed: %v", err)
		}

		// Add the selected entry
		selectedEntries = append(selectedEntries, timeEntries[index])

		// Ask if user wants to select more
		continueOptions := []string{"Select another entry", "Proceed with deletion"}
		continuePrompt := promptui.Select{
			Label: fmt.Sprintf("%d entries selected. What would you like to do?", len(selectedEntries)),
			Items: continueOptions,
		}

		continueIndex, _, err := continuePrompt.Run()
		if err != nil {
			log.Fatalf("Prompt failed: %v", err)
		}

		if continueIndex == 1 {
			// Proceed with deletion
			break
		}
	}

	if len(selectedEntries) == 0 {
		fmt.Println("No entries selected. Operation cancelled.")
		return
	}

	// Display selected entries
	fmt.Println("\nSelected Time Entries:")
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

	// Confirm deletion with options
	deleteOptions := []string{"Delete selected time entries", "Cancel deletion"}
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
	fmt.Printf("\nDeletion Summary: %d successful, %d failed\n", successCount, failCount)
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
