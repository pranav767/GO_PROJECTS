# GitHub Activity CLI

A command-line tool to fetch and display recent GitHub activity for any user. This tool uses the GitHub API to retrieve public events and displays them in a clean, readable format.

## Features

- 📊 Display recent GitHub activity for any user
- 🔍 Shows various event types including:
  - Push events (with commit count)
  - Pull request activities (opened, closed, merged)
  - Issue activities (opened, closed)
  - Repository stars and forks
- 🚀 Fast and lightweight
- 💻 Cross-platform support

## Installation

### Option 1: Run from Source

Make sure you have Go installed (version 1.21 or later), then:

```bash
# Clone the repository
git clone <repository-url>
cd github-activity

# Install dependencies
go mod download

# Run the application
go run main.go <username>
```

### Option 2: Build Binary

Build the binary for your platform:

```bash
# Build for current platform
go build -o github-activity

# Run the binary
./github-activity <username>
```

### Option 3: Cross-Platform Builds

Build for different platforms:

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o github-activity.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o github-activity-macos

# Linux
GOOS=linux GOARCH=amd64 go build -o github-activity-linux
```

## Usage

### Basic Usage

```bash
# Using source code
go run main.go <username>

# Using binary
./github-activity <username>
```

### Examples

```bash
# View activity for user 'timhourigan'
go run main.go timhourigan
# or with binary:
./github-activity timhourigan
```

**Sample Output:**
```
Pushed 2 commits to timhourigan/nix-config
Closed a pull request in timhourigan/nix-config
Opened a pull request in timhourigan/nix-config
Pushed 1 commits to timhourigan/nix-config
Pushed 6 commits to timhourigan/nix-config
```

### Command Options

```bash
github-activity <username> [flags]

Flags:
  -h, --help     Show help information
  -t, --toggle   Help message for toggle
```

### Error Handling

If you don't provide a username, you'll see:
```
Error: accepts 1 arg(s), received 0
Usage:
  github-activity <username> [flags]
```

## How It Works

1. **API Integration**: The tool uses the GitHub REST API endpoint `/users/{username}/events` to fetch public events
2. **Event Processing**: It processes various GitHub event types and formats them for display
3. **Real-time Data**: Shows the most recent public activities (typically last 30 events)

## Supported Event Types

- **PushEvent**: Shows commits pushed to repositories
- **PullRequestEvent**: Shows pull request activities (opened, closed, merged)
- **IssuesEvent**: Shows issue activities (opened, closed)
- **WatchEvent**: Shows when repositories are starred
- **ForkEvent**: Shows when repositories are forked

## Requirements

- Go 1.21 or later
- Internet connection (to access GitHub API)
- Valid GitHub username

## Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework for Go
- Standard Go libraries for HTTP requests and JSON parsing

## API Rate Limits

This tool uses the GitHub API without authentication, which means:
- Rate limit: 60 requests per hour per IP address
- Only public events are accessible
- No personal or private repository data is shown

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the terms specified in the LICENSE file.

## Troubleshooting

### Common Issues

1. **"User not found" errors**: Ensure the username exists and is spelled correctly
2. **Rate limit exceeded**: Wait for the rate limit to reset (resets every hour)
3. **Network errors**: Check your internet connection

### Getting Help

Use the help flag for usage information:
```bash
go run main.go --help
# or
./github-activity --help
```

## Development

### Project Structure

```
github-activity/
├── main.go           # Entry point
├── cmd/
│   └── root.go       # CLI command definitions
├── events/
│   └── events.go     # GitHub API integration and event processing
├── go.mod            # Go module definition
├── go.sum            # Go module checksums
└── README.md         # This file
```

### Building for Distribution

```bash
# Build optimized binary
go build -ldflags="-w -s" -o github-activity
```
