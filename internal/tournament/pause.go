package tournament

import (
	"context"
	"sync"
)

// PauseController requests a safe pause between fights. The active fight is
// allowed to finish and be persisted before the runner blocks.
type PauseController struct {
	mu     sync.Mutex
	paused bool
	resume chan struct{}
}

// Pause requests that the tournament stop before starting its next fight.
func (controller *PauseController) Pause() {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if controller.paused {
		return
	}
	controller.paused = true
	controller.resume = make(chan struct{})
}

// Resume allows a paused tournament to continue.
func (controller *PauseController) Resume() {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if !controller.paused {
		return
	}
	controller.paused = false
	close(controller.resume)
	controller.resume = nil
}

// Wait blocks while paused and returns when resumed or cancelled.
func (controller *PauseController) Wait(ctx context.Context, onPaused func()) error {
	controller.mu.Lock()
	if !controller.paused {
		controller.mu.Unlock()
		return ctx.Err()
	}
	resume := controller.resume
	controller.mu.Unlock()

	if onPaused != nil {
		onPaused()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-resume:
		return nil
	}
}
