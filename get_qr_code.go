package openwa

import (
	"context"
	"net/http"
	"strings"
)

// GetQRCode reads the QR without starting the session or waiting for a scan.
// OpenWA may return HTTP 400 while the QR is not ready or already authenticated.
func (c *Client) GetQRCode(ctx context.Context, sessionID string) (*QRCode, error) {
	var result QRCode
	_, err := c.doJSON(ctx, http.MethodGet, []string{"api", "sessions", sessionID, "qr"}, nil, http.StatusOK, &result)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(result.QRCode) == "" || strings.TrimSpace(string(result.Status)) == "" {
		return nil, ErrInvalidResponse
	}
	return &result, nil
}
