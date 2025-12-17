# Slack Export Tool for User Research

A simple graphical application to export your Slack workspace conversations into a structured database format for analysis and research purposes.

![Screenshot placeholder](doc/slackdump.webp)

## What This Tool Does

This application helps you:

- **Export Slack conversations** from public channels, private channels, and direct messages
- **Select specific channels** you want to export (no need to export everything)
- **Filter by date** - export only messages from a specific year (e.g., 2025)
- **Store in SQLite database** - structured, machine-readable format for analysis
- **Privacy-focused** - all data stays on your local machine, no cloud storage

## Perfect For

- **User research** - analyzing communication patterns and work artifacts
- **Skills evaluation** - understanding how teams collaborate and solve problems
- **Knowledge preservation** - saving important conversations before they're lost
- **Personal archiving** - keeping your own work history

## Getting Started

### Prerequisites

You'll need:
1. A Slack workspace URL (e.g., `your-workspace.slack.com`)
2. Your Slack authentication cookie (we'll show you how to get this)

### Installation

#### Option 1: Download Pre-built Application (Recommended)

1. Download the latest release for your operating system from the [releases page](https://github.com/rusq/slackdump/releases/)
2. Unpack the archive to any directory
3. Run `slackdump-gui` (or `slackdump-gui.exe` on Windows)

> **Note for macOS/Windows users:** You may see an "Unknown developer" warning. This is normal because the app isn't signed with a developer certificate.
>
> To bypass:
> - **Windows**: Click "More info" → "Run Anyway"
> - **macOS 14 and earlier**: Open in Finder, hold Option, double-click, choose "Run"
> - **macOS 15+**: Try to open the app, then go to System Settings → Privacy & Security → "Open Anyway"

#### Option 2: Build from Source

If you have Go installed:

```bash
git clone https://github.com/rusq/slackdump.git
cd slackdump
go run cmd/slackdump-gui/main.go
```

### Getting Your Slack Cookie

You need to get your authentication cookie from Slack's web interface:

1. Open Slack in your web browser and log in
2. Open browser Developer Tools (F12 or Cmd+Option+I on Mac)
3. Go to the "Application" or "Storage" tab
4. Find "Cookies" in the left sidebar
5. Look for a cookie starting with `xoxd-`
6. Copy the entire cookie value

**Important:** Never share this cookie with anyone - it's like your password!

### Using the Application

1. **Launch the app** - Run `slackdump-gui`

2. **Step 1: Authentication**
   - Enter your workspace subdomain (e.g., `your-workspace`)
   - Paste your Slack cookie (starting with `xoxd-`)
   - Choose the output folder (default: `~/Documents/koberesearch/`)
   - Select the year to export (e.g., 2025)
   - Click "Authenticate & Fetch Channels"

3. **Step 2: Select Channels**
   - Browse the list of available channels
   - Use "Select All" or choose specific channels
   - Click "Export Selected"

4. **Step 3: Export**
   - Watch the progress as your data is exported
   - When complete, you'll see statistics about the export
   - Click "Open Export Folder" to view your database file

### What Gets Exported

✅ **Included:**
- All message text and threads
- User information
- Channel names and metadata
- Message timestamps
- Thread relationships

❌ **Not included:**
- Files and attachments
- Images and media
- Emojis and reactions
- User avatars

### Understanding Your Export

Your export is saved as a SQLite database file (`.db`) in a timestamped folder:

```
~/Documents/koberesearch/
└── export_20250117_143022/
    └── slackdump.db
```

This database contains structured tables:
- `MESSAGE` - all messages and threads
- `CHANNEL` - channel information
- `S_USER` - user profiles

You can analyze this database using:
- SQLite browser tools
- SQL queries
- AI tools like Claude Code
- Data analysis scripts

## Privacy & Data Handling

🔒 **Your data stays private:**
- All exports are stored **locally on your machine**
- No data is sent to any cloud services
- No unencrypted online storage
- Easy to delete - just remove the export folder

⚠️ **Security reminders:**
- Keep your Slack cookie secure (don't share it)
- Delete exports when you're done with them
- Store exports in encrypted folders if needed

## Limitations

- **Free Slack workspaces**: Only exports messages from the last 90 days (Slack API limitation)
- **Export speed**: Depends on the number of messages and channels
- **Year filtering**: Only exports messages from the selected year
- **No media files**: Attachments and images are not downloaded

## Troubleshooting

### "Auth failed" error
- Check that your cookie is correct and hasn't expired
- Try getting a fresh cookie from Slack
- Make sure your workspace subdomain is correct

### "No channels found"
- Verify you're logged into the correct workspace
- Check your cookie hasn't expired
- Ensure you have access to channels in the workspace

### Export is slow
- This is normal for workspaces with many messages
- Reduce the number of channels selected
- Consider exporting fewer months of data

### Application won't start (macOS)
- Follow the security bypass steps above
- Check System Settings → Privacy & Security

## Getting Help

- **Technical issues**: [Open an issue on GitHub](https://github.com/rusq/slackdump/issues)
- **Questions**: Check existing issues or start a discussion
- **Documentation**: See the main [slackdump documentation](https://github.com/rusq/slackdump)

## Advanced Usage

This GUI application is built on top of [slackdump](https://github.com/rusq/slackdump), a powerful command-line tool. For more advanced features like:
- Viewing exported data
- Converting between formats
- Automated exports
- Emoji downloads

Check out the [full slackdump documentation](https://github.com/rusq/slackdump).

## Technical Details

- **Language**: Go
- **GUI Framework**: Fyne
- **Database**: SQLite (pure Go implementation, no CGO required)
- **Slack API**: Uses official Slack Web API
- **License**: See LICENSE file

## Contributing

This is a research tool under active development. Contributions are welcome:
- Report bugs via GitHub issues
- Suggest features
- Submit pull requests
- Improve documentation

## Credits

Built on [slackdump](https://github.com/rusq/slackdump) by [@rusq](https://github.com/rusq)

GUI application developed for user research and skills evaluation studies.

## License

See LICENSE file for details.
