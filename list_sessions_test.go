package openwa

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestListSessionsPaginationAndEmptyPage(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/api/sessions" {
			t.Error("wrong list request")
		}
		if r.URL.Query().Get("limit") != "100" {
			t.Error("missing limit")
		}
		if r.URL.Query().Get("offset") == "100" {
			fmt.Fprint(w, `[]`)
			return
		}
		fmt.Fprintf(w, "[%s]", sessionJSON)
	})
	page, err := c.ListSessions(context.Background(), &ListSessionsOptions{Limit: 100})
	if err != nil || len(page) != 1 || page[0].ID != "session-1" {
		t.Fatalf("page=%v err=%v", page, err)
	}
	page, err = c.ListSessions(context.Background(), &ListSessionsOptions{Limit: 100, Offset: 100})
	if err != nil || page == nil || len(page) != 0 {
		t.Fatalf("page=%v err=%v", page, err)
	}
}

func TestListSessionsValidatesPaginationAndResponse(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) { t.Error("invalid options dispatched") })
	for _, option := range []*ListSessionsOptions{{Limit: -1}, {Limit: 1001}, {Offset: -1}} {
		if _, err := c.ListSessions(context.Background(), option); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("err=%v", err)
		}
	}
	for _, body := range []string{`{}`, `null`, `[{}]`} {
		c = testClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) })
		if _, err := c.ListSessions(context.Background(), nil); !errors.Is(err, ErrInvalidResponse) {
			t.Fatalf("body=%s err=%v", body, err)
		}
	}
}
