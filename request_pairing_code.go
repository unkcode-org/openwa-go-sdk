package openwa

import (
	"context"
	"net/http"
	"strings"
)

// RequestPairingCode requests a code for an already started, unlinked session.
func (c *Client) RequestPairingCode(ctx context.Context, sessionID string, input PairingCodeInput) (*PairingCode, error) {
	if len(input.PhoneNumber) < 6 || len(input.PhoneNumber) > 15 {
		return nil, ErrInvalidInput
	}
	for _, ch := range input.PhoneNumber {
		if ch < '0' || ch > '9' {
			return nil, ErrInvalidInput
		}
	}
	var result PairingCode
	_, err := c.doJSON(ctx, http.MethodPost, []string{"api", "sessions", sessionID, "pairing-code"}, input, http.StatusCreated, &result)
	if err != nil {
		return nil, err
	}
	if len(result.PairingCode) != 8 || strings.TrimSpace(string(result.Status)) == "" {
		return nil, ErrInvalidResponse
	}
	return &result, nil
}
