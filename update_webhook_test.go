package openwa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestUpdateWebhookKeepsExplicitZeroValues(t *testing.T) {
	active, secret, retry := false, "", 0
	headers := map[string]string{}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/sessions/session-1/webhooks/hook-1" {
			t.Error("wrong update endpoint")
		}
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		for key, want := range map[string]string{"active": "false", "secret": `""`, "retryCount": "0", "headers": "{}", "filters": "null"} {
			if string(body[key]) != want {
				t.Errorf("%s should contain %s", key, want)
			}
		}
		if len(body) != 5 {
			t.Errorf("expected exactly five patch fields, got %d", len(body))
		}
		fmt.Fprint(w, webhookJSON)
	})
	hook, err := c.UpdateWebhook(context.Background(), "session-1", "hook-1", UpdateWebhookInput{
		Active: &active, Secret: &secret, RetryCount: &retry, Headers: &headers, Filters: &ConfigValue[WebhookFilters]{},
	})
	if err != nil || hook.ID != "hook-1" || hook.Active {
		t.Fatalf("hook=%v, err=%v", hook, err)
	}
}

func TestUpdateWebhookPreservesAPIError(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "private upstream error", http.StatusNotFound)
	})
	_, err := c.UpdateWebhook(context.Background(), "session-1", "hook-1", UpdateWebhookInput{})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound || errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("err=%v", err)
	}
}
