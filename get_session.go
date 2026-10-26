package openwa

import (
	"context"
	"net/http"
)

// GetSession reads an existing session without starting it or fetching its QR.
func (c *Client) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	var result sessionResponse
	_, err := c.doJSON(ctx, http.MethodGet, []string{"api", "sessions", sessionID}, nil, http.StatusOK, &result)
	if err != nil {
		return nil, err
	}
	return result.session(sessionID)
}
