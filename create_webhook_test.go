package openwa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestCreateWebhook(t *testing.T) {
	const secret = "fixture-signing-secret"
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/sessions/session-1/webhooks" {
			t.Error("wrong create endpoint")
		}
		var body CreateWebhookInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.URL != "https://example.com/receipts" ||
			len(body.Events) != 1 || body.Events[0] != "message.ack" || body.Secret != secret {
			t.Error("wrong subscription payload")
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, webhookJSON)
	})
	hook, err := c.CreateWebhook(context.Background(), "session-1", CreateWebhookInput{
		URL: "https://example.com/receipts", Events: []string{"message.ack"}, Secret: secret,
	})
	if err != nil || hook.ID != "hook-1" {
		t.Fatalf("hook=%v, err=%v", hook, err)
	}
}

func TestCreateWebhookRejectsInvalidInput(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid webhook dispatched") })
	for _, input := range []CreateWebhookInput{
		{URL: "file:///secret"},
		{URL: "https://example.com/receipts", Secret: "short"},
		{URL: "https://example.com/receipts", Events: []string{}},
	} {
		if _, err := c.CreateWebhook(context.Background(), "session-1", input); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("err=%v", err)
		}
	}
}
