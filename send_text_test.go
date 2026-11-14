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

func TestSendTextPayloadAndResult(t *testing.T) {
	preview := false
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/sessions/session-1/messages/send-text" || r.Header.Get("Content-Type") != "application/json" {
			t.Error("wrong request")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["chatId"] != "549351@c.us" || body["text"] != " Hola 👋 " || body["linkPreview"] != false || body["quotedMessageId"] != "quoted-1" {
			t.Errorf("payload=%v", body)
		}
		w.WriteHeader(201)
		fmt.Fprint(w, messageJSON)
	})
	result, err := c.SendText(context.Background(), "session-1", TextMessage{ChatID: "549351@c.us", Text: " Hola 👋 ", LinkPreview: &preview, QuotedMessageID: "quoted-1"})
	if err != nil || result.MessageID != "message-1" || result.Timestamp != 1793016000 {
		t.Fatalf("result=%v err=%v", result, err)
	}
}

func TestSendTextCustomPreviewThumbnails(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Preview CustomLinkPreview `json:"customLinkPreview"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Preview.ThumbnailURL != "https://example.com/image.jpg" || body.Preview.ThumbnailBase64 != "aW1hZ2U=" {
			t.Errorf("unexpected preview: %+v", body.Preview)
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, messageJSON)
	})
	_, err := c.SendText(context.Background(), "session-1", TextMessage{
		ChatID: "123456@c.us", Text: "https://example.com",
		CustomLinkPreview: &CustomLinkPreview{URL: "https://example.com", Title: "Example", ThumbnailURL: "https://example.com/image.jpg", ThumbnailBase64: "aW1hZ2U="},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTextValidationBeforeNetwork(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid input dispatched") })
	for _, input := range []TextMessage{{ChatID: "", Text: "hello"}, {ChatID: "x@c.us", Text: " "}, {ChatID: "x@c.us", Text: strings.Repeat("x", 4097)}, {ChatID: "x@c.us", Text: string([]byte{0xff})}, {ChatID: "x@c.us", Text: "hello", CustomLinkPreview: &CustomLinkPreview{URL: "file:///tmp/x", Title: "x"}}} {
		if _, err := c.SendText(context.Background(), "session-1", input); !errors.Is(err, ErrInvalidInput) || errors.Is(err, ErrOutcomeUnknown) {
			t.Fatalf("err=%v", err)
		}
	}
	preview := false
	_, err := c.SendText(context.Background(), "session-1", TextMessage{ChatID: "x@c.us", Text: "hello", LinkPreview: &preview, CustomLinkPreview: &CustomLinkPreview{URL: "https://example.com", Title: "x"}})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err=%v", err)
	}
}

func TestUnconfirmedSendResultsAreUnknown(t *testing.T) {
	for _, body := range []string{`{}`, `null`, `{"messageId":""}`, `{"messageId":"m"}`, `{"messageId":"m","timestamp":null}`, `{"messageId":"m","timestamp":-1}`, `{"messageId":"m","timestamp":"x"}`, `{"messageId":"m","timestamp":1} extra`} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(201); fmt.Fprint(w, body) })
		result, err := c.SendText(context.Background(), "session-1", TextMessage{ChatID: "x@c.us", Text: "hello"})
		if result != nil || !errors.Is(err, ErrOutcomeUnknown) || !errors.Is(err, ErrInvalidResponse) {
			t.Fatalf("body=%s result=%v err=%v", body, result, err)
		}
	}
}

func TestRequestLimitIsDefinitiveLocalFailure(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("oversized request dispatched") }, func(cfg *Config) { cfg.MaxRequestBytes = 10 })
	_, err := c.SendText(context.Background(), "session-1", TextMessage{ChatID: "x@c.us", Text: "hello"})
	if !errors.Is(err, ErrRequestTooLarge) || errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("err=%v", err)
	}
}
