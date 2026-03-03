package model

import (
	"sync"
	"testing"
	"time"
)

func TestHotSwap_SwapNil(t *testing.T) {
	hs := NewHotSwapper()
	if err := hs.Swap(nil); err == nil {
		t.Fatal("expected error when swapping to nil")
	}
}

func TestHotSwap_SwapAndActive(t *testing.T) {
	hs := NewHotSwapper()

	m := &LoadedModel{
		Model: &Model{
			Name:    "test",
			Version: "1.0",
		},
	}

	if err := hs.Swap(m); err != nil {
		t.Fatal(err)
	}

	active := hs.Active()
	if active == nil {
		t.Fatal("expected active model after swap")
	}
	if active.Model.Name != "test" {
		t.Fatalf("expected model name 'test', got '%s'", active.Model.Name)
	}
}

func TestHotSwap_SwapPreservesOldAsRollback(t *testing.T) {
	hs := NewHotSwapper()

	v1 := &LoadedModel{Model: &Model{Name: "model", Version: "1.0"}}
	v2 := &LoadedModel{Model: &Model{Name: "model", Version: "2.0"}}

	hs.Swap(v1)
	hs.Swap(v2)

	active := hs.Active()
	if active.Model.Version != "2.0" {
		t.Fatalf("expected active version 2.0, got %s", active.Model.Version)
	}

	if !hs.HasRollback() {
		t.Fatal("expected rollback to be available")
	}
}

func TestHotSwap_Rollback(t *testing.T) {
	hs := NewHotSwapper()

	v1 := &LoadedModel{Model: &Model{Name: "model", Version: "1.0"}}
	v2 := &LoadedModel{Model: &Model{Name: "model", Version: "2.0"}}

	hs.Swap(v1)
	hs.Swap(v2)

	if err := hs.Rollback(); err != nil {
		t.Fatal(err)
	}

	active := hs.Active()
	if active.Model.Version != "1.0" {
		t.Fatalf("expected version 1.0 after rollback, got %s", active.Model.Version)
	}
}

func TestHotSwap_RollbackNoModel(t *testing.T) {
	hs := NewHotSwapper()
	if err := hs.Rollback(); err == nil {
		t.Fatal("expected error when no rollback model available")
	}
}

func TestHotSwap_ConcurrentAccess(t *testing.T) {
	hs := NewHotSwapper()

	initial := &LoadedModel{Model: &Model{Name: "model", Version: "1.0"}}
	hs.Swap(initial)

	var wg sync.WaitGroup
	errs := make(chan error, 100)

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				m := hs.Active()
				if m == nil {
					errs <- nil // nil is ok during swap
				}
			}
		}()
	}

	// Concurrent swaps
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			m := &LoadedModel{Model: &Model{
				Name:    "model",
				Version: time.Now().String(),
			}}
			if err := hs.Swap(m); err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent error: %v", err)
		}
	}
}

func TestHotSwap_HasRollbackInitially(t *testing.T) {
	hs := NewHotSwapper()
	if hs.HasRollback() {
		t.Fatal("expected no rollback initially")
	}
}
