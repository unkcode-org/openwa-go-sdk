package openwa

import (
	"context"
	"net/http"
)

// StartSession starts an existing session. Inspect EngineLoaded first: a loaded
// engine may be reconnecting. A lifecycle conflict is returned to the caller.
func (c *Client) StartSession(ctx context.Context, sessionID string) (*Session, error) {
	var result sessionResponse
	_, err := c.doJSON(ctx, http.MethodPost, []string{"api", "sessions", sessionID, "start"}, nil, http.StatusOK, &result)
	if err != nil {
		return nil, err
	}
	return result.session(sessionID)
}
