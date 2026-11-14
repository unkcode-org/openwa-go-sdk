package openwa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestSendImage(t *testing.T) {
	for _, input := range []ImageMessage{
		{ChatID: "549351@c.us", Base64: "YQ==", MIMEType: "image/png", Caption: "Foto"},
		{ChatID: "549351@c.us", URL: "https://example.com/image.jpg", QuotedMessageID: "quoted-id"},
	} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || r.URL.Path != "/api/sessions/session-1/messages/send-image" {
				t.Error("wrong image endpoint")
			}
			var body ImageMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || !reflect.DeepEqual(body, input) {
				t.Errorf("wrong image payload: %v", err)
			}
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, messageJSON)
		})
		result, err := c.SendImage(context.Background(), "session-1", input)
		if err != nil || result.MessageID != "message-1" {
			t.Fatalf("result=%v err=%v", result, err)
		}
	}
}

func TestSendImageValidationAndUncertainOutcome(t *testing.T) {
	input := ImageMessage{ChatID: "549351@c.us", Base64: "YQ==", MIMEType: "image/png"}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("invalid image dispatched")
	}, func(cfg *Config) { cfg.MaxRequestBytes = 32 })
	if _, err := c.SendImage(context.Background(), "session-1", input); !errors.Is(err, ErrRequestTooLarge) || errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("request limit: %v", err)
	}
	invalid := input
	invalid.URL = "https://example.com/image.jpg"
	if _, err := c.SendImage(context.Background(), "session-1", invalid); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("multiple sources: %v", err)
	}

	c = testClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream body", http.StatusServiceUnavailable)
	})
	_, err := c.SendImage(context.Background(), "session-1", input)
	var apiErr *APIError
	if !errors.Is(err, ErrOutcomeUnknown) || !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("uncertain outcome: %v", err)
	}
}
