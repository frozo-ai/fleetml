package storage

import (
	"context"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// NewPostgresPool
// ---------------------------------------------------------------------------

// TestNewPostgresPool_InvalidConnectionString verifies that an invalid DSN
// returns an error from pgxpool.ParseConfig before any network connection.
func TestNewPostgresPool_InvalidConnectionString(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// A completely malformed connection string should fail at parse time.
	_, err := NewPostgresPool(ctx, "://invalid-url-no-scheme", 5)
	if err == nil {
		t.Fatal("expected error for invalid connection string")
	}
}

// TestNewPostgresPool_UnreachableHost verifies that a syntactically valid
// but unreachable host returns an error (from the Ping step).
func TestNewPostgresPool_UnreachableHost(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// A valid DSN pointing to a host that almost certainly isn't running
	// Postgres should fail either at NewWithConfig or at Ping.
	_, err := NewPostgresPool(ctx, "postgres://user:pass@127.0.0.1:59999/noexist?sslmode=disable", 1)
	if err == nil {
		t.Fatal("expected error when connecting to unreachable host")
	}
}

// ---------------------------------------------------------------------------
// RunMigrations
// ---------------------------------------------------------------------------

// TestRunMigrations_NilPool documents the expected behavior: calling
// RunMigrations with a nil pool will panic (no nil guard in production code).
func TestRunMigrations_NilPool(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil pool")
		}
	}()
	_ = RunMigrations(nil, "/nonexistent/path")
}

func TestNewPostgresPool_ZeroMaxConns(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := NewPostgresPool(ctx, "postgres://user:pass@127.0.0.1:59998/db?sslmode=disable", 0)
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
}

func TestNewPostgresPool_NegativeMaxConns(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := NewPostgresPool(ctx, "postgres://user:pass@127.0.0.1:59997/db?sslmode=disable", -1)
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
}

func TestNewPostgresPool_EmptyURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := NewPostgresPool(ctx, "", 5)
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestNewPostgresPool_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewPostgresPool(ctx, "postgres://user:pass@127.0.0.1:59996/db?sslmode=disable", 5)
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestNewPostgresPool_LargeMaxConns(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := NewPostgresPool(ctx, "postgres://user:pass@127.0.0.1:59995/db?sslmode=disable", 1000)
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
}

func TestNewPostgresPool_DifferentPort(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := NewPostgresPool(ctx, "postgres://admin:admin@127.0.0.1:15432/fleetml?sslmode=disable", 5)
	if err == nil {
		t.Fatal("expected error for unreachable port")
	}
}

func TestNewPostgresPool_SpecialCharsInPassword(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := NewPostgresPool(ctx, "postgres://user:p%40ss%23word@127.0.0.1:59994/db?sslmode=disable", 5)
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
}
