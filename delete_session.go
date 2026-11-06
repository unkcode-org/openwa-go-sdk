package openwa

import (
	"context"
	"net/http"
)

// DeleteSession explicitly deletes a session. A 404/409 is preserved; the SDK
// never retries, stops, logs out or recreates a session implicitly.
func (c *Client) DeleteSession(ctx context.Context, sessionID string) error {
	_, err := c.doJSON(ctx, http.MethodDelete, []string{"api", "sessions", sessionID}, nil, http.StatusNoContent, nil)
	return err
}
