package main

import "go.i3wm.org/i3/v4"

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
