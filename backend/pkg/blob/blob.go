package blob

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// Store abstracts attachment blob storage (local disk or S3-compatible).
type Store interface {
	EnsureReady(ctx context.Context) error
	Write(ctx context.Context, key string, r io.Reader, size int64) (int64, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Remove(ctx context.Context, key string) error
}

// NewFromEnv selects local or S3 storage from ATTACHMENTS_STORAGE (local|s3).
func NewFromEnv(localDir string) (Store, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("ATTACHMENTS_STORAGE")))
	if mode == "s3" {
		return NewS3FromEnv()
	}
	if localDir == "" {
		localDir = "/data/attachments"
	}
	return NewLocal(localDir), nil
}

func requiredEnv(key string) (string, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return "", fmt.Errorf("%s is required when ATTACHMENTS_STORAGE=s3", key)
	}
	return v, nil
}
