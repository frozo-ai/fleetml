package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ObjectStore defines the interface for object storage operations.
type ObjectStore interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}

// S3Store implements ObjectStore using MinIO/S3-compatible storage.
type S3Store struct {
	client *minio.Client
	bucket string
}

// NewS3Store creates a new S3-compatible object store.
func NewS3Store(endpoint, accessKey, secretKey, bucket, region string) (*S3Store, error) {
	// Parse endpoint to determine SSL
	// Handle endpoints with or without scheme (e.g. "minio:9000" or "http://minio:9000")
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}

	useSSL := u.Scheme == "https"
	host := u.Host
	if host == "" {
		// No scheme provided — treat entire endpoint as host:port
		host = endpoint
		useSSL = false
	}

	client, err := minio.New(host, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	return &S3Store{
		client: client,
		bucket: bucket,
	}, nil
}

// EnsureBucket creates the bucket if it doesn't exist.
func (s *S3Store) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
	}
	return nil
}

// Upload uploads an object to the store.
func (s *S3Store) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}
	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, opts)
	if err != nil {
		return fmt.Errorf("upload object: %w", err)
	}
	return nil
}

// Download downloads an object from the store.
func (s *S3Store) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("download object: %w", err)
	}
	return obj, nil
}

// GeneratePresignedURL generates a pre-signed download URL.
func (s *S3Store) GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("generate presigned url: %w", err)
	}
	return presignedURL.String(), nil
}

// Delete removes an object from the store.
func (s *S3Store) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}
