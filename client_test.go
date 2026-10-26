package openwa

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestClientConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*Config)
	}{
		{"empty URL", func(c *Config) { c.BaseURL = "" }},
		{"relative URL", func(c *Config) { c.BaseURL = "/gateway" }},
		{"scheme", func(c *Config) { c.BaseURL = "file://host/gateway" }},
		{"credentials", func(c *Config) { c.BaseURL = "https://user:password@example.com" }},
		{"query", func(c *Config) { c.BaseURL += "?token=private" }},
		{"empty query", func(c *Config) { c.BaseURL += "?" }},
		{"fragment", func(c *Config) { c.BaseURL += "#private" }},
		{"traversal", func(c *Config) { c.BaseURL += "/../private" }},
		{"encoded traversal", func(c *Config) { c.BaseURL += "/%2e%2e/private" }},
		{"empty key", func(c *Config) { c.APIKey = " " }},
		{"header injection", func(c *Config) { c.APIKey = "key\r\nX-Other: value" }},
		{"negative timeout", func(c *Config) { c.Timeout = -time.Second }},
		{"negative response limit", func(c *Config) { c.MaxResponseBytes = -1 }},
		{"negative request limit", func(c *Config) { c.MaxRequestBytes = -1 }},
		{"negative document limit", func(c *Config) { c.MaxDocumentBytes = -1 }},
		{"negative injected timeout", func(c *Config) { c.HTTPClient = &http.Client{Timeout: -time.Second} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{BaseURL: "https://example.com", APIKey: "key"}
			tc.change(&cfg)
			if _, err := NewClient(cfg); !errors.Is(err, ErrInvalidConfig) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestClientDefaultsAndInjectedClient(t *testing.T) {
	injected := &http.Client{Timeout: 4 * time.Second}
	originalRedirect := func(*http.Request, []*http.Request) error { return nil }
	injected.CheckRedirect = originalRedirect
	c, err := NewClient(Config{BaseURL: "https://example.com/gateway/", APIKey: "key", HTTPClient: injected, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if injected.Timeout != 4*time.Second || injected.CheckRedirect(nil, nil) != nil {
		t.Fatal("injected client changed")
	}
	if c.httpClient.Timeout != time.Second || !errors.Is(c.httpClient.CheckRedirect(nil, nil), ErrRedirect) {
		t.Fatal("client policies not applied")
	}
	c, err = NewClient(Config{BaseURL: "http://example.com", APIKey: "key"})
	if err != nil {
		t.Fatal(err)
	}
	if c.httpClient.Timeout != DefaultTimeout || c.maxDocumentBytes != DefaultMaxDocumentBytes || c.maxResponseBytes != DefaultMaxResponseBytes || c.maxRequestBytes != DefaultMaxRequestBytes {
		t.Fatal("wrong defaults")
	}
}
