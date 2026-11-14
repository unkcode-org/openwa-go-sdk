package openwa

import (
	"context"
	"net/http"
)

// SendImage waits for gateway acceptance. MaxRequestBytes bounds the JSON
// payload, including the encoded image. The gateway applies its own media limits.
func (c *Client) SendImage(ctx context.Context, sessionID string, input ImageMessage) (*MessageResult, error) {
	if int64(len(input.Base64)) > c.maxRequestBytes {
		return nil, ErrRequestTooLarge
	}
	if err := validateMedia(input.ChatID, input.URL, input.Base64, input.MIMEType, input.Caption, c.maxRequestBytes, ErrRequestTooLarge); err != nil {
		return nil, err
	}
	return c.sendMessage(ctx, sessionID, "send-image", input, http.StatusCreated)
}
