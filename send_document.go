package openwa

import (
	"context"
	"encoding/base64"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"
)

// SendDocument waits for gateway acceptance. The decoded byte limit applies to
// Base64 documents; gateway-fetched URLs are limited by the gateway itself.
func (c *Client) SendDocument(ctx context.Context, sessionID string, input DocumentMessage) (*MessageResult, error) {
	if strings.TrimSpace(input.ChatID) == "" || (input.Base64 == "") == (input.URL == "") || !utf8.ValidString(input.Caption) || utf8.RuneCountInString(input.Caption) > 1024 {
		return nil, ErrInvalidInput
	}
	if input.Base64 != "" {
		// Check encoded size before allocating a decoded buffer. CR/LF and data
		// URL prefixes are not accepted; callers pass standard base64 bytes.
		if strings.ContainsAny(input.Base64, "\r\n") {
			return nil, ErrInvalidInput
		}
		n := int64(base64.StdEncoding.DecodedLen(len(input.Base64)))
		if strings.HasSuffix(input.Base64, "==") {
			n -= 2
		} else if strings.HasSuffix(input.Base64, "=") {
			n--
		}
		if n > c.maxDocumentBytes {
			return nil, ErrDocumentTooLarge
		}
		data, err := base64.StdEncoding.Strict().DecodeString(input.Base64)
		if err != nil || len(data) == 0 {
			return nil, ErrInvalidInput
		}
		if int64(len(data)) > c.maxDocumentBytes {
			return nil, ErrDocumentTooLarge
		}
		if _, _, err := mime.ParseMediaType(input.MIMEType); err != nil {
			return nil, ErrInvalidInput
		}
	} else if !validHTTPURL(input.URL) {
		return nil, ErrInvalidInput
	}
	return c.sendMessage(ctx, sessionID, "send-document", input, http.StatusCreated)
}
