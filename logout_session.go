package openwa

import (
	"context"
	"net/http"
)

// LogoutSession explicitly unlinks the device and clears credentials. A fresh link is required after success.
// No retry or fallback is performed, including after incomplete teardown.
func (c *Client) LogoutSession(ctx context.Context, sessionID string) (*Session, error) {
	var response sessionResponse
	_, err := c.doJSON(ctx, http.MethodPost, []string{"api", "sessions", sessionID, "logout"}, nil, http.StatusOK, &response)
	if err != nil {
		return nil, err
	}
	return response.session(sessionID)
}
