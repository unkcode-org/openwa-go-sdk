package openwa

import (
	"context"
	"net/http"
)

// CheckNumber checks a number as supplied, without applying regional phone rules.
// A lookup failure is an error, never a synthetic Exists=false result.
func (c *Client) CheckNumber(ctx context.Context, sessionID, number string) (*NumberCheckResult, error) {
	var result struct {
		NumberCheckResult
		Exists *bool `json:"exists"`
	}
	_, err := c.doJSON(ctx, http.MethodGet, []string{"api", "sessions", sessionID, "contacts", "check", number}, nil, http.StatusOK, &result)
	if err != nil {
		return nil, err
	}
	if result.Number != number || result.Exists == nil {
		return nil, ErrInvalidResponse
	}
	if *result.Exists && (result.WhatsAppID == nil || *result.WhatsAppID == "") {
		return nil, ErrInvalidResponse
	}
	result.NumberCheckResult.Exists = *result.Exists
	return &result.NumberCheckResult, nil
}
