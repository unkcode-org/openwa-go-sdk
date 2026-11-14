package openwa

import (
	"context"
	"net/http"
)

// SendDocument waits for gateway acceptance. The decoded byte limit applies to
// Base64 documents; gateway-fetched URLs are limited by the gateway itself.
func (c *Client) SendDocument(ctx context.Context, sessionID string, input DocumentMessage) (*MessageResult, error) {
	if err := validateMedia(input.ChatID, input.URL, input.Base64, input.MIMEType, input.Caption, c.maxDocumentBytes, ErrDocumentTooLarge); err != nil {
		return nil, err
	}
	return c.sendMessage(ctx, sessionID, "send-document", input, http.StatusCreated)
}
