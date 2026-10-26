package openwa

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const sessionJSON = `{"id":"session-1","name":"test-session","status":"ready","phone":"5493510000000","pushName":"Test","connectedAt":"2026-10-26T12:00:00Z","createdAt":"2026-10-26T10:00:00Z","updatedAt":"2026-10-26T12:00:00Z","engineLoaded":true,"restriction":{"kind":"reachout_timelock","code":"BIZ_QUALITY"}}`
const messageJSON = `{"messageId":"message-1","timestamp":1793016000}`

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func testClient(t *testing.T, handler http.HandlerFunc, configure ...func(*Config)) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	cfg := Config{BaseURL: server.URL, APIKey: "test-secret", Timeout: time.Second}
	for _, f := range configure {
		f(&cfg)
	}
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestAPIRejectionsAreBoundedAndRedacted(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404, 408, 409, 413, 415, 429, 500, 501, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var attempts atomic.Int32
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				attempts.Add(1)
				w.Header().Set("Retry-After", "7")
				w.WriteHeader(status)
				fmt.Fprint(w, `{"code":"SESSION_NOT_READY","message":"test-secret private-message private-document"}`)
			})
			_, err := c.SendText(context.Background(), "session-1", TextMessage{ChatID: "549351@c.us", Text: "hello"})
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != status || apiErr.Code != "SESSION_NOT_READY" || apiErr.RetryAfter != "7" {
				t.Fatalf("err=%v", err)
			}
			if errors.Is(err, ErrOutcomeUnknown) != (status >= 500 || status == 408) {
				t.Fatalf("wrong outcome: %v", err)
			}
			if strings.Contains(err.Error(), "test-secret") || strings.Contains(err.Error(), "private-") {
				t.Fatal("sensitive error")
			}
			if attempts.Load() != 1 {
				t.Fatal("automatic retry")
			}
		})
	}
}

func TestNoRedirectCanForwardAPIKey(t *testing.T) {
	for _, status := range []int{301, 302, 303, 307, 308} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var forwarded atomic.Int32
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forwarded.Add(1) }))
			defer target.Close()
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, status) })
			_, err := c.SendText(context.Background(), "session-1", TextMessage{ChatID: "x@c.us", Text: "hello"})
			if !errors.Is(err, ErrRedirect) || !errors.Is(err, ErrOutcomeUnknown) {
				t.Fatalf("err=%v", err)
			}
			if forwarded.Load() != 0 {
				t.Fatal("credentials forwarded")
			}
		})
	}
}

func TestResponseBoundsAndMalformedJSON(t *testing.T) {
	for _, body := range []string{"", "null", "[]", "{", "{} {}", "{} trailing"} {
		t.Run(fmt.Sprintf("body %q", body), func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) })
			if _, err := c.GetSession(context.Background(), "session-1"); !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("err=%v", err)
			}
		})
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, strings.Repeat("x", 33)) }, func(cfg *Config) { cfg.MaxResponseBytes = 32 })
	if _, err := c.GetSession(context.Background(), "session-1"); !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("err=%v", err)
	}
	// An oversized diagnostic must not hide a definitive HTTP rejection.
	c = testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		fmt.Fprint(w, strings.Repeat("x", 33))
	}, func(cfg *Config) { cfg.MaxResponseBytes = 32 })
	_, err := c.GetSession(context.Background(), "session-1")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 403 {
		t.Fatalf("err=%v", err)
	}
}

func TestPathInjectionRejectedBeforeDispatch(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid ID dispatched") })
	for _, id := range []string{"", " ", ".", "..", "../other", "a/b", "a\\b", "%2e%2e", "%2f", "x\n"} {
		_, err := c.SendText(context.Background(), id, TextMessage{ChatID: "x@c.us", Text: "hello"})
		if !errors.Is(err, ErrInvalidInput) || errors.Is(err, ErrOutcomeUnknown) {
			t.Fatalf("id=%q err=%v", id, err)
		}
	}
}

func TestPathPrefixAndReservedCharactersArePreserved(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/gateway/api/sessions/session%3F%23%20id/qr" || r.URL.RawQuery != "" {
			t.Errorf("URI=%s", r.RequestURI)
		}
		fmt.Fprint(w, `{"qrCode":"data:image/png;base64,aGVsbG8=","status":"qr_ready"}`)
	}, func(cfg *Config) { cfg.BaseURL += "/gateway/" })
	if _, err := c.GetQRCode(context.Background(), "session?# id"); err != nil {
		t.Fatal(err)
	}
}

func TestDeadlineAndCancellation(t *testing.T) {
	injected := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	c, err := NewClient(Config{BaseURL: "https://example.com", APIKey: "key", Timeout: 10 * time.Millisecond, HTTPClient: injected})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.SendText(context.Background(), "session-1", TextMessage{ChatID: "x@c.us", Text: "hello"})
	if !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, ErrOutcomeUnknown) || !errors.Is(err, ErrTransport) {
		t.Fatalf("err=%v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = c.SendText(ctx, "session-1", TextMessage{ChatID: "x@c.us", Text: "hello"})
	if !errors.Is(err, context.Canceled) || errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("pre-cancel=%v", err)
	}
	_, err = c.GetSession(nil, "session-1")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("nil context=%v", err)
	}
}

func TestTransportErrorDoesNotPrintSensitiveCause(t *testing.T) {
	c, err := NewClient(Config{BaseURL: "https://example.com/private-base", APIKey: "test-secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("test-secret private-base private-document")
	})}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.GetSession(context.Background(), "session-1")
	if !errors.Is(err, ErrTransport) || strings.Contains(err.Error(), "private-") || strings.Contains(err.Error(), "test-secret") {
		t.Fatalf("err=%v", err)
	}
}

func TestClientConcurrentUse(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, sessionJSON) })
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.GetSession(context.Background(), "session-1"); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}

func TestTruncatedSuccessIsUnknown(t *testing.T) {
	c, err := NewClient(Config{BaseURL: "https://example.com", APIKey: "key", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 201, Header: make(http.Header), Body: io.NopCloser(&brokenReader{})}, nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.SendText(context.Background(), "session-1", TextMessage{ChatID: "x@c.us", Text: "hello"})
	if !errors.Is(err, ErrOutcomeUnknown) || !errors.Is(err, ErrTransport) {
		t.Fatalf("err=%v", err)
	}
}

type brokenReader struct{}

func (*brokenReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestBodyDeadlinePreservesContextCause(t *testing.T) {
	c, err := NewClient(Config{BaseURL: "https://example.com", APIKey: "key", Timeout: 10 * time.Millisecond,
		HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 201, Header: make(http.Header), Body: io.NopCloser(contextReader{ctx: r.Context()})}, nil
		})}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.SendText(context.Background(), "session-1", TextMessage{ChatID: "x@c.us", Text: "hello"})
	if !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, ErrOutcomeUnknown) || !errors.Is(err, ErrTransport) {
		t.Fatalf("err=%v", err)
	}
}

type contextReader struct{ ctx context.Context }

func (r contextReader) Read([]byte) (int, error) {
	<-r.ctx.Done()
	return 0, r.ctx.Err()
}
