package openwa

import (
	"context"
	"net/http"
)

// CreateSession creates a named session. Names are unique in the gateway. Callers
// should persist the intended name before creating and recover it with
// ListSessions after a lost response. Creation does not link a WhatsApp account.
func (c *Client) CreateSession(ctx context.Context, input CreateSessionInput) (*Session, error) {
	if len(input.Name) < 3 || len(input.Name) > 50 {
		return nil, ErrInvalidInput
	}
	for _, r := range input.Name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-') {
			return nil, ErrInvalidInput
		}
	}
	if cfg := input.Config; cfg != nil {
		if cfg.MaxReconnectAttempts != nil && (*cfg.MaxReconnectAttempts < 0 || *cfg.MaxReconnectAttempts > 20) {
			return nil, ErrInvalidInput
		}
		if cfg.ReconnectBaseDelay != nil && (*cfg.ReconnectBaseDelay < 1000 || *cfg.ReconnectBaseDelay > 300000) {
			return nil, ErrInvalidInput
		}
	}
	if !validProxy(input.ProxyURL, input.ProxyType) {
		return nil, ErrInvalidInput
	}
	var response sessionResponse
	_, err := c.doJSON(ctx, http.MethodPost, []string{"api", "sessions"}, input, http.StatusCreated, &response)
	if err != nil {
		return nil, err
	}
	if response.Name != input.Name {
		return nil, ErrInvalidResponse
	}
	return response.session("")
}
