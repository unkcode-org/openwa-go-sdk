package openwa

import (
	"context"
	"math/rand"
	"time"
)

// Run supervises one explicitly owned, desired-active session until cancellation.
// It reconciles immediately and then polls, slowing repeated failures/recovery
// attempts with capped exponential backoff and jitter. OnChange fires on state
// transitions. Missing/unlinked/restricted sessions are observed, never recreated.
// It returns ctx.Err() after in-flight work ends; no goroutines are left behind.
func (s *Supervisor) Run(ctx context.Context, id string) error {
	if ctx == nil || !validSegment(id) {
		return ErrInvalidInput
	}
	s.mu.Lock()
	if s.active[id] {
		s.mu.Unlock()
		return ErrAlreadySupervised
	}
	s.active[id] = true
	s.mu.Unlock()
	defer func() { s.mu.Lock(); delete(s.active, id); s.mu.Unlock() }()
	delay := s.config.PollInterval
	previous := ConnectionState("")
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		result, err := s.Reconcile(ctx, id)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if result.State != previous {
			previous = result.State
			if s.config.OnChange != nil {
				s.config.OnChange(result)
			}
		}
		if err == nil && result.State == ConnectionReady {
			delay = s.config.PollInterval
		} else if delay < s.config.MaxBackoff {
			if delay > s.config.MaxBackoff/2 {
				delay = s.config.MaxBackoff
			} else {
				delay *= 2
			}
		}
		// Jitter downwards keeps the configured cap an actual upper bound.
		wait := delay/2 + time.Duration(rand.Int63n(int64(delay-delay/2)+1))
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}
