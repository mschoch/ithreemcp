package main

import (
	"testing"

	"go.i3wm.org/i3/v4"
)

// mockI3Client implements I3Client for testing
type mockI3Client struct {
	tree       i3.Tree
	workspaces []i3.Workspace
	cmdResults []i3.CommandResult
	cmdErr     error
}

func (m *mockI3Client) GetVersion() (i3.Version, error) {
	return i3.Version{Major: 4, Minor: 23}, nil
}

func (m *mockI3Client) GetTree() (i3.Tree, error) {
	return m.tree, nil
}

func (m *mockI3Client) GetWorkspaces() ([]i3.Workspace, error) {
	return m.workspaces, nil
}

func (m *mockI3Client) RunCommand(cmd string) ([]i3.CommandResult, error) {
	return m.cmdResults, m.cmdErr
}

// newTestServer creates an i3MCPServer without registering MCP tools (avoids schema validation)
func newTestServer(client I3Client) *i3MCPServer {
	return &i3MCPServer{
		i3client: client,
	}
}

func TestFindWindowsByClass(t *testing.T) {
	mock := &mockI3Client{
		tree: i3.Tree{
			Root: &i3.Node{
				Type: i3.WorkspaceNode,
				Name: "1",
				Nodes: []*i3.Node{
					{
						Window: 12345,
						ID:     100,
						Name:   "Mozilla Firefox",
						WindowProperties: i3.WindowProperties{
							Class:    "firefox",
							Instance: "Navigator",
						},
					},
					{
						Window: 12346,
						ID:     101,
						Name:   "Terminal",
						WindowProperties: i3.WindowProperties{
							Class:    "Alacritty",
							Instance: "alacritty",
						},
					},
				},
			},
		},
	}

	srv := newTestServer(mock)

	_, out, err := srv.findWindows(nil, nil, FindWindowsIn{
		Class: "firefox",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Windows) != 1 {
		t.Errorf("expected 1 window, got %d", len(out.Windows))
	}
	if out.Windows[0].Class != "firefox" {
		t.Errorf("expected class 'firefox', got %q", out.Windows[0].Class)
	}
	if out.Windows[0].Name != "Mozilla Firefox" {
		t.Errorf("expected name 'Mozilla Firefox', got %q", out.Windows[0].Name)
	}
}

func TestFindWindowsByName(t *testing.T) {
	mock := &mockI3Client{
		tree: i3.Tree{
			Root: &i3.Node{
				Type: i3.WorkspaceNode,
				Name: "1",
				Nodes: []*i3.Node{
					{
						Window: 12345,
						ID:     100,
						Name:   "Mozilla Firefox",
						WindowProperties: i3.WindowProperties{
							Class:    "firefox",
							Instance: "Navigator",
						},
					},
					{
						Window: 12346,
						ID:     101,
						Name:   "Terminal",
						WindowProperties: i3.WindowProperties{
							Class:    "Alacritty",
							Instance: "alacritty",
						},
					},
				},
			},
		},
	}

	srv := newTestServer(mock)

	// Case-insensitive search
	_, out, err := srv.findWindows(nil, nil, FindWindowsIn{
		Name: "terminal",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Windows) != 1 {
		t.Errorf("expected 1 window, got %d", len(out.Windows))
	}
	if out.Windows[0].Class != "Alacritty" {
		t.Errorf("expected class 'Alacritty', got %q", out.Windows[0].Class)
	}
}

func TestFindWindowsNoCriteria(t *testing.T) {
	mock := &mockI3Client{
		tree: i3.Tree{
			Root: &i3.Node{
				Type: i3.WorkspaceNode,
				Name: "1",
				Nodes: []*i3.Node{
					{
						Window: 12345,
						ID:     100,
						Name:   "Mozilla Firefox",
						WindowProperties: i3.WindowProperties{
							Class:    "firefox",
							Instance: "Navigator",
						},
					},
					{
						Window: 12346,
						ID:     101,
						Name:   "Terminal",
						WindowProperties: i3.WindowProperties{
							Class:    "Alacritty",
							Instance: "alacritty",
						},
					},
				},
			},
		},
	}

	srv := newTestServer(mock)

	// No criteria returns all windows
	_, out, err := srv.findWindows(nil, nil, FindWindowsIn{})
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Windows) != 2 {
		t.Errorf("expected 2 windows, got %d", len(out.Windows))
	}
}

func TestFindWindowsNestedTree(t *testing.T) {
	mock := &mockI3Client{
		tree: i3.Tree{
			Root: &i3.Node{
				Type: i3.Root,
				Nodes: []*i3.Node{
					{
						Type: i3.OutputNode,
						Name: "eDP-1",
						Nodes: []*i3.Node{
							{
								Type: i3.WorkspaceNode,
								Name: "1",
								Nodes: []*i3.Node{
									{
										Window: 12345,
										ID:     100,
										Name:   "Firefox",
										WindowProperties: i3.WindowProperties{
											Class: "firefox",
										},
									},
								},
							},
							{
								Type: i3.WorkspaceNode,
								Name: "2",
								Nodes: []*i3.Node{
									{
										Window: 12346,
										ID:     101,
										Name:   "Code",
										WindowProperties: i3.WindowProperties{
											Class: "Code",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	srv := newTestServer(mock)

	_, out, err := srv.findWindows(nil, nil, FindWindowsIn{
		Class: "Code",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Windows) != 1 {
		t.Errorf("expected 1 window, got %d", len(out.Windows))
	}
	if out.Windows[0].Workspace != "2" {
		t.Errorf("expected workspace '2', got %q", out.Windows[0].Workspace)
	}
}

func TestGetWorkspaces(t *testing.T) {
	mock := &mockI3Client{
		workspaces: []i3.Workspace{
			{Name: "1", Num: 1, Visible: true, Focused: true},
			{Name: "2", Num: 2, Visible: false, Focused: false},
		},
	}

	srv := newTestServer(mock)

	_, out, err := srv.getWorkspaces(nil, nil, struct{}{})
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Workspaces) != 2 {
		t.Errorf("expected 2 workspaces, got %d", len(out.Workspaces))
	}
	if out.Workspaces[0].Name != "1" {
		t.Errorf("expected workspace name '1', got %q", out.Workspaces[0].Name)
	}
}

func TestRunCommand(t *testing.T) {
	mock := &mockI3Client{
		cmdResults: []i3.CommandResult{
			{Success: true},
		},
	}

	srv := newTestServer(mock)

	_, out, err := srv.runCommand(nil, nil, RunCommandIn{
		Command: "[con_id=123] focus",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(out.Results))
	}
	if !out.Results[0].Success {
		t.Error("expected command to succeed")
	}
}

func TestRunCommandError(t *testing.T) {
	mock := &mockI3Client{
		cmdResults: []i3.CommandResult{
			{Success: false, Error: "No matching container"},
		},
	}

	srv := newTestServer(mock)

	_, out, err := srv.runCommand(nil, nil, RunCommandIn{
		Command: "[con_id=999] focus",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(out.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(out.Results))
	}
	if out.Results[0].Success {
		t.Error("expected command to fail")
	}
	if out.Results[0].Error != "No matching container" {
		t.Errorf("expected error 'No matching container', got %q", out.Results[0].Error)
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"Firefox", "fire", true},
		{"Firefox", "FIRE", true},
		{"Firefox", "fox", true},
		{"Firefox", "chrome", false},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		got := containsIgnoreCase(tt.s, tt.substr)
		if got != tt.want {
			t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}
