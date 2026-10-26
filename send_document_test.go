package openwa

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestSendDocumentSourcesAndPayload(t *testing.T) {
	for _, fromURL := range []bool{false, true} {
		input := DocumentMessage{ChatID: "549351@c.us", Filename: "factura.pdf", Caption: "Tu factura", MIMEType: "application/pdf"}
		if fromURL {
			input.URL = "https://example.com/factura.pdf"
		} else {
			input.Base64 = base64.StdEncoding.EncodeToString([]byte("%PDF-test"))
		}
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" || r.URL.Path != "/api/sessions/session-1/messages/send-document" {
				t.Error("wrong request")
			}
			var body DocumentMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body.ChatID != input.ChatID || body.Filename != input.Filename || body.Caption != input.Caption || body.Base64 != input.Base64 || body.URL != input.URL || body.MIMEType != input.MIMEType {
				t.Error("document modified")
			}
			w.WriteHeader(201)
			fmt.Fprint(w, messageJSON)
		})
		result, err := c.SendDocument(context.Background(), "session-1", input)
		if err != nil || result.MessageID != "message-1" {
			t.Fatalf("result=%v err=%v", result, err)
		}
	}
}

func TestDocumentValidationAndByteLimits(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid document dispatched") }, func(cfg *Config) { cfg.MaxDocumentBytes = 3 })
	for _, input := range []DocumentMessage{
		{ChatID: "x@c.us"},
		{ChatID: "x@c.us", URL: "https://example.com/x", Base64: "YQ==", MIMEType: "application/pdf"},
		{ChatID: "x@c.us", Base64: "YQ=="},
		{ChatID: "x@c.us", Base64: "!!!!", MIMEType: "application/pdf"},
		{ChatID: "x@c.us", Base64: "YQ==\n", MIMEType: "application/pdf"},
		{ChatID: "x@c.us", URL: "file:///private/document"},
		{ChatID: "x@c.us", URL: "https://user:pass@example.com/x"},
		{ChatID: "x@c.us", URL: "https://example.com/x", Caption: strings.Repeat("x", 1025)},
	} {
		_, err := c.SendDocument(context.Background(), "session-1", input)
		if !errors.Is(err, ErrInvalidInput) || errors.Is(err, ErrOutcomeUnknown) {
			t.Fatalf("err=%v", err)
		}
	}
	_, err := c.SendDocument(context.Background(), "session-1", DocumentMessage{ChatID: "x@c.us", Base64: base64.StdEncoding.EncodeToString([]byte("four")), MIMEType: "application/pdf"})
	if !errors.Is(err, ErrDocumentTooLarge) || errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("err=%v", err)
	}
}

func TestDocumentLimitAcceptsExactDecodedLength(t *testing.T) {
	for size := 1; size <= 4; size++ {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(201); fmt.Fprint(w, messageJSON) }, func(cfg *Config) { cfg.MaxDocumentBytes = int64(size) })
		_, err := c.SendDocument(context.Background(), "session-1", DocumentMessage{ChatID: "x@c.us", Base64: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("x", size))), MIMEType: "application/pdf"})
		if err != nil {
			t.Fatalf("size=%d err=%v", size, err)
		}
	}
}
