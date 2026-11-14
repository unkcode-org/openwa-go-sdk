package openwa

import (
	"context"
	"net/http"
)

// ListWebhooks lists subscriptions belonging to the specified session.
func (c *Client) ListWebhooks(ctx context.Context, sessionID string) ([]Webhook, error) {
	var response []webhookResponse
	if _, err := c.doJSON(ctx, http.MethodGet, []string{"api", "sessions", sessionID, "webhooks"}, nil, http.StatusOK, &response); err != nil {
		return nil, err
	}

	result := make([]Webhook, 0, len(response))
	for _, wire := range response {
		hook, err := wire.webhook(sessionID, "")
		if err != nil {
			return nil, err
		}
		result = append(result, *hook)
	}
	return result, nil
}
