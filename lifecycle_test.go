package openwa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestExplicitLifecycleAndPairing(t *testing.T) {
	for _, action := range []string{"stop", "logout", "delete", "pairing-code"} {
		t.Run(action, func(t *testing.T) {
			calls := 0
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				path := "/api/sessions/session-1"
				method := "POST"
				if action == "delete" {
					method = "DELETE"
				} else {
					path += "/" + action
				}
				if r.Method != method || r.URL.Path != path {
					t.Errorf("wrong request %s %s", r.Method, r.URL.Path)
				}
				if action == "delete" {
					w.WriteHeader(204)
					return
				}
				if action == "pairing-code" {
					var input PairingCodeInput
					_ = json.NewDecoder(r.Body).Decode(&input)
					if input.PhoneNumber != "5493510000000" {
						t.Error("wrong phone")
					}
					w.WriteHeader(201)
					fmt.Fprint(w, `{"pairingCode":"ABCD1234","status":"qr_ready"}`)
					return
				}
				fmt.Fprint(w, strings.Replace(sessionJSON, `"status":"ready"`, `"status":"disconnected"`, 1))
			})
			var err error
			switch action {
			case "stop":
				_, err = c.StopSession(context.Background(), "session-1")
			case "logout":
				_, err = c.LogoutSession(context.Background(), "session-1")
			case "delete":
				err = c.DeleteSession(context.Background(), "session-1")
			case "pairing-code":
				_, err = c.RequestPairingCode(context.Background(), "session-1", PairingCodeInput{PhoneNumber: "5493510000000"})
			}
			if err != nil || calls != 1 {
				t.Fatalf("calls=%d err=%v", calls, err)
			}
		})
	}
}
func TestLifecycleRejectionsNeverRetry(t *testing.T) {
	for _, status := range []int{400, 404, 409, 502} {
		calls := 0
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(status) })
		actions := []func() error{func() error { _, e := c.StopSession(context.Background(), "session-1"); return e }, func() error { _, e := c.LogoutSession(context.Background(), "session-1"); return e }, func() error { return c.DeleteSession(context.Background(), "session-1") }, func() error {
			_, e := c.RequestPairingCode(context.Background(), "session-1", PairingCodeInput{PhoneNumber: "549351"})
			return e
		}}
		for _, action := range actions {
			var api *APIError
			if e := action(); !errors.As(e, &api) || api.StatusCode != status {
				t.Fatalf("lost status: %v", e)
			}
		}
		if calls != 4 {
			t.Fatal("retried destructive operation")
		}
	}
}
func TestProxyAndConfigPatch(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]json.RawMessage
		_ = json.NewDecoder(r.Body).Decode(&body)
		switch r.URL.Path {
		case "/api/sessions":
			if string(body["proxyUrl"]) != `"socks5://user:pass@proxy:1080"` || string(body["proxyType"]) != `"socks5"` {
				t.Error("proxy lost")
			}
			w.WriteHeader(201)
			fmt.Fprint(w, sessionJSON)
		case "/api/sessions/session-1/config":
			if r.Method == "PATCH" {
				if string(body["maxReconnectAttempts"]) != "null" || string(body["autoRejectCalls"]) != "false" || body["reconnectBaseDelay"] != nil {
					t.Errorf("patch lost null/false/omission: %s", body)
				}
			}
			fmt.Fprint(w, `{"autoRejectCalls":false,"maxReconnectAttempts":null,"reconnectBaseDelay":5000}`)
		}
	})
	if _, e := c.CreateSession(context.Background(), CreateSessionInput{Name: "test-session", ProxyURL: "socks5://user:pass@proxy:1080", ProxyType: ProxyTypeSOCKS5}); e != nil {
		t.Fatal(e)
	}
	b := false
	cfg, e := c.UpdateSessionConfig(context.Background(), "session-1", SessionConfigPatch{AutoRejectCalls: &ConfigValue[bool]{Value: &b}, MaxReconnectAttempts: &ConfigValue[int]{}})
	if e != nil || cfg.MaxReconnectAttempts != nil || cfg.AutoRejectCalls {
		t.Fatalf("cfg=%v err=%v", cfg, e)
	}
	if _, e = c.GetSessionConfig(context.Background(), "session-1"); e != nil {
		t.Fatal(e)
	}
}
func TestLifecycleInputAndResponseValidation(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid request dispatched") })
	for _, phone := range []string{"", "+549351", "12a", "12345", "1234567890123456"} {
		if _, e := c.RequestPairingCode(context.Background(), "session-1", PairingCodeInput{PhoneNumber: phone}); !errors.Is(e, ErrInvalidInput) {
			t.Fatal(e)
		}
	}
	for _, input := range []CreateSessionInput{{Name: "test", ProxyURL: "file:///secret"}, {Name: "test", ProxyType: ProxyTypeHTTP}, {Name: "test", ProxyURL: "http://proxy:80", ProxyType: ProxyTypeSOCKS5}} {
		if _, e := c.CreateSession(context.Background(), input); !errors.Is(e, ErrInvalidInput) {
			t.Fatal(e)
		}
	}
	bad := 0
	if _, e := c.UpdateSessionConfig(context.Background(), "session-1", SessionConfigPatch{ReconnectBaseDelay: &ConfigValue[int]{Value: &bad}}); !errors.Is(e, ErrInvalidInput) {
		t.Fatal(e)
	}
	for _, body := range []string{`{}`, `{"autoRejectCalls":false,"reconnectBaseDelay":5000}`, `{"autoRejectCalls":false,"maxReconnectAttempts":21,"reconnectBaseDelay":5000}`} {
		c = testClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) })
		if _, e := c.GetSessionConfig(context.Background(), "session-1"); !errors.Is(e, ErrInvalidResponse) {
			t.Fatalf("body=%s err=%v", body, e)
		}
	}
}
