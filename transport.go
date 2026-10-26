package openwa

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) doJSON(ctx context.Context, method string, segments []string, input any, expected int, output any) (bool, error) {
	return c.doJSONQuery(ctx, method, segments, nil, input, expected, output)
}

func (c *Client) doJSONQuery(ctx context.Context, method string, segments []string, query url.Values, input any, expected int, output any) (bool, error) {
	if ctx == nil {
		return false, ErrInvalidInput
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	// Carry the timeout in the request context as well as in http.Client.
	// Older Go clients cancel custom RoundTrippers through a legacy channel,
	// which can otherwise turn a client timeout into context.Canceled.
	ctx, cancel := context.WithTimeout(ctx, c.httpClient.Timeout)
	defer cancel()
	for _, segment := range segments {
		if !validSegment(segment) {
			return false, ErrInvalidInput
		}
	}
	u := c.baseURL
	u.Path = strings.TrimRight(u.Path, "/") + "/" + strings.Join(segments, "/")
	u.RawPath = ""
	u.RawQuery = query.Encode()
	var payload []byte
	if input != nil {
		var err error
		payload, err = json.Marshal(input)
		if err != nil {
			return false, ErrInvalidInput
		}
		if int64(len(payload)) > c.maxRequestBytes {
			return false, ErrRequestTooLarge
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(payload))
	if err != nil {
		return false, ErrInvalidInput
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Discard the URL-bearing wrapper, keeping the cause for errors.Is/As.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		// Go 1.22's client-timeout wrapper has no Unwrap/Is method. Preserve
		// the request context's definitive cancellation cause explicitly.
		if ctxErr := ctx.Err(); ctxErr != nil {
			err = ctxErr
		}
		return true, &transportError{cause: err}
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, c.maxResponseBytes+1))
	if resp.StatusCode != expected {
		apiErr := &APIError{StatusCode: resp.StatusCode, RetryAfter: boundedHeader(resp.Header.Get("Retry-After"))}
		if readErr == nil && int64(len(body)) <= c.maxResponseBytes {
			var envelope struct {
				Code string `json:"code"`
			}
			if json.Unmarshal(body, &envelope) == nil && len(envelope.Code) <= 128 {
				apiErr.Code = envelope.Code
			}
		}
		// Keep the HTTP rejection even if its diagnostic body is unreadable or
		// oversized; no body content is exposed to the caller.
		return true, apiErr
	}
	if readErr != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			readErr = ctxErr
		}
		return true, &transportError{cause: readErr}
	}
	if int64(len(body)) > c.maxResponseBytes {
		return true, ErrResponseTooLarge
	}
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' && trimmed[0] != '[' {
		return true, ErrInvalidResponse
	}
	if json.Unmarshal(body, output) != nil {
		return true, ErrInvalidResponse
	}
	return true, nil
}

func boundedHeader(value string) string {
	if len(value) > 128 || containsControl(value) {
		return ""
	}
	return value
}

func validSegment(value string) bool {
	return strings.TrimSpace(value) != "" && value != "." && value != ".." &&
		!strings.ContainsAny(value, "/\\%") && !containsControl(value)
}

func containsControl(value string) bool {
	for _, r := range value {
		if r < 32 || r == 127 {
			return true
		}
	}
	return false
}
