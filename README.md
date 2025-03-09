# Harvest CLI Utility

A simple command-line utility for logging time entries for your work.

## Installation

```bash
# Clone the repository
git clone https://github.com/amanangira/harvest-cli-utility.git

# Navigate to the project directory
cd harvest-cli-utility

# Build the project
go build -o h

# Move the binary to a directory in your PATH (optional)
mv h /usr/local/bin/
```

## Configuration

The application uses a `config.json` file to store project and task information. The file should be placed in the same directory as the executable or in the current working directory.

Example `config.json`:
```json
{
  "projects": [
    {
      "id": 123,
      "name": "Company A | Project AA",
      "tasks": [
        {
          "id": 456,
          "name": "Software Development"
        },
        {
          "id": 678,
          "name": "Non-Billable"
        }
      ]
    },
    {
      "id": 111,
      "name": "Company B | Project BB",
      "tasks": [
        {
          "id": 222,
          "name": "Software Development"
        },
        {
          "id": 333,
          "name": "Fun"
        }
      ]
    }
  ],
  "default_project": "Company A | Project AA",
  "default_task": "Software Development",
  "harvest_api": {
    "account_id": "YOUR_HARVEST_ACCOUNT_ID",
    "token": "YOUR_HARVEST_API_TOKEN"
  }
}
```

### Harvest API Configuration

To use the Harvest API integration, you need to:

1. Get your Harvest Account ID and API Token from your Harvest account
2. Add them to the `harvest_api` section in your `config.json` file

## Usage

The CLI utility uses a simple syntax:

```
h [command] [arguments]
```

### Available Commands

#### Create a Time Entry

```
h create [flags]
```

Flags:
- `-d, --date string`: Date in YYYY-MM-DD format (default: today)
- `-p, --project string`: Project name (must match a name in config.json)
- `-a, --action string`: Action/Task name (must match a task name for the selected project in config.json)
- `-t, --duration string`: Duration in HH:MM format (e.g., "7:30" for 7 hours and 30 minutes)
- `-D, --default-mode`: Use default mode (uses default project and task from config)

Examples:
```bash
# Create a time entry with all arguments
h create -d 2023-03-06 -p "Corporate Visions | vPlaybook" -a "Software Development" -t "7:30"

# Create a time entry with interactive prompts
h create

# Create a time entry with some arguments and prompts for others
h create -d 2023-03-06 -t "4:30"

# Create a time entry using default mode (uses default project and task from config)
h create -D
```

#### Delete Time Entries

```
h delete [timeEntryID] [flags]
```

Flags:
- `-n, --non-interactive`: Use non-interactive mode with a time entry ID
- `-d, --date string`: Date in YYYY-MM-DD format (default: today)

By default, the delete command operates in interactive mode, allowing you to:
- Confirm or modify the default date before proceeding
- Select multiple time entries to delete one by one
- View a summary of selected entries before deletion
- Get a deletion summary showing successful and failed operations

Examples:
```bash
# Delete time entries interactively (default behavior)
h delete

# Delete time entries from a specific date
h delete -d 2023-03-06

# Delete a specific time entry by ID in non-interactive mode
h delete 123456789 -n
```

#### Update a Time Entry

```
h update [flags]
```

Flags:
- `-i, --interactive`: Use interactive mode (deprecated, now the default behavior)
- `-d, --date string`: Date in YYYY-MM-DD format (default: today) - used to specify the date for time entry selection

The update process:
- Allows you to confirm or modify the default date
- Shows all time entries for the selected date
- Lets you select which time entry to update
- Shows a summary of changes before confirming the update
- Provides an intuitive menu interface for confirming actions
- Displays detailed information about the updated time entry

Examples:
```bash
# Update a time entry (shows all entries for today)
h update

# Update a time entry from a specific date
h update -d 2023-03-06
```

#### List Time Entries

```
h list [flags]
```

Flags:
- `-d, --date string`: Date in YYYY-MM-DD format (default: today) - used to list entries for a specific date
- `-m, --monthly`: Show monthly summary
- `-w, --weekly`: Show weekly summary

Examples:
```bash
# List time entries for today
h list

# List time entries for a specific date
h list -d 2023-03-06

# Show weekly summary for the current week
h list -w

# Show weekly summary for a specific week (using a date within that week)
h list -w -d 2023-03-06

# Show monthly summary for the current month
h list -m

# Show monthly summary for a specific month (using a date within that month)
h list -m -d 2023-03-06
```

## Features

- Interactive prompts for missing arguments
- Date validation in YYYY-MM-DD format
- Project selection from configuration
- Task selection based on the selected project
- Time input in HH:MM format
- Default mode for quick time entry creation
- Integration with Harvest API for creating, updating, and deleting time entries
- Daily, weekly, and monthly summaries of time entries
- Tabular output format for better readability
- Date filtering for all commands
- Intuitive menu-based confirmation interfaces
- Detailed summaries of changes before confirmation
- Multiple time entry selection for batch operations
- Interactive navigation through time entries
- Default interactive mode for better user experience

## License

This project is licensed under the MIT License - see the LICENSE file for details. 