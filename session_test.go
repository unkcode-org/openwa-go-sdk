package openwa

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestSessionOperationsAndHeaders(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				expected := "/api/sessions/session-1"
				if method == http.MethodPost {
					expected += "/start"
				}
				if r.Method != method || r.RequestURI != expected || r.Header.Get("X-API-Key") != "test-secret" || r.Header.Get("Accept") != "application/json" {
					t.Errorf("request=%s %s", r.Method, r.RequestURI)
				}
				fmt.Fprint(w, sessionJSON)
			})
			var result *Session
			var err error
			if method == http.MethodGet {
				result, err = c.GetSession(context.Background(), "session-1")
			} else {
				result, err = c.StartSession(context.Background(), "session-1")
			}
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != SessionStatusReady || !result.EngineLoaded || result.Phone == nil || result.Restriction == nil || result.Restriction.Code != "BIZ_QUALITY" || result.ConnectedAt == nil {
				t.Fatalf("unexpected result: %#v", result)
			}
		})
	}
}

func TestSessionStateDoesNotTriggerLifecycleActions(t *testing.T) {
	for _, state := range []string{"created", "initializing", "qr_ready", "authenticating", "ready", "disconnected", "action_required", "failed", "future_status"} {
		t.Run(state, func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || strings.HasSuffix(r.URL.Path, "/start") {
					t.Error("side effect during status read")
				}
				fmt.Fprint(w, strings.Replace(sessionJSON, `"status":"ready"`, `"status":"`+state+`"`, 1))
			})
			result, err := c.GetSession(context.Background(), "session-1")
			if err != nil || result.Status != SessionStatus(state) {
				t.Fatalf("result=%v err=%v", result, err)
			}
		})
	}
}

func TestMissingRequiredSessionFields(t *testing.T) {
	for _, body := range []string{"{}", strings.Replace(sessionJSON, `"engineLoaded":true,`, "", 1), strings.Replace(sessionJSON, `"id":"session-1"`, `"id":""`, 1), strings.Replace(sessionJSON, `"id":"session-1"`, `"id":"other-session"`, 1), strings.Replace(sessionJSON, `"createdAt":"2026-10-26T10:00:00Z"`, `"createdAt":null`, 1)} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) })
		if _, err := c.GetSession(context.Background(), "session-1"); !errors.Is(err, ErrInvalidResponse) {
			t.Fatalf("err=%v", err)
		}
	}
}

func TestQRAndNumberLookup(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/sessions/session-1/qr":
			fmt.Fprint(w, `{"qrCode":"data:image/png;base64,aGVsbG8=","status":"qr_ready"}`)
		case "/api/sessions/session-1/contacts/check/549351":
			fmt.Fprint(w, `{"number":"549351","exists":true,"whatsappId":"549351@c.us"}`)
		case "/api/sessions/session-1/contacts/check/549352":
			fmt.Fprint(w, `{"number":"549352","exists":false,"whatsappId":null}`)
		default:
			t.Errorf("path=%s", r.URL.Path)
		}
	})
	qr, err := c.GetQRCode(context.Background(), "session-1")
	if err != nil || qr.Status != SessionStatusQRReady {
		t.Fatalf("qr=%v err=%v", qr, err)
	}
	for _, number := range []string{"549351", "549352"} {
		result, err := c.CheckNumber(context.Background(), "session-1", number)
		if err != nil || result.Exists != (number == "549351") {
			t.Fatalf("number=%s result=%v err=%v", number, result, err)
		}
	}
}

func TestLookupFailureNeverBecomesNonexistentNumber(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) })
	result, err := c.CheckNumber(context.Background(), "session-1", "549351")
	if err == nil || result != nil {
		t.Fatal("lookup failure flattened")
	}
	for _, body := range []string{`{"number":"549351"}`, `{"number":"other","exists":false}`, `{"number":"549351","exists":true,"whatsappId":null}`} {
		c = testClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) })
		if _, err = c.CheckNumber(context.Background(), "session-1", "549351"); !errors.Is(err, ErrInvalidResponse) {
			t.Fatalf("err=%v", err)
		}
	}
}

func TestLifecycleAndQRConflictsAreNotSilentlyAccepted(t *testing.T) {
	for _, status := range []int{400, 409} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) })
		if result, err := c.StartSession(context.Background(), "session-1"); err == nil || result != nil {
			t.Fatal("lifecycle rejection treated as success")
		}
		if result, err := c.GetQRCode(context.Background(), "session-1"); err == nil || result != nil {
			t.Fatal("QR rejection treated as success")
		}
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"status":"qr_ready"}`) })
	if _, err := c.GetQRCode(context.Background(), "session-1"); !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("err=%v", err)
	}
}
