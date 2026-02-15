package attachments

import (
	"context"
	"io"
	"time"
)

type Storage interface {
	Put(ctx context.Context, key string, r io.Reader) (PutResult, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error)
}

type PutResult struct {
	SizeBytes int64
	SHA256Hex string
}
