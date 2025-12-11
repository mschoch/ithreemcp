# ithreemcp

An MCP server exposing functionality of the connected i3 window manager.

Control your i3 windows using natural language through AI assistants. Ask your LLM to move windows between workspaces, focus applications, toggle floating mode, and more.

![ithreemcp demo](demo.gif)

## Requirements

- [i3 window manager](https://i3wm.org/) running with socket access
- Go 1.23+ (for building from source)

## Installation

### From Source

```bash
go install github.com/mschoch/ithreemcp@latest
```

Or clone and build manually:

```bash
git clone https://github.com/mschoch/ithreemcp.git
cd ithreemcp
go build
```

## Building

`go build`

## Running

**Prerequisites:** You need the i3 window manager running, with permission to access its socket.

### Quick Start with ollmcp

The easiest way to use ithreemcp is with [ollmcp](https://github.com/jonigl/mcp-client-for-ollama), an MCP client for Ollama.

1. Copy the sample config and adjust for your environment:
   ```bash
   cp ithreemcp.json.sample ithreemcp.json
   ```

2. Edit `ithreemcp.json` to ensure the path to `ithreemcp` is correct and the `DISPLAY` environment variable matches your setup.

3. Run ollmcp with your config:
   ```bash
   uvx ollmcp --servers-json ithreemcp.json
   ```

Now you can chat with Ollama and ask it to manage your i3 windows!

### Running the Server Directly

If you're integrating with a different MCP client, you can run the server directly:

```bash
./ithreemcp
```

The server communicates over stdio using the MCP protocol.

## How it works

The ithreemcp program is an MCP Server, allowing MCP clients to interact with the running i3 window manager.
The ithreemcp program uses the [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) to construct the MCP Server wrapping the functionality.
The ithreemcp program uses the [go-i3](https://github.com/i3/go-i3) library to communicate with the i3 window manager using its IPC interface.

## Supported Operations

### GetTree
Returns the i3 layout tree, using the exact format returned by the underlying operation.

### GetWorkspaces
Returns details about i3's current workspaces.

### FindWindows
Searches for windows matching the given criteria. Returns matching windows with their `con_id`, which can be used with `RunCommand`.

**Parameters:**
- `name` (optional) - Match window title (case-insensitive substring)
- `class` (optional) - Match window class (e.g., "firefox", "Alacritty")
- `instance` (optional) - Match window instance

**Returns:** List of windows with `con_id`, `name`, `class`, `instance`, `workspace`, and `focused` status.

### RunCommand
Executes an i3 command. Use [i3 command syntax](https://i3wm.org/docs/userguide.html#command_criteria) for criteria and actions.

**Parameters:**
- `command` (required) - The i3 command to execute

**Returns:** Array of results with `success` status and optional `error` message.

## Examples

The `FindWindows` and `RunCommand` tools work together to enable powerful window management through natural language requests.

### Move a window to another workspace

**User:** "Move Firefox to workspace 7"

1. `FindWindows(class: "firefox")` → finds the window and returns its `con_id`
2. `RunCommand(command: "[con_id=94285673947] move to workspace 7")`

### Focus a specific application

**User:** "Switch to my terminal"

1. `FindWindows(class: "Alacritty")` → returns matching terminals
2. `RunCommand(command: "[con_id=94285673948] focus")`

### Close all windows of an application

**User:** "Close all my browser windows"

1. `FindWindows(class: "firefox")` → returns all Firefox windows
2. For each window: `RunCommand(command: "[con_id=...] kill")`

### Move window to scratchpad

**User:** "Hide Slack in the scratchpad"

1. `FindWindows(class: "Slack")` → finds Slack window
2. `RunCommand(command: "[con_id=94285673949] move scratchpad")`

### Toggle floating mode

**User:** "Make the video player float"

1. `FindWindows(name: "VLC")` → finds VLC by window title
2. `RunCommand(command: "[con_id=94285673950] floating toggle")`

### Resize a window

**User:** "Make my editor wider"

1. `FindWindows(class: "code")` → finds VS Code
2. `RunCommand(command: "[con_id=94285673951] resize grow width 200 px")`

### Move window to a specific output/monitor

**User:** "Move this to my external monitor"

`RunCommand(command: "[focused] move to output HDMI-1")`

### Fullscreen toggle

**User:** "Fullscreen my browser"

1. `FindWindows(class: "firefox")`
2. `RunCommand(command: "[con_id=...] fullscreen toggle")`

### Using i3 criteria directly

For simple cases, `RunCommand` can use i3's built-in criteria without needing `FindWindows`:

```
RunCommand(command: "[class=\"firefox\"] move to workspace 7")
RunCommand(command: "[urgent=latest] focus")
RunCommand(command: "[workspace=3] move to workspace 5")
```

This is useful when targeting windows by class, but `FindWindows` is preferred when you need to:
- Search by partial window title
- Get information about which windows match before acting
- Handle multiple matches individually
- Confirm the target window with the user

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
