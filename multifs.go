package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"golang.org/x/net/webdav"
)

// MultiFS manages multiple filesystems and executes operations across them with rollback support.
type MultiFS struct {
	filesystems []webdav.FileSystem
	mu          sync.Mutex
}

func (fs *MultiFS) executeWithRollback(ctx context.Context, op Operation) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	succeeded := make([]int, 0)

	for i, filesystem := range fs.filesystems {
		if err := op.Execute(ctx, filesystem); err != nil {
			for _, successIndex := range succeeded {
				if rollbackErr := op.Rollback(ctx, fs.filesystems[successIndex]); rollbackErr != nil {
					log.Printf("Rollback failed for filesystem %d: %v", successIndex, rollbackErr)
				}
			}
			return fmt.Errorf("operation failed on filesystem %d: %v", i, err)
		}
		succeeded = append(succeeded, i)
	}

	return nil
}

func (fs *MultiFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	op := &MkdirOperation{Name: name, Perm: perm}
	return fs.executeWithRollback(ctx, op)
}

func (fs *MultiFS) RemoveAll(ctx context.Context, name string) error {
	info, err := fs.Stat(ctx, name)
	isDir := err == nil && info.IsDir()

	op := &RemoveOperation{Name: name, IsDir: isDir}
	return fs.executeWithRollback(ctx, op)
}

func (fs *MultiFS) Rename(ctx context.Context, oldName, newName string) error {
	op := &RenameOperation{OldName: oldName, NewName: newName}
	return fs.executeWithRollback(ctx, op)
}

func (fs *MultiFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	var files []webdav.File

	for i, filesystem := range fs.filesystems {
		f, err := filesystem.OpenFile(ctx, name, flag, perm)
		if err != nil {
			for _, openFile := range files {
				openFile.Close()
			}
			return nil, fmt.Errorf("failed to open file in filesystem %d: %v", i, err)
		}
		files = append(files, f)
	}

	return &multiFile{files: files, name: name}, nil
}

func (fs *MultiFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return fs.filesystems[0].Stat(ctx, name)
}
