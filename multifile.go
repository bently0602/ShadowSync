package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/webdav"
)

// multiFile represents a file across multiple filesystems with rollback support for writes.
type multiFile struct {
	files     []webdav.File
	name      string
	writeLog  [][]byte
	positions []int64
	mu        sync.Mutex
}

func (mf *multiFile) Close() error {
	var errors []string
	for i, f := range mf.files {
		if err := f.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("failed to close file %d: %v", i, err))
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}

func (mf *multiFile) Read(p []byte) (n int, err error) {
	return mf.files[0].Read(p)
}

func (mf *multiFile) Seek(offset int64, whence int) (int64, error) {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	var lastPos int64
	var errors []string

	for i, f := range mf.files {
		pos, err := f.Seek(offset, whence)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to seek file %d: %v", i, err))
		}
		lastPos = pos
	}

	if len(errors) > 0 {
		return lastPos, fmt.Errorf(strings.Join(errors, "; "))
	}
	return lastPos, nil
}

func (mf *multiFile) Write(p []byte) (n int, err error) {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	mf.writeLog = append(mf.writeLog, make([]byte, len(p)))
	copy(mf.writeLog[len(mf.writeLog)-1], p)

	positions := make([]int64, len(mf.files))
	for i, f := range mf.files {
		if pos, err := f.Seek(0, io.SeekCurrent); err == nil {
			positions[i] = pos
		}
	}
	mf.positions = positions

	succeeded := make([]int, 0)
	var lastN int

	for i, f := range mf.files {
		n, err := f.Write(p)
		if err != nil {
			for _, succIndex := range succeeded {
				mf.files[succIndex].Seek(positions[succIndex], io.SeekStart)
				if trunc, ok := mf.files[succIndex].(*os.File); ok {
					trunc.Truncate(positions[succIndex])
				}
			}
			return n, fmt.Errorf("write failed on file %d: %v", i, err)
		}
		succeeded = append(succeeded, i)
		lastN = n
	}

	return lastN, nil
}

func (mf *multiFile) Readdir(count int) ([]os.FileInfo, error) {
	return mf.files[0].Readdir(count)
}

func (mf *multiFile) Stat() (os.FileInfo, error) {
	return mf.files[0].Stat()
}
