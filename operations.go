package main

import (
	"context"
	"io"
	"os"

	"golang.org/x/net/webdav"
)

// Operation interface defines methods for executing and rolling back file operations.
type Operation interface {
	Execute(ctx context.Context, fs webdav.FileSystem) error
	Rollback(ctx context.Context, fs webdav.FileSystem) error
}

// MkdirOperation represents an operation to create a directory.
type MkdirOperation struct {
	Name string
	Perm os.FileMode
}

func (op *MkdirOperation) Execute(ctx context.Context, fs webdav.FileSystem) error {
	return fs.Mkdir(ctx, op.Name, op.Perm)
}

func (op *MkdirOperation) Rollback(ctx context.Context, fs webdav.FileSystem) error {
	return fs.RemoveAll(ctx, op.Name)
}

// RemoveOperation represents an operation to remove a file or directory.
type RemoveOperation struct {
	Name     string
	OldState []byte
	IsDir    bool
}

func (op *RemoveOperation) Execute(ctx context.Context, fs webdav.FileSystem) error {
	if !op.IsDir {
		file, err := fs.OpenFile(ctx, op.Name, os.O_RDONLY, 0)
		if err == nil {
			defer file.Close()
			op.OldState, _ = io.ReadAll(file)
		}
	}
	return fs.RemoveAll(ctx, op.Name)
}

func (op *RemoveOperation) Rollback(ctx context.Context, fs webdav.FileSystem) error {
	if op.IsDir {
		return fs.Mkdir(ctx, op.Name, 0755)
	}
	file, err := fs.OpenFile(ctx, op.Name, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(op.OldState)
	return err
}

// RenameOperation represents an operation to rename a file or directory.
type RenameOperation struct {
	OldName string
	NewName string
}

func (op *RenameOperation) Execute(ctx context.Context, fs webdav.FileSystem) error {
	return fs.Rename(ctx, op.OldName, op.NewName)
}

func (op *RenameOperation) Rollback(ctx context.Context, fs webdav.FileSystem) error {
	return fs.Rename(ctx, op.NewName, op.OldName)
}
