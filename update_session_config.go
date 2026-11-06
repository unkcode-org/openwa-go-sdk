package openwa

import (
	"context"
	"encoding/json"
	"net/http"
)

// UpdateSessionConfig merges explicit changes. It never restarts the session.
func (c *Client) UpdateSessionConfig(ctx context.Context, sessionID string, patch SessionConfigPatch) (*EffectiveSessionConfig, error) {
	if v := patch.MaxReconnectAttempts; v != nil && v.Value != nil && (*v.Value < 0 || *v.Value > 20) {
		return nil, ErrInvalidInput
	}
	if v := patch.ReconnectBaseDelay; v != nil && v.Value != nil && (*v.Value < 1000 || *v.Value > 300000) {
		return nil, ErrInvalidInput
	}
	var raw json.RawMessage
	_, err := c.doJSON(ctx, http.MethodPatch, []string{"api", "sessions", sessionID, "config"}, patch, http.StatusOK, &raw)
	if err != nil {
		return nil, err
	}
	return decodeSessionConfig(raw)
}
