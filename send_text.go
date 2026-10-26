package openwa

import (
	"context"
	"net/http"
	"strings"
	"unicode/utf8"
)

// SendText waits for gateway acceptance and never automatically retries.
func (c *Client) SendText(ctx context.Context, sessionID string, input TextMessage) (*MessageResult, error) {
	if strings.TrimSpace(input.ChatID) == "" || strings.TrimSpace(input.Text) == "" || !utf8.ValidString(input.Text) || utf8.RuneCountInString(input.Text) > 4096 {
		return nil, ErrInvalidInput
	}
	if p := input.CustomLinkPreview; p != nil {
		if !validHTTPURL(p.URL) || strings.TrimSpace(p.Title) == "" || utf8.RuneCountInString(p.Title) > 256 || utf8.RuneCountInString(p.Description) > 1024 ||
			(input.LinkPreview != nil && !*input.LinkPreview) {
			return nil, ErrInvalidInput
		}
	}
	return c.sendMessage(ctx, sessionID, "send-text", input, http.StatusCreated)
}
