package openwa

import (
	"context"
	"encoding/json"
	"net/http"
)

// GetSessionConfig reads tunables without starting or reconnecting a session.
func (c *Client) GetSessionConfig(ctx context.Context, sessionID string) (*EffectiveSessionConfig, error) {
	var raw json.RawMessage
	_, err := c.doJSON(ctx, http.MethodGet, []string{"api", "sessions", sessionID, "config"}, nil, http.StatusOK, &raw)
	if err != nil {
		return nil, err
	}
	return decodeSessionConfig(raw)
}
