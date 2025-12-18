package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Parse command-line flags
	var debugPath string
	flag.StringVar(&debugPath, "debug", "", "path to log file for MCP request/response logging")
	flag.Parse()

	// Create a new i3 MCP server
	srv, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create i3 MCP server: %v", err)
	}

	// Make sure to close the server when done
	defer func() {
		if err := srv.Close(); err != nil {
			log.Printf("Error closing i3 MCP server: %v", err)
		}
	}()

	// Set up the transport, optionally wrapping with logging
	var transport mcp.Transport = &mcp.StdioTransport{}
	if debugPath != "" {
		debugFile, err := os.OpenFile(debugPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Failed to open debug log file: %v", err)
		}
		defer debugFile.Close()
		transport = &mcp.LoggingTransport{
			Transport: transport,
			Writer:    debugFile,
		}
		log.Printf("Debug logging enabled, writing to: %s", debugPath)
	}

	// Start the server
	log.Println("i3 MCP server started. Press Ctrl+C to stop.")
	if err := srv.Run(context.Background(), transport); err != nil {
		log.Printf("i3 MCP server terminated with error: %v", err)
	}
}
