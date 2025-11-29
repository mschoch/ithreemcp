# ithreemcp

An MCP server exposing functionality of the connected i3 window manager.

## Building

`go build`

## Running

Prerequisite: You already have the i3 window manager running, and you have permission to access to the socket.

./ithreemcp

## How it works

the ithreemcp program is an MCP Server, allowing MCP clients to interact with the running i3 window manager.
The ithreemcp program uses the [MCP Go SDK](github.com/modelcontextprotocol/go-sdk/mcp) to construct the MCP Server wrapping the functionality.
The ithreemcp program uses the [go-i3](go.i3wm.org/i3/v4) library to communicate with the i3 window manager using its IPC interface.

## Supported Operations

Currently only two operations are supported:

- GetTree - returns the i3 layout tree, using the exact format returned by the underlying operation
- GetWorkspaces - returns details about i3's current workspaces
