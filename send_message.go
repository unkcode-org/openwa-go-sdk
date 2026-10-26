package openwa

import (
	"context"
	"net/http"
	"strings"
)

func (c *Client) sendMessage(ctx context.Context, sessionID, operation string, input any, expected int) (*MessageResult, error) {
	var response struct {
		MessageID string   `json:"messageId"`
		Timestamp *float64 `json:"timestamp"`
	}
	dispatched, err := c.doJSON(ctx, http.MethodPost, []string{"api", "sessions", sessionID, "messages", operation}, input, expected, &response)
	if err != nil {
		return nil, sendError(err, dispatched)
	}
	if strings.TrimSpace(response.MessageID) == "" || response.Timestamp == nil || *response.Timestamp < 0 {
		return nil, sendError(ErrInvalidResponse, dispatched)
	}
	return &MessageResult{MessageID: response.MessageID, Timestamp: *response.Timestamp}, nil
}
