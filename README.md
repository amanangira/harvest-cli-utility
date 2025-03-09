# Harvest CLI Utility

A simple command-line utility for logging time entries directly from your terminal to your Harvest account. This tool helps developers and teams track their time efficiently without leaving the command line.

Harvest CLI Demo

## Features

- ✅ Create, update, delete, and list time entries directly from your terminal
- ✅ Interactive prompts with smart defaults for quick time logging
- ✅ Project and task selection from your Harvest account configuration
- ✅ Daily, weekly, and monthly time summaries
- ✅ Multiple time entry selection for batch operations
- ✅ Tabular output format for better readability
- ✅ Date filtering for all commands
- ✅ Default interactive mode for better user experience

## Quick Start

### Prerequisites

- Go 1.18 or higher
- A Harvest account with API access

### Installation

#### Automatic Installation

We provide installation scripts for easy setup:

**macOS/Linux:**
```bash
# Clone the repository
git clone https://github.com/amanangira/harvest-cli-utility.git

# Navigate to the project directory
cd harvest-cli-utility

# Run the installation script
chmod +x install.sh
./install.sh
```

**Windows:**
```bash
# Clone the repository
git clone https://github.com/amanangira/harvest-cli-utility.git

# Navigate to the project directory
cd harvest-cli-utility

# Run the installation script (Command Prompt)
install.bat

# Or if using Git Bash
chmod +x install.sh
./install.sh
```

The installation scripts will:
1. Check if Go is installed with the right version
2. Build the project
3. Install the binary to your PATH
4. Create a sample config.json if one doesn't exist
5. Provide instructions for configuration

#### Manual Installation

If you prefer to install manually:

```bash
# Clone the repository
git clone https://github.com/amanangira/harvest-cli-utility.git

# Navigate to the project directory
cd harvest-cli-utility

# Build the project
go build -o h

# Move the binary to a directory in your PATH (optional)
# For macOS/Linux:
mv h /usr/local/bin/
# For Windows (in PowerShell, run as Administrator):
# Copy-Item -Path .\h.exe -Destination "$env:USERPROFILE\bin\h.exe"
```

### Configuration

Create a `config.json` file in the same directory as the executable with your Harvest API credentials and project information:

```json
{
  "projects": [
    {
      "id": 123,
      "name": "Project A",
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
    }
  ],
  "default_project": "Project A",
  "default_task": "Software Development",
  "harvest_api": {
    "account_id": "YOUR_HARVEST_ACCOUNT_ID",
    "token": "YOUR_HARVEST_API_TOKEN"
  }
}
```

#### Getting Your Harvest API Credentials

1. Log in to your Harvest account
2. Go to Settings > Developer
3. Create a new personal access token
4. Note your Account ID and Token

## Usage Guide

The CLI utility uses a simple syntax:

```
h [command] [arguments]
```

### Available Commands

#### Create a Time Entry

```bash
# Quick create with defaults
h create -D

# Interactive create (prompts for all fields)
h create

# Create with specific values
h create -d 2023-03-06 -p "Project A" -a "Software Development" -t "7:30"
```

Flags:
- `-d, --date string`: Date in YYYY-MM-DD format (default: today)
- `-p, --project string`: Project name (must match a name in config.json)
- `-a, --action string`: Action/Task name (must match a task name for the selected project)
- `-t, --duration string`: Duration in HH:MM format (e.g., "7:30" for 7 hours and 30 minutes)
- `-D, --default-mode`: Use default project and task from config

#### Delete Time Entries

```bash
# Delete entries interactively (select multiple)
h delete

# Delete entries for a specific date
h delete -d 2023-03-06

# Delete a specific entry by ID
h delete 123456789 -n
```

Flags:
- `-n, --non-interactive`: Use non-interactive mode with a time entry ID
- `-d, --date string`: Date in YYYY-MM-DD format (default: today)

**Multi-Selection Interface:**

The delete command features an intuitive multi-selection interface that allows you to:
- Select/deselect individual entries by entering their number
- Select all entries at once with 'a'
- Deselect all entries with 'n'
- Proceed with deletion with 'd'
- Cancel the operation with 'q'

This makes it easy to delete multiple time entries in a single operation.

#### Update a Time Entry

```bash
# Update an entry (select from today's entries)
h update

# Update an entry from a specific date
h update -d 2023-03-06
```

Flags:
- `-d, --date string`: Date in YYYY-MM-DD format (default: today)

#### List Time Entries

```bash
# List today's entries
h list

# List entries for a specific date
h list -d 2023-03-06

# Show weekly summary
h list -w

# Show monthly summary
h list -m
```

Flags:
- `-d, --date string`: Date in YYYY-MM-DD format (default: today)
- `-m, --monthly`: Show monthly summary
- `-w, --weekly`: Show weekly summary

## Common Workflows

### Log Time for Today

```bash
# Quick logging with defaults
h create -D

# Interactive logging
h create
```

### Review and Update Today's Entries

```bash
# List today's entries
h list

# Update an entry if needed
h update
```

### Weekly Review

```bash
# View weekly summary
h list -w

# Delete incorrect entries
h delete
```

### Batch Delete Multiple Entries

```bash
# Select the date containing entries to delete
h delete -d 2023-03-06

# Use the multi-select interface:
# - Enter numbers to select/deselect entries
# - Enter 'a' to select all
# - Enter 'd' when done
```

## Troubleshooting

### Common Issues

1. **API Authentication Errors**: Ensure your Harvest API credentials are correct in config.json
2. **Project/Task Not Found**: Verify that project and task names match exactly with your config.json
3. **Command Not Found**: Make sure the binary is in your PATH or use the full path to the executable

### Getting Help

For each command, you can use the `--help` flag to see available options:

```bash
h --help
h create --help
h delete --help
h update --help
h list --help
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 
