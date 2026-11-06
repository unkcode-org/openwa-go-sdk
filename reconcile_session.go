package openwa

import (
	"context"
	"errors"
)

// Reconcile explicitly manages one desired-active session. Call only for owned,
// enabled sessions and while holding any application distributed lease. It may
// start an unloaded engine, but never interrupts a loaded/reconnecting engine.
// Unlike Run, repeated external calls must implement their own backoff.
func (s *Supervisor) Reconcile(ctx context.Context, id string) (SupervisionResult, error) {
	result := SupervisionResult{SessionID: id, State: ConnectionUnavailable}
	if ctx == nil || !validSegment(id) {
		return result, ErrInvalidInput
	}
	ctx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
	defer cancel()
	release, err := s.acquire(ctx, id)
	if err != nil {
		return result, err
	}
	defer release()
	session, err := s.client.GetSession(ctx, id)
	if err != nil {
		var api *APIError
		if errors.As(err, &api) && api.StatusCode == 404 {
			result.State = ConnectionMissing
		}
		return result, err
	}
	if session == nil || session.ID != id {
		return result, ErrInvalidResponse
	}
	result.State = connectionState(session)
	if session.EngineLoaded || result.State == ConnectionNeedsAttention || session.Status == SessionStatusAuthenticating {
		return result, nil
	}
	session, err = s.client.StartSession(ctx, id)
	if err != nil {
		var api *APIError
		if errors.As(err, &api) && (api.StatusCode == 400 || api.StatusCode == 409) {
			// Another caller/owner may have started it after our GET. Re-read once.
			current, readErr := s.client.GetSession(ctx, id)
			if readErr == nil && current != nil && current.ID == id {
				result.State = connectionState(current)
				if current.EngineLoaded {
					return result, nil
				}
			}
		}
		return result, err
	}
	if session == nil || session.ID != id {
		return result, ErrInvalidResponse
	}
	result.State = connectionState(session)
	result.Started = true
	return result, nil
}
