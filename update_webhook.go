package openwa

import (
	"context"
	"net/http"
)

// UpdateWebhook updates the supplied properties of an existing subscription.
func (c *Client) UpdateWebhook(ctx context.Context, sessionID, webhookID string, input UpdateWebhookInput) (*Webhook, error) {
	var secret string
	if input.Secret != nil {
		secret = *input.Secret
	}
	if (input.URL != nil && !validHTTPURL(*input.URL)) || !validWebhookOptions(input.Events, secret, input.RetryCount) {
		return nil, ErrInvalidInput
	}

	var response webhookResponse
	if _, err := c.doJSON(ctx, http.MethodPut, []string{"api", "sessions", sessionID, "webhooks", webhookID}, input, http.StatusOK, &response); err != nil {
		return nil, err
	}
	return response.webhook(sessionID, webhookID)
}
