package library

import (
	"context"
	"fmt"
	"io"
	"strings"

	storagerepo "github.com/maximka-44js/lc/internal/storage/repo"
)

type ObjectReader interface {
	Open(ctx context.Context, key string) (io.ReadCloser, error)
}

type ObjectReaderPort struct {
	storage storagerepo.ObjectStorage
}

func NewObjectReaderPort(storage storagerepo.ObjectStorage) *ObjectReaderPort {
	return &ObjectReaderPort{storage: storage}
}

func (p *ObjectReaderPort) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if p == nil || p.storage == nil {
		return nil, fmt.Errorf("library compress adapter: object storage is nil")
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("library compress adapter: key is required")
	}

	r, err := p.storage.GetObject(ctx, key)
	if err != nil {
		return nil, err
	}

	return r, nil
}
