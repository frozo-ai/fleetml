package model

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// SwapState represents the state of a background swap operation.
type SwapState int

const (
	SwapIdle       SwapState = iota
	SwapLoading              // Model is being loaded in background
	SwapVerifying            // Model loaded, verifying integrity
	SwapReady                // Verified, ready to atomically swap
	SwapComplete             // Swap completed
	SwapFailed               // Background load or verification failed
)

// SwapProgress reports on a background swap operation.
type SwapProgress struct {
	State   SwapState
	Model   string // "name:version"
	Error   error
}

// HotSwapper performs atomic model swaps with zero dropped inferences.
// Supports background model loading so the active model keeps serving
// throughout the entire download + load + verify cycle.
type HotSwapper struct {
	activeModel   atomic.Pointer[LoadedModel]
	rollbackModel atomic.Pointer[LoadedModel]

	// Background swap tracking
	mu           sync.Mutex
	swapProgress atomic.Pointer[SwapProgress]
}

func NewHotSwapper() *HotSwapper {
	h := &HotSwapper{}
	idle := &SwapProgress{State: SwapIdle}
	h.swapProgress.Store(idle)
	return h
}

// Active returns the currently active model.
func (h *HotSwapper) Active() *LoadedModel {
	return h.activeModel.Load()
}

// Swap atomically switches the active model to newModel.
// The old model becomes the rollback target.
func (h *HotSwapper) Swap(newModel *LoadedModel) error {
	if newModel == nil {
		return fmt.Errorf("cannot swap to nil model")
	}

	old := h.activeModel.Swap(newModel)

	// Save old model as rollback target
	if old != nil {
		h.rollbackModel.Store(old)
	}

	return nil
}

// BackgroundSwap loads and verifies the model in the background, then
// atomically swaps it in. The active model continues serving throughout.
// The provided loadFn should download/load the model and return a LoadedModel.
// The verifyFn should validate the loaded model (e.g., test inference).
func (h *HotSwapper) BackgroundSwap(
	modelName string,
	loadFn func() (*LoadedModel, error),
	verifyFn func(*LoadedModel) error,
	doneCh chan<- error,
) {
	h.mu.Lock()
	current := h.swapProgress.Load()
	if current.State == SwapLoading || current.State == SwapVerifying {
		h.mu.Unlock()
		if doneCh != nil {
			doneCh <- fmt.Errorf("swap already in progress for %s", current.Model)
		}
		return
	}
	h.swapProgress.Store(&SwapProgress{State: SwapLoading, Model: modelName})
	h.mu.Unlock()

	go func() {
		var err error
		defer func() {
			if doneCh != nil {
				doneCh <- err
			}
		}()

		// Phase 1: Load model in background
		newModel, loadErr := loadFn()
		if loadErr != nil {
			err = fmt.Errorf("background load failed: %w", loadErr)
			h.swapProgress.Store(&SwapProgress{State: SwapFailed, Model: modelName, Error: err})
			return
		}

		// Phase 2: Verify model
		h.swapProgress.Store(&SwapProgress{State: SwapVerifying, Model: modelName})
		if verifyFn != nil {
			if verifyErr := verifyFn(newModel); verifyErr != nil {
				err = fmt.Errorf("model verification failed: %w", verifyErr)
				h.swapProgress.Store(&SwapProgress{State: SwapFailed, Model: modelName, Error: err})
				// Unload the failed model
				if newModel.Runtime != nil {
					newModel.Runtime.Unload()
				}
				return
			}
		}

		// Phase 3: Atomic swap
		h.swapProgress.Store(&SwapProgress{State: SwapReady, Model: modelName})
		if swapErr := h.Swap(newModel); swapErr != nil {
			err = fmt.Errorf("atomic swap failed: %w", swapErr)
			h.swapProgress.Store(&SwapProgress{State: SwapFailed, Model: modelName, Error: err})
			return
		}

		h.swapProgress.Store(&SwapProgress{State: SwapComplete, Model: modelName})
	}()
}

// Progress returns the current state of a background swap.
func (h *HotSwapper) Progress() *SwapProgress {
	return h.swapProgress.Load()
}

// Rollback reverts to the previous model.
func (h *HotSwapper) Rollback() error {
	prev := h.rollbackModel.Load()
	if prev == nil {
		return fmt.Errorf("no rollback model available")
	}

	current := h.activeModel.Swap(prev)
	h.rollbackModel.Store(current)

	return nil
}

// HasRollback returns true if a rollback model is available.
func (h *HotSwapper) HasRollback() bool {
	return h.rollbackModel.Load() != nil
}
