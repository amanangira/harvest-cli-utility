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
h [option] [arguments]
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

# Create a time entry with default values for unspecified arguments
h create --default -p "Tech Internal"

# Create a time entry using default mode (uses default project and task from config)
h create -D
```

## Features

- Interactive prompts for missing arguments
- Date validation in YYYY-MM-DD format
- Project selection from configuration
- Task selection based on the selected project
- Time input in HH:MM format
- Default values for convenience
- Default mode for quick time entry creation
- Integration with Harvest API for creating time entries

## License

This project is licensed under the MIT License - see the LICENSE file for details. 