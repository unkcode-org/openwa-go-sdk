package openwa

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidConfig    = errors.New("openwa: invalid client configuration")
	ErrInvalidInput     = errors.New("openwa: invalid operation input")
	ErrTransport        = errors.New("openwa: HTTP request failed")
	ErrInvalidResponse  = errors.New("openwa: invalid API response")
	ErrResponseTooLarge = errors.New("openwa: response exceeds configured limit")
	ErrRequestTooLarge  = errors.New("openwa: request exceeds configured limit")
	ErrDocumentTooLarge = errors.New("openwa: document exceeds configured limit")
	ErrRedirect         = errors.New("openwa: redirects are disabled")
	// ErrOutcomeUnknown means that a send may have taken effect. Persist and
	// reconcile the attempt instead of blindly resending it.
	ErrOutcomeUnknown = errors.New("openwa: send outcome is unknown")
)

// APIError contains bounded metadata from an unsuccessful HTTP response.
// Code is untrusted upstream data; do not log it without sanitization.
// Error deliberately omits the upstream body, message, code, URL and credentials.
type APIError struct {
	StatusCode int
	Code       string
	RetryAfter string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("openwa: API returned HTTP %d", e.StatusCode)
}

// SendError wraps an ambiguous result after dispatch. errors.Is with
// ErrOutcomeUnknown identifies it; errors.As still exposes an underlying APIError.
type SendError struct{ cause error }

func (e *SendError) Error() string        { return ErrOutcomeUnknown.Error() }
func (e *SendError) Unwrap() error        { return e.cause }
func (e *SendError) Is(target error) bool { return target == ErrOutcomeUnknown }

// transportError preserves errors.Is/As for cancellation and network errors,
// while avoiding net/http's URL-bearing error message in the default string.
type transportError struct{ cause error }

func (e *transportError) Error() string        { return ErrTransport.Error() }
func (e *transportError) Unwrap() error        { return e.cause }
func (e *transportError) Is(target error) bool { return target == ErrTransport }

func sendError(err error, dispatched bool) error {
	if err == nil || !dispatched {
		return err
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != 408 {
		return err
	}
	return &SendError{cause: err}
}
