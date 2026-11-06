package openwa

import (
	"context"
	"net/http"
)

// StopSession explicitly disconnects the engine without unlinking credentials.
// No retry or fallback is performed, including after incomplete teardown.
func (c *Client) StopSession(ctx context.Context, sessionID string) (*Session, error) {
	var response sessionResponse
	_, err := c.doJSON(ctx, http.MethodPost, []string{"api", "sessions", sessionID, "stop"}, nil, http.StatusOK, &response)
	if err != nil {
		return nil, err
	}
	return response.session(sessionID)
}
