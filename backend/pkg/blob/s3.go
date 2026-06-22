package blob

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Store struct {
	client *minio.Client
	bucket string
	prefix string
}

func NewS3FromEnv() (*S3Store, error) {
	endpoint, err := requiredEnv("S3_ENDPOINT")
	if err != nil {
		return nil, err
	}
	bucket, err := requiredEnv("S3_BUCKET")
	if err != nil {
		return nil, err
	}
	accessKey, err := requiredEnv("S3_ACCESS_KEY")
	if err != nil {
		return nil, err
	}
	secretKey, err := requiredEnv("S3_SECRET_KEY")
	if err != nil {
		return nil, err
	}
	useSSL := strings.EqualFold(os.Getenv("S3_USE_SSL"), "true")
	region := strings.TrimSpace(os.Getenv("S3_REGION"))
	if region == "" {
		region = "us-east-1"
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, err
	}
	prefix := strings.Trim(strings.TrimSpace(os.Getenv("S3_PREFIX")), "/")
	return &S3Store{client: client, bucket: bucket, prefix: prefix}, nil
}

func (s *S3Store) objectKey(key string) string {
	if s.prefix == "" {
		return key
	}
	return s.prefix + "/" + key
}

func (s *S3Store) EnsureReady(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("s3 bucket %q does not exist", s.bucket)
	}
	return nil
}

func (s *S3Store) Write(ctx context.Context, key string, r io.Reader, size int64) (int64, error) {
	if err := s.EnsureReady(ctx); err != nil {
		return 0, err
	}
	info, err := s.client.PutObject(ctx, s.bucket, s.objectKey(key), r, size, minio.PutObjectOptions{})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

func (s *S3Store) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, s.objectKey(key), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *S3Store) Remove(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, s.objectKey(key), minio.RemoveObjectOptions{})
}
