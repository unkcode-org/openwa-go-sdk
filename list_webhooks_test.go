package openwa

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

const webhookJSON = `{"id":"hook-1","sessionId":"session-1","url":"https://example.com/receipts","events":["message.ack"],"active":false,"retryCount":0,"lastTriggeredAt":null,"createdAt":"2026-11-14T03:00:00Z","updatedAt":"2026-11-14T03:00:00Z"}`

func TestListWebhooks(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		wantCount  int
		wantError  bool
	}{
		{"empty", "[]", 0, false},
		{"a subscription", "[" + webhookJSON + "]", 1, false},
		{"foreign session", "[" + strings.Replace(webhookJSON, "session-1", "session-2", 1) + "]", 0, true},
		{"null member", "[null]", 0, true},
		{"wrong response shape", webhookJSON, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/api/sessions/session-1/webhooks" {
					t.Error("wrong list endpoint")
				}
				fmt.Fprint(w, tc.body)
			})
			hooks, err := c.ListWebhooks(context.Background(), "session-1")
			if tc.wantError {
				if !errors.Is(err, ErrInvalidResponse) || hooks != nil {
					t.Fatalf("hooks=%v, err=%v", hooks, err)
				}
				return
			}
			if err != nil || hooks == nil || len(hooks) != tc.wantCount {
				t.Fatalf("hooks=%v, err=%v", hooks, err)
			}
			if len(hooks) == 1 && (hooks[0].ID != "hook-1" || hooks[0].Active || hooks[0].RetryCount != 0) {
				t.Fatalf("wrong subscription: %+v", hooks[0])
			}
		})
	}
}
