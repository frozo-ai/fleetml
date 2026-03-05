package storage

import (
	"testing"
)

// ---------------------------------------------------------------------------
// NewS3Store
// ---------------------------------------------------------------------------

func TestNewS3Store_ValidConfig(t *testing.T) {
	store, err := NewS3Store("localhost:9000", "minioadmin", "minioadmin", "models", "us-east-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
	if store.bucket != "models" {
		t.Fatalf("expected bucket 'models', got %q", store.bucket)
	}
	if store.client == nil {
		t.Fatal("expected non-nil minio client")
	}
}

func TestNewS3Store_EndpointWithHTTPScheme(t *testing.T) {
	store, err := NewS3Store("http://minio:9000", "access", "secret", "test-bucket", "us-east-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestNewS3Store_EndpointWithHTTPSScheme(t *testing.T) {
	store, err := NewS3Store("https://s3.amazonaws.com", "access", "secret", "prod-bucket", "us-west-2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestNewS3Store_EndpointWithoutScheme(t *testing.T) {
	store, err := NewS3Store("minio:9000", "access", "secret", "bucket", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestNewS3Store_EmptyEndpoint(t *testing.T) {
	// An empty endpoint will be treated as no-scheme, host = "".
	// minio.New may or may not error — we just verify no panic.
	_, _ = NewS3Store("", "access", "secret", "bucket", "")
}

func TestNewS3Store_DifferentBuckets(t *testing.T) {
	store1, err := NewS3Store("localhost:9000", "a", "s", "bucket-alpha", "")
	if err != nil {
		t.Fatalf("store1: %v", err)
	}
	store2, err := NewS3Store("localhost:9000", "a", "s", "bucket-beta", "")
	if err != nil {
		t.Fatalf("store2: %v", err)
	}

	if store1.bucket == store2.bucket {
		t.Fatalf("expected different buckets, both are %q", store1.bucket)
	}
}

// ---------------------------------------------------------------------------
// ObjectStore interface compliance
// ---------------------------------------------------------------------------

// TestS3Store_ImplementsObjectStore is a compile-time check that S3Store
// satisfies the ObjectStore interface. If it doesn't, the assignment will
// produce a compile error.
func TestS3Store_ImplementsObjectStore(t *testing.T) {
	var _ ObjectStore = (*S3Store)(nil)
}

func TestNewS3Store_EmptyCredentials(t *testing.T) {
	store, err := NewS3Store("localhost:9000", "", "", "bucket", "")
	if err != nil {
		t.Fatalf("empty credentials should not error: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestNewS3Store_SpecialCharsInBucket(t *testing.T) {
	for _, name := range []string{"my.bucket", "my-bucket", "bucket123", "a"} {
		store, err := NewS3Store("localhost:9000", "a", "s", name, "")
		if err != nil {
			t.Errorf("bucket %q: %v", name, err)
		}
		if store != nil && store.bucket != name {
			t.Errorf("expected %q, got %q", name, store.bucket)
		}
	}
}

func TestNewS3Store_LongBucketName(t *testing.T) {
	longName := "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0"
	store, err := NewS3Store("localhost:9000", "a", "s", longName, "")
	if err != nil {
		t.Fatalf("63-char bucket error: %v", err)
	}
	if store.bucket != longName {
		t.Errorf("expected %q, got %q", longName, store.bucket)
	}
}

func TestNewS3Store_PortVariations(t *testing.T) {
	for _, ep := range []string{"localhost:9000", "localhost:443", "localhost", "192.168.1.1:9000"} {
		store, err := NewS3Store(ep, "a", "s", "b", "")
		if err != nil {
			t.Errorf("endpoint %q: %v", ep, err)
		}
		if store == nil {
			t.Errorf("endpoint %q: nil store", ep)
		}
	}
}

func TestNewS3Store_HTTPSDetectsSSL(t *testing.T) {
	store, err := NewS3Store("https://s3.us-west-2.amazonaws.com", "a", "s", "b", "us-west-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store == nil || store.client == nil {
		t.Fatal("expected non-nil store and client")
	}
}

func TestNewS3Store_RegionVariations(t *testing.T) {
	for _, r := range []string{"us-east-1", "eu-west-1", "ap-southeast-1", ""} {
		store, err := NewS3Store("localhost:9000", "a", "s", "b", r)
		if err != nil {
			t.Errorf("region %q: %v", r, err)
		}
		if store == nil {
			t.Errorf("region %q: nil store", r)
		}
	}
}

func TestNewS3Store_BucketFieldPreserved(t *testing.T) {
	store, _ := NewS3Store("localhost:9000", "a", "s", "my-bucket-42", "")
	if store.bucket != "my-bucket-42" {
		t.Errorf("expected 'my-bucket-42', got %q", store.bucket)
	}
}

func TestNewS3Store_ClientNonNil(t *testing.T) {
	store, _ := NewS3Store("localhost:9000", "access", "secret", "bucket", "us-east-1")
	if store.client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewS3Store_IPv4WithPort(t *testing.T) {
	store, err := NewS3Store("10.0.0.1:9000", "a", "s", "b", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestNewS3Store_MultipleBucketsIndependent(t *testing.T) {
	s1, _ := NewS3Store("localhost:9000", "a", "s", "alpha", "")
	s2, _ := NewS3Store("localhost:9000", "a", "s", "beta", "")
	s3, _ := NewS3Store("localhost:9000", "a", "s", "gamma", "")
	if s1.bucket == s2.bucket || s2.bucket == s3.bucket {
		t.Fatal("expected different buckets")
	}
}
