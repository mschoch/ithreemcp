package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.i3wm.org/i3/v4"
)

// i3MCPServer implements the MCP server for i3 window manager
type i3MCPServer struct {
	// mcp.Server implements the basic MCP server functionality
	// and allows us to register operations and handle client connections
	*mcp.Server
}

// New creates a new i3MCPServer
func New() (*i3MCPServer, error) {
	// Check the version to find problems early
	_, err := i3.GetVersion()

	if err != nil {
		return nil, err
	}

	// Create a new MCP server
	server := mcp.NewServer(&mcp.Implementation{Name: "i3"}, nil)

	// Create the server instance
	srv := &i3MCPServer{
		Server: server,
	}

	// Register the supported operations
	// This will make them available to MCP clients
	mcp.AddTool(server, &mcp.Tool{
		Description: "gets the i3 layout tree",
		Name:        "GetTree",
	}, srv.getTree)

	mcp.AddTool(server, &mcp.Tool{
		Description: "gets the details about i3's current workspaces",
		Name:        "GetWorkspaces",
	}, srv.getWorkspaces)

	return srv, nil
}

// getTree returns the i3 layout tree
func (s *i3MCPServer) getTree(ctx context.Context, request *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
	// Request the tree structure from the i3 window manager
	tree, err := i3.GetTree()
	if err != nil {
		return nil, i3.Tree{}, err
	}

	return nil, tree, nil
}

type WorkspacesOut struct {
	workspaces []i3.Workspace
}

// getWorkspaces returns details about i3's current workspaces
func (s *i3MCPServer) getWorkspaces(ctx context.Context, request *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, WorkspacesOut, error) {
	// Request the workspace information from the i3 window manager
	workspaces, err := i3.GetWorkspaces()
	if err != nil {
		return nil, WorkspacesOut{}, err
	}

	return nil, WorkspacesOut{workspaces: workspaces}, nil
}

// Run starts the MCP server
func (s *i3MCPServer) Run(ctx context.Context, t mcp.Transport) error {
	// Start the server
	log.Println("Starting i3 MCP server...")

	// Start the server and wait for connections
	// This will block until the server stops
	return s.Server.Run(ctx, t)
}

// Close does nothing
func (s *i3MCPServer) Close() error {
	return nil
}

// main function to start the server
func main() {
	// Create a new i3 MCP server
	srv, err := New()
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
