package blob

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type LocalStore struct {
	dir string
}

func NewLocal(dir string) *LocalStore {
	return &LocalStore{dir: dir}
}

func (s *LocalStore) EnsureReady(_ context.Context) error {
	return os.MkdirAll(s.dir, 0o750)
}

func (s *LocalStore) Write(_ context.Context, key string, r io.Reader, size int64) (int64, error) {
	if err := s.EnsureReady(context.Background()); err != nil {
		return 0, err
	}
	dest := filepath.Join(s.dir, key)
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return 0, err
	}
	written, err := io.Copy(f, io.LimitReader(r, size+1))
	f.Close()
	if err != nil {
		os.Remove(dest)
		return 0, err
	}
	return written, nil
}

func (s *LocalStore) Open(_ context.Context, key string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(s.dir, key))
}

func (s *LocalStore) Remove(_ context.Context, key string) error {
	return os.Remove(filepath.Join(s.dir, key))
}
