package openwa

import (
	"net/http"
	"time"
)

const (
	DefaultTimeout                = 30 * time.Second
	DefaultMaxResponseBytes int64 = 1 << 20
	DefaultMaxRequestBytes  int64 = 16 << 20
	DefaultMaxDocumentBytes int64 = 10 << 20
)

// Config configures one immutable client. Zero limits select the defaults.
// BaseURL is the gateway root, optionally behind a path prefix, without /api.
// HTTP is supported for private networks; use HTTPS across untrusted networks.
type Config struct {
	BaseURL          string
	APIKey           string
	Timeout          time.Duration
	MaxResponseBytes int64
	MaxRequestBytes  int64
	MaxDocumentBytes int64
	// HTTPClient is copied, preserving its Transport and timeout unless Timeout
	// overrides it. Redirects are always disabled. Custom transports must honor
	// context cancellation and must not retry sends or log credentials/bodies.
	HTTPClient *http.Client
}
