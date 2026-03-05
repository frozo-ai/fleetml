package model

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestBackgroundSwap_Success(t *testing.T) {
	h := NewHotSwapper()

	// Set initial model
	initial := &LoadedModel{
		Model:   &Model{Name: "m1", Version: "1.0"},
		Runtime: NewONNXRuntime(),
	}
	h.Swap(initial)

	doneCh := make(chan error, 1)
	h.BackgroundSwap("m1:2.0",
		func() (*LoadedModel, error) {
			return &LoadedModel{
				Model:   &Model{Name: "m1", Version: "2.0"},
				Runtime: NewONNXRuntime(),
			}, nil
		},
		nil, // no verify
		doneCh,
	)

	err := <-doneCh
	if err != nil {
		t.Fatalf("background swap failed: %v", err)
	}

	active := h.Active()
	if active == nil || active.Model.Version != "2.0" {
		t.Error("expected active model v2.0")
	}

	if h.Progress().State != SwapComplete {
		t.Errorf("expected SwapComplete, got %v", h.Progress().State)
	}
}

func TestBackgroundSwap_LoadFailure(t *testing.T) {
	h := NewHotSwapper()

	initial := &LoadedModel{
		Model:   &Model{Name: "m1", Version: "1.0"},
		Runtime: NewONNXRuntime(),
	}
	h.Swap(initial)

	doneCh := make(chan error, 1)
	h.BackgroundSwap("m1:2.0",
		func() (*LoadedModel, error) {
			return nil, fmt.Errorf("download failed")
		},
		nil,
		doneCh,
	)

	err := <-doneCh
	if err == nil {
		t.Fatal("expected error from failed load")
	}

	// Original model should still be active
	active := h.Active()
	if active == nil || active.Model.Version != "1.0" {
		t.Error("expected original model v1.0 to still be active")
	}

	if h.Progress().State != SwapFailed {
		t.Errorf("expected SwapFailed, got %v", h.Progress().State)
	}
}

func TestBackgroundSwap_VerifyFailure(t *testing.T) {
	h := NewHotSwapper()

	initial := &LoadedModel{
		Model:   &Model{Name: "m1", Version: "1.0"},
		Runtime: NewONNXRuntime(),
	}
	h.Swap(initial)

	doneCh := make(chan error, 1)
	h.BackgroundSwap("m1:2.0",
		func() (*LoadedModel, error) {
			return &LoadedModel{
				Model:   &Model{Name: "m1", Version: "2.0"},
				Runtime: NewONNXRuntime(),
			}, nil
		},
		func(lm *LoadedModel) error {
			return fmt.Errorf("inference test failed")
		},
		doneCh,
	)

	err := <-doneCh
	if err == nil {
		t.Fatal("expected error from failed verification")
	}

	// Original model should still be active
	active := h.Active()
	if active == nil || active.Model.Version != "1.0" {
		t.Error("expected original model v1.0 to still be active after verify failure")
	}
}

func TestBackgroundSwap_ConcurrentInference(t *testing.T) {
	h := NewHotSwapper()

	initial := &LoadedModel{
		Model:   &Model{Name: "m1", Version: "1.0"},
		Runtime: NewONNXRuntime(),
	}
	h.Swap(initial)

	// Start concurrent "inference" readers
	var wg sync.WaitGroup
	stop := make(chan struct{})
	inferenceCount := 0
	var mu sync.Mutex

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					active := h.Active()
					if active != nil {
						active.Runtime.Infer([]byte("input"))
						mu.Lock()
						inferenceCount++
						mu.Unlock()
					}
				}
			}
		}()
	}

	// Perform background swap while readers are running
	doneCh := make(chan error, 1)
	h.BackgroundSwap("m1:2.0",
		func() (*LoadedModel, error) {
			time.Sleep(10 * time.Millisecond)
			return &LoadedModel{
				Model:   &Model{Name: "m1", Version: "2.0"},
				Runtime: NewONNXRuntime(),
			}, nil
		},
		nil,
		doneCh,
	)

	<-doneCh
	close(stop)
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if inferenceCount == 0 {
		t.Error("expected some inferences during swap")
	}

	active := h.Active()
	if active == nil || active.Model.Version != "2.0" {
		t.Error("expected new model v2.0 after swap")
	}
}

func TestBackgroundSwap_DuplicateRejected(t *testing.T) {
	h := NewHotSwapper()

	// Start a slow swap
	firstDone := make(chan error, 1)
	h.BackgroundSwap("m1:2.0",
		func() (*LoadedModel, error) {
			time.Sleep(100 * time.Millisecond)
			return &LoadedModel{
				Model:   &Model{Name: "m1", Version: "2.0"},
				Runtime: NewONNXRuntime(),
			}, nil
		},
		nil,
		firstDone,
	)

	// Try to start another swap immediately
	time.Sleep(5 * time.Millisecond) // Let the first swap start
	secondDone := make(chan error, 1)
	h.BackgroundSwap("m1:3.0",
		func() (*LoadedModel, error) {
			return &LoadedModel{
				Model:   &Model{Name: "m1", Version: "3.0"},
				Runtime: NewONNXRuntime(),
			}, nil
		},
		nil,
		secondDone,
	)

	// Second swap should fail
	err := <-secondDone
	if err == nil {
		t.Error("expected duplicate swap to be rejected")
	}

	// First swap should succeed
	err = <-firstDone
	if err != nil {
		t.Fatalf("first swap should succeed: %v", err)
	}
}

func TestProgressStates(t *testing.T) {
	h := NewHotSwapper()

	if h.Progress().State != SwapIdle {
		t.Errorf("expected SwapIdle initially, got %v", h.Progress().State)
	}
}
