package openwa

import "strings"

// The explicit pointer distinguishes an absent engineLoaded from false.
type sessionResponse struct {
	Session
	EngineLoaded *bool `json:"engineLoaded"`
}

func (r sessionResponse) session(expectedID string) (*Session, error) {
	if r.ID != expectedID || strings.TrimSpace(r.Name) == "" || strings.TrimSpace(string(r.Status)) == "" ||
		r.CreatedAt.IsZero() || r.UpdatedAt.IsZero() || r.EngineLoaded == nil {
		return nil, ErrInvalidResponse
	}
	r.Session.EngineLoaded = *r.EngineLoaded
	return &r.Session, nil
}
