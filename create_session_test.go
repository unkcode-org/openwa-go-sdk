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

func TestCreateSessionUsesServerIDAndExactName(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/sessions" || r.Header.Get("X-API-Key") != "test-secret" {
			t.Error("wrong creation request")
		}
		var input CreateSessionInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Error(err)
		}
		if input.Name != "test-session" {
			t.Error("name changed")
		}
		w.WriteHeader(201)
		fmt.Fprint(w, sessionJSON)
	})
	created, err := c.CreateSession(context.Background(), CreateSessionInput{Name: "test-session"})
	if err != nil || created.ID != "session-1" {
		t.Fatalf("created=%v err=%v", created, err)
	}
}

func TestCreateSessionValidationAndConflict(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid input dispatched") })
	for _, name := range []string{"", "ab", "my session", "../session", strings.Repeat("x", 51)} {
		if _, err := c.CreateSession(context.Background(), CreateSessionInput{Name: name}); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("name=%q err=%v", name, err)
		}
	}
	c = testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(409)
		fmt.Fprint(w, `{"message":"already exists"}`)
	})
	_, err := c.CreateSession(context.Background(), CreateSessionInput{Name: "test-session"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 409 {
		t.Fatalf("err=%v", err)
	}
}

func TestCreateSessionRejectsDifferentOrMalformedResponse(t *testing.T) {
	for _, body := range []string{sessionJSON, `{}`, `[]`, `null`} {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(201); fmt.Fprint(w, body) })
		if _, err := c.CreateSession(context.Background(), CreateSessionInput{Name: "another-session"}); !errors.Is(err, ErrInvalidResponse) {
			t.Fatalf("err=%v", err)
		}
	}
}
