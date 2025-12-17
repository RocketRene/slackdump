# Kobe User Research Slack Export Tool

A graphical user interface for the Slackdump tool, specifically designed for research purposes.

## Features

- **Simple Authentication**: Supports subdomain and cookie-based authentication
- **Channel Listing**: Get a full list of all channels in your workspace
- **Date Filtering**: Default export from 2025 (configurable)
- **Direct Integration**: Uses the slackdump package directly (not as an external binary)
- **SQLite Database Export**: Saves all data to a SQLite database using the same format as the archive command
- **Automatic Viewer**: Automatically launches the built-in web viewer after export completes

## Building

To build the GUI application:

```bash
make gui
```

Or from the GUI directory:

```bash
cd cmd/slackdump-gui
go build
```

## Running

After building, run the executable:

```bash
./slackdump-gui
```

## Usage

1. Enter your workspace subdomain (e.g., "myworkspace" for myworkspace.slack.com)
2. Enter your Slack 'd' cookie value
3. Select the year for export (defaults to 2025)
4. Click "Next: Select Channels" to authenticate and retrieve channels
5. Select the channels you want to export
6. Click "Export Selected" to start the export process
7. After export completes, the viewer will automatically launch in your browser at `http://127.0.0.1:8080`

## Export Format

The tool exports data to a SQLite database in the same format as the `slackdump archive` command:

- **Database File**: `slackdump.sqlite` in the export folder
- **Database Tables**:
  - `SESSION`: Export session metadata
  - `CHANNEL`: Channel information
  - `MESSAGE`: All messages and threads
  - `S_USER`: User information
  - `FILE`: File metadata
  - `WORKSPACE`: Workspace information
  - Additional tables for search results and mappings

The database can be:
- Queried directly using SQLite tools (e.g., SQLite Browser, DBeaver)
- Converted to other formats using `slackdump convert`
- Viewed using `slackdump view <database-file>`

## Built-in Viewer

After the export completes, the tool automatically launches a built-in web viewer:

- **URL**: `http://127.0.0.1:8080`
- **Features**:
  - Browse all exported channels
  - View messages and threads
  - Search conversations
  - Display user information
  - Show file attachments

The viewer provides an intuitive web interface to explore your exported Slack data without requiring any additional tools.

## Authentication

This tool requires:
- **Workspace Subdomain**: The name of your Slack workspace
- **Cookie (d value)**: The 'd' cookie from your Slack session

To get the 'd' cookie:
1. Log in to Slack in your browser
2. Open browser developer tools (F12)
3. Go to Application/Storage > Cookies > https://slack.com
4. Find the cookie named 'd' and copy its value

## Requirements

- Go 1.24 or later
- X11 development libraries (Linux)
- Fyne dependencies (handled by go mod)
