package openwa

import (
	"math"
	"net/http"
	"net/url"
	"strings"
)

// Client can be shared by concurrent goroutines. Its configuration is immutable.
type Client struct {
	baseURL          url.URL
	apiKey           string
	httpClient       http.Client
	maxResponseBytes int64
	maxRequestBytes  int64
	maxDocumentBytes int64
}

// NewClient validates configuration without making any network requests.
func NewClient(cfg Config) (*Client, error) {
	u, err := url.Parse(strings.TrimSpace(cfg.BaseURL))
	if err != nil || u == nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, ErrInvalidConfig
	}
	if u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" || strings.Contains(u.Path, "\\") {
		return nil, ErrInvalidConfig
	}
	for _, segment := range strings.Split(u.Path, "/") {
		if segment == "." || segment == ".." || strings.Contains(segment, "%") {
			return nil, ErrInvalidConfig
		}
	}
	if strings.TrimSpace(cfg.APIKey) == "" || strings.TrimSpace(cfg.APIKey) != cfg.APIKey || containsControl(cfg.APIKey) {
		return nil, ErrInvalidConfig
	}
	if cfg.Timeout < 0 || invalidLimit(cfg.MaxResponseBytes) || invalidLimit(cfg.MaxRequestBytes) || invalidLimit(cfg.MaxDocumentBytes) {
		return nil, ErrInvalidConfig
	}
	c := &Client{baseURL: *u, apiKey: cfg.APIKey,
		maxResponseBytes: defaultLimit(cfg.MaxResponseBytes, DefaultMaxResponseBytes),
		maxRequestBytes:  defaultLimit(cfg.MaxRequestBytes, DefaultMaxRequestBytes),
		maxDocumentBytes: defaultLimit(cfg.MaxDocumentBytes, DefaultMaxDocumentBytes)}
	if cfg.HTTPClient != nil {
		c.httpClient = *cfg.HTTPClient
	}
	if c.httpClient.Timeout < 0 {
		return nil, ErrInvalidConfig
	}
	if cfg.Timeout != 0 {
		c.httpClient.Timeout = cfg.Timeout
	}
	if c.httpClient.Timeout == 0 {
		c.httpClient.Timeout = DefaultTimeout
	}
	// A custom header such as X-API-Key must never follow a redirect, including
	// a redirect to a different host. Do not mutate an injected client.
	c.httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return ErrRedirect }
	return c, nil
}

func invalidLimit(n int64) bool { return n < 0 || n == math.MaxInt64 }

func defaultLimit(n, fallback int64) int64 {
	if n == 0 {
		return fallback
	}
	return n
}
