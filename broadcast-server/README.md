# WebSocket Broadcast Server

A command-line WebSocket broadcast server and client implementation in Go, allowing multiple clients to connect and communicate in a chat-like environment.

## Features

- WebSocket-based real-time communication
- Command-line interface using Cobra
- Support for multiple concurrent clients
- Named client connections
- System messages for join/leave events
- Name change functionality
- Clean connection handling and error management

## Project Structure

```
broadcast-server/
├── cmd/
│   ├── connect.go    # Client connection command
│   ├── root.go      # Root command configuration
│   └── start.go     # Server start command
├── client/
│   └── client.go    # WebSocket client implementation
├── server/
│   └── server.go    # WebSocket server implementation
├── main.go          # Application entry point
├── go.mod          # Go module file
└── README.md       # This file
```

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd broadcast-server
```

2. Install dependencies:
```bash
go mod tidy
```

## Usage

### Starting the Server

```bash
go run main.go start
```

The server will start on `localhost:8080` by default.

### Connecting Clients

```bash
go run main.go connect --name "YourName"
```

Options:
- `--name, -n`: Set your display name (default: "anonymous")
- `--host, -H`: Server host address (default: "localhost")
- `--port, -p`: Server port (default: 8080)

### Chat Commands

- Send a message: Just type and press Enter
- Change name: Type `/name NewName`
- Quit: Press Ctrl+C

## Implementation Details

### Server (`server/server.go`)
- Uses Gorilla WebSocket for WebSocket functionality
- Maintains a thread-safe client map
- Broadcasts messages to all connected clients
- Handles client disconnections gracefully
- Supports system messages for events

### Client (`client/client.go`)
- Connects to the WebSocket server
- Manages user input and message display
- Handles server messages in a separate goroutine
- Supports command processing
- Clean disconnection handling

### Commands (`cmd/`)
- Uses Cobra for command-line interface
- Provides intuitive command structure
- Handles command-line flags
- Clear command documentation

## Dependencies

- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket implementation
- [Cobra](https://github.com/spf13/cobra) - Command-line interface

## Development

### Building
```bash
go build
```

### Running Tests
```bash
go test ./...
```