package folder

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	storagerepo "github.com/maximka-44js/lc/internal/storage/repo"
)

type ReadAtSeekCloser interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

type ObjectReader interface {
	Open(ctx context.Context, key string) (ReadAtSeekCloser, error)
}

type ObjectReaderPort struct {
	storage storagerepo.ObjectStorage
}

func NewObjectReaderPort(storage storagerepo.ObjectStorage) *ObjectReaderPort {
	return &ObjectReaderPort{storage: storage}
}

func (p *ObjectReaderPort) Open(ctx context.Context, key string) (ReadAtSeekCloser, error) {
	if p == nil || p.storage == nil {
		return nil, fmt.Errorf("folder extract adapter: object storage is nil")
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("folder extract adapter: key is required")
	}

	source, err := p.storage.GetObject(ctx, key)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = source.Close()
	}()

	tempFile, err := os.CreateTemp("", "lc-storage-object-*")
	if err != nil {
		return nil, fmt.Errorf("folder extract adapter: create temp object: %w", err)
	}

	if _, err := io.Copy(tempFile, source); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
		return nil, fmt.Errorf("folder extract adapter: download object: %w", err)
	}
	if _, err := tempFile.Seek(0, io.SeekStart); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
		return nil, fmt.Errorf("folder extract adapter: seek temp object: %w", err)
	}

	return &tempObjectFile{File: tempFile}, nil
}

type tempObjectFile struct {
	*os.File
}

func (f *tempObjectFile) Close() error {
	name := f.Name()
	err := f.File.Close()
	if removeErr := os.Remove(name); err == nil {
		err = removeErr
	}
	return err
}
