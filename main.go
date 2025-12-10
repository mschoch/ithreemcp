package main

import (
	"context"
	"log"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.i3wm.org/i3/v4"
)

// I3Client abstracts i3 IPC operations for testability
type I3Client interface {
	GetVersion() (i3.Version, error)
	GetTree() (i3.Tree, error)
	GetWorkspaces() ([]i3.Workspace, error)
	RunCommand(command string) ([]i3.CommandResult, error)
}

// defaultI3Client wraps the real i3 package
type defaultI3Client struct{}

func (c *defaultI3Client) GetVersion() (i3.Version, error) {
	return i3.GetVersion()
}

func (c *defaultI3Client) GetTree() (i3.Tree, error) {
	return i3.GetTree()
}

func (c *defaultI3Client) GetWorkspaces() ([]i3.Workspace, error) {
	return i3.GetWorkspaces()
}

func (c *defaultI3Client) RunCommand(command string) ([]i3.CommandResult, error) {
	return i3.RunCommand(command)
}

// i3MCPServer implements the MCP server for i3 window manager
type i3MCPServer struct {
	// mcp.Server implements the basic MCP server functionality
	// and allows us to register operations and handle client connections
	*mcp.Server
	i3client I3Client
}

// WorkspacesOut represents the output of GetWorkspaces
type WorkspacesOut struct {
	Workspaces []i3.Workspace `json:"workspaces"`
}

// FindWindowsIn represents the input parameters for FindWindows
type FindWindowsIn struct {
	Name     string `json:"name,omitempty" jsonschema:"Match window title (case-insensitive substring match)"`
	Class    string `json:"class,omitempty" jsonschema:"Match window class (e.g. firefox, Alacritty)"`
	Instance string `json:"instance,omitempty" jsonschema:"Match window instance"`
}

// WindowInfo represents a found window with relevant details
type WindowInfo struct {
	ConID     i3.NodeID `json:"con_id"`
	Name      string    `json:"name"`
	Class     string    `json:"class"`
	Instance  string    `json:"instance"`
	Workspace string    `json:"workspace"`
	Focused   bool      `json:"focused"`
}

// FindWindowsOut represents the output of FindWindows
type FindWindowsOut struct {
	Windows []WindowInfo `json:"windows"`
}

// RunCommandIn represents the input parameters for RunCommand
type RunCommandIn struct {
	Command string `json:"command" jsonschema:"The i3 command to execute,required"`
}

// CommandResult represents the result of a single command
type CommandResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// RunCommandOut represents the output of RunCommand
type RunCommandOut struct {
	Results []CommandResult `json:"results"`
}

// New creates a new i3MCPServer. If client is nil, uses the real i3 IPC.
func New(client I3Client) (*i3MCPServer, error) {
	if client == nil {
		client = &defaultI3Client{}
	}

	// Check the version to find problems early
	_, err := client.GetVersion()
	if err != nil {
		return nil, err
	}

	// Create a new MCP server
	server := mcp.NewServer(&mcp.Implementation{Name: "i3"}, nil)

	// Create the server instance
	srv := &i3MCPServer{
		Server:   server,
		i3client: client,
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

	mcp.AddTool(server, &mcp.Tool{
		Description: "searches for windows matching the given criteria (name, class, or instance). Returns matching windows with their con_id which can be used with RunCommand.",
		Name:        "FindWindows",
	}, srv.findWindows)

	mcp.AddTool(server, &mcp.Tool{
		Description: "executes an i3 command. Use i3 command syntax, e.g. '[con_id=123] move to workspace 7' or '[class=\"firefox\"] focus'. See https://i3wm.org/docs/userguide.html#command_criteria for criteria syntax.",
		Name:        "RunCommand",
	}, srv.runCommand)

	return srv, nil
}

// getTree returns the i3 layout tree
func (s *i3MCPServer) getTree(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	// Request the tree structure from the i3 window manager
	tree, err := s.i3client.GetTree()
	if err != nil {
		return nil, i3.Tree{}, err
	}

	return nil, tree, nil
}

// getWorkspaces returns details about i3's current workspaces
func (s *i3MCPServer) getWorkspaces(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, WorkspacesOut, error) {
	// Request the workspace information from the i3 window manager
	workspaces, err := s.i3client.GetWorkspaces()
	if err != nil {
		return nil, WorkspacesOut{}, err
	}

	return nil, WorkspacesOut{Workspaces: workspaces}, nil
}

// findWindows searches for windows matching the given criteria
func (s *i3MCPServer) findWindows(_ context.Context, _ *mcp.CallToolRequest, in FindWindowsIn) (*mcp.CallToolResult, FindWindowsOut, error) {
	tree, err := s.i3client.GetTree()
	if err != nil {
		return nil, FindWindowsOut{}, err
	}

	var windows []WindowInfo
	findWindowsRecursive(tree.Root, "", in, &windows)

	return nil, FindWindowsOut{Windows: windows}, nil
}

// findWindowsRecursive traverses the tree to find matching windows
func findWindowsRecursive(node *i3.Node, workspace string, criteria FindWindowsIn, results *[]WindowInfo) {
	if node == nil {
		return
	}

	// Track current workspace name
	currentWorkspace := workspace
	if node.Type == i3.WorkspaceNode {
		currentWorkspace = node.Name
	}

	// Check if this is a window (con with X11 window ID)
	if node.Window != 0 {
		props := node.WindowProperties

		// Check if window matches criteria
		matches := true

		if criteria.Name != "" && !containsIgnoreCase(node.Name, criteria.Name) {
			matches = false
		}
		if criteria.Class != "" && !containsIgnoreCase(props.Class, criteria.Class) {
			matches = false
		}
		if criteria.Instance != "" && !containsIgnoreCase(props.Instance, criteria.Instance) {
			matches = false
		}

		if matches {
			*results = append(*results, WindowInfo{
				ConID:     node.ID,
				Name:      node.Name,
				Class:     props.Class,
				Instance:  props.Instance,
				Workspace: currentWorkspace,
				Focused:   node.Focused,
			})
		}
	}

	// Recurse into children
	for _, child := range node.Nodes {
		findWindowsRecursive(child, currentWorkspace, criteria, results)
	}
	for _, child := range node.FloatingNodes {
		findWindowsRecursive(child, currentWorkspace, criteria, results)
	}
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// runCommand executes an i3 command
func (s *i3MCPServer) runCommand(_ context.Context, _ *mcp.CallToolRequest, in RunCommandIn) (*mcp.CallToolResult, RunCommandOut, error) {
	results, err := s.i3client.RunCommand(in.Command)
	if err != nil {
		return nil, RunCommandOut{}, err
	}

	cmdResults := make([]CommandResult, 0, len(results))
	for _, r := range results {
		cmdResults = append(cmdResults, CommandResult{Success: r.Success, Error: r.Error})
	}

	return nil, RunCommandOut{Results: cmdResults}, nil
}

// Run starts the MCP server
func (s *i3MCPServer) Run(ctx context.Context, t mcp.Transport) error {
	// Start the server
	log.Println("Starting i3 MCP server...")

	// Start the server and wait for connections
	// This will block until the server stops
	return s.Server.Run(ctx, t)
}

// Close cleans up server resources.
func (s *i3MCPServer) Close() error {
	return nil
}

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
