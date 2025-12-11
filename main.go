package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Create a new i3 MCP server
	srv, err := New(nil)
	if err != nil {
		log.Fatalf("Failed to create i3 MCP server: %v", err)
	}

	// Make sure to close the server when done
	defer func() {
		if err := srv.Close(); err != nil {
			log.Printf("Error closing i3 MCP server: %v", err)
		}
	}()

	// Start the server
	log.Println("i3 MCP server started. Press Ctrl+C to stop.")
	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Printf("i3 MCP server terminated with error: %v", err)
	}
}
