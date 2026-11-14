package openwa

import (
	"context"
	"net/http"
)

// CreateWebhook registers one subscription. If the outcome is unclear, inspect
// ListWebhooks before trying to create it again.
func (c *Client) CreateWebhook(ctx context.Context, sessionID string, input CreateWebhookInput) (*Webhook, error) {
	if !validHTTPURL(input.URL) || !validWebhookOptions(input.Events, input.Secret, input.RetryCount) {
		return nil, ErrInvalidInput
	}

	var response webhookResponse
	if _, err := c.doJSON(ctx, http.MethodPost, []string{"api", "sessions", sessionID, "webhooks"}, input, http.StatusCreated, &response); err != nil {
		return nil, err
	}
	return response.webhook(sessionID, "")
}
