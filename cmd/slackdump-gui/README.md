# Kobe User Research Slack Export Tool

A graphical user interface for the Slackdump tool, specifically designed for research purposes.

## Features

- **Simple Authentication**: Supports subdomain and cookie-based authentication
- **Channel Listing**: Get a full list of all channels in your workspace
- **Date Filtering**: Default export from 2025 (configurable)
- **Direct Integration**: Uses the slackdump package directly (not as an external binary)

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
4. Click "Get Full List of Channels" to authenticate and retrieve channels
5. Use "Export Selected Channels" for future export functionality

## Authentication

This tool requires:
- **Workspace Subdomain**: The name of your Slack workspace
- **Cookie (d value)**: The 'd' cookie from your Slack session

To get the 'd' cookie:
1. Log in to Slack in your browser
2. Open browser developer tools (F12)
3. Go to Application/Storage > Cookies > https://slack.com
4. Find the cookie named 'd' and copy its value

## Note

This is a minimal MVP (Minimum Viable Product) for research purposes. The export functionality is a placeholder and would need to be implemented based on specific requirements.

## Requirements

- Go 1.24 or later
- X11 development libraries (Linux)
- Fyne dependencies (handled by go mod)
