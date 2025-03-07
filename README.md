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
      "id": 41788170,
      "name": "Corporate Visions | vPlaybook",
      "tasks": [
        {
          "id": 16593200,
          "name": "Software Development"
        },
        {
          "id": 16593732,
          "name": "Non-Billable"
        }
      ]
    },
    {
      "id": 28708733,
      "name": "Tech9 Internal",
      "tasks": [
        {
          "id": 16593200,
          "name": "Software Development"
        },
        {
          "id": 16593732,
          "name": "Non Billable"
        }
      ]
    }
  ]
}
```

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
- `--task string`: Task name (must match a task name for the selected project in config.json)
- `-t, --time float`: Time in decimal hours (e.g., 7.5)
- `--default`: Use default values for unspecified arguments

Examples:
```bash
# Create a time entry with all arguments
h create -d 2023-03-06 -p "Corporate Visions | vPlaybook" --task "Software Development" -t 7.5

# Create a time entry with interactive prompts
h create

# Create a time entry with some arguments and prompts for others
h create -d 2023-03-06 -t 4.5

# Create a time entry with default values for unspecified arguments
h create --default -p "Tech9 Internal"
```

## Features

- Interactive prompts for missing arguments
- Date validation in YYYY-MM-DD format
- Project selection from configuration
- Task selection based on the selected project
- Time input in decimal hours with conversion to HH:MM format
- Default values for convenience

## License

This project is licensed under the MIT License - see the LICENSE file for details. 