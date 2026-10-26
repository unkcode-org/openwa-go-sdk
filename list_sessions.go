package openwa

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ListSessions returns one page, never selecting a session for the application.
// A nil option uses the gateway default (up to 1000 sessions). Restricted API keys
// may only see a subset; callers must provision with appropriate permissions.
func (c *Client) ListSessions(ctx context.Context, options *ListSessionsOptions) ([]Session, error) {
	query := make(url.Values)
	if options != nil {
		if options.Limit < 0 || options.Limit > 1000 || options.Offset < 0 {
			return nil, ErrInvalidInput
		}
		if options.Limit != 0 {
			query.Set("limit", strconv.Itoa(options.Limit))
		}
		if options.Offset != 0 {
			query.Set("offset", strconv.Itoa(options.Offset))
		}
	}
	var response []sessionResponse
	_, err := c.doJSONQuery(ctx, http.MethodGet, []string{"api", "sessions"}, query, nil, http.StatusOK, &response)
	if err != nil {
		return nil, err
	}
	result := make([]Session, 0, len(response))
	for _, wire := range response {
		session, err := wire.session("")
		if err != nil {
			return nil, err
		}
		result = append(result, *session)
	}
	return result, nil
}
