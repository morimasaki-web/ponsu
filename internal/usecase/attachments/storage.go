package attachments

import (
	"context"
	"io"
)

type Storage interface {
	Put(ctx context.Context, key string, r io.Reader) (PutResult, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type PutResult struct {
	SizeBytes int64
	SHA256Hex string
}
