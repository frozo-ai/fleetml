package model

import (
	"fmt"
	"strings"
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

func TestHotSwapper_Active_BeforeAnySwap(t *testing.T) {
	hs := NewHotSwapper()

	active := hs.Active()
	if active != nil {
		t.Fatalf("expected Active() to return nil before any swap, got %+v", active)
	}
}

func TestHotSwapper_Rollback_NoRollbackAvailable(t *testing.T) {
	hs := NewHotSwapper()

	// Verify HasRollback is false
	if hs.HasRollback() {
		t.Fatal("expected HasRollback() = false on fresh swapper")
	}

	// Rollback should return an error
	err := hs.Rollback()
	if err == nil {
		t.Fatal("expected error when calling Rollback() with no rollback available")
	}
	if !strings.Contains(err.Error(), "no rollback model available") {
		t.Fatalf("expected 'no rollback model available' error, got: %v", err)
	}
}

func TestHotSwapper_Rollback_Chain(t *testing.T) {
	hs := NewHotSwapper()

	vA := &LoadedModel{Model: &Model{Name: "model", Version: "A"}}
	vB := &LoadedModel{Model: &Model{Name: "model", Version: "B"}}

	// Swap A in (no rollback yet, since there was no previous model)
	hs.Swap(vA)
	if hs.HasRollback() {
		t.Fatal("expected no rollback after first swap (nil was previous)")
	}

	// Swap B in (A becomes rollback target)
	hs.Swap(vB)
	if !hs.HasRollback() {
		t.Fatal("expected rollback to be available after second swap")
	}

	// Verify active is B
	active := hs.Active()
	if active.Model.Version != "B" {
		t.Fatalf("expected active version B, got %s", active.Model.Version)
	}

	// Rollback to A
	if err := hs.Rollback(); err != nil {
		t.Fatalf("Rollback() failed: %v", err)
	}

	active = hs.Active()
	if active.Model.Version != "A" {
		t.Fatalf("expected active version A after rollback, got %s", active.Model.Version)
	}

	// After rollback, the Rollback function stores the current (B) as rollback.
	// So HasRollback should still be true (B is now the rollback target).
	if !hs.HasRollback() {
		t.Fatal("expected HasRollback() = true after rollback (B should be stored)")
	}

	// Rollback again should go back to B
	if err := hs.Rollback(); err != nil {
		t.Fatalf("second Rollback() failed: %v", err)
	}
	active = hs.Active()
	if active.Model.Version != "B" {
		t.Fatalf("expected active version B after second rollback, got %s", active.Model.Version)
	}
}

func TestHotSwapper_Swap_NilModel(t *testing.T) {
	hs := NewHotSwapper()

	err := hs.Swap(nil)
	if err == nil {
		t.Fatal("expected error when swapping nil model")
	}
	if !strings.Contains(err.Error(), "cannot swap to nil model") {
		t.Fatalf("expected 'cannot swap to nil model' error, got: %v", err)
	}

	// Active should still be nil (nothing was swapped)
	if hs.Active() != nil {
		t.Fatal("expected Active() to remain nil after failed nil swap")
	}
}

// mockRuntime implements Runtime for testing BackgroundSwap.
type mockRuntime struct {
	name     string
	unloaded bool
}

func (m *mockRuntime) Name() string              { return m.name }
func (m *mockRuntime) Load(path string) error     { return nil }
func (m *mockRuntime) Infer(input []byte) ([]byte, error) { return input, nil }
func (m *mockRuntime) Unload() error              { m.unloaded = true; return nil }
func (m *mockRuntime) IsSupported() bool          { return true }

func TestHotSwapper_BackgroundSwap_LoadFnError(t *testing.T) {
	hs := NewHotSwapper()

	// Set an initial active model so we can verify state is unchanged after failure
	initial := &LoadedModel{Model: &Model{Name: "model", Version: "1.0"}}
	hs.Swap(initial)

	doneCh := make(chan error, 1)

	loadFn := func() (*LoadedModel, error) {
		return nil, fmt.Errorf("download failed: network timeout")
	}

	verifyFn := func(m *LoadedModel) error {
		t.Fatal("verifyFn should not be called when loadFn fails")
		return nil
	}

	hs.BackgroundSwap("model:2.0", loadFn, verifyFn, doneCh)

	select {
	case err := <-doneCh:
		if err == nil {
			t.Fatal("expected error from BackgroundSwap when loadFn fails")
		}
		if !strings.Contains(err.Error(), "background load failed") {
			t.Fatalf("expected 'background load failed' in error, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for BackgroundSwap to complete")
	}

	// Active model should still be the initial one
	active := hs.Active()
	if active == nil {
		t.Fatal("expected active model to remain after failed background swap")
	}
	if active.Model.Version != "1.0" {
		t.Fatalf("expected active version 1.0 to remain, got %s", active.Model.Version)
	}

	// Progress should show SwapFailed
	progress := hs.Progress()
	if progress.State != SwapFailed {
		t.Fatalf("expected SwapFailed state, got %d", progress.State)
	}
}
