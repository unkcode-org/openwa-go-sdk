package openwa

type webhookResponse struct {
	Webhook
	Active     *bool `json:"active"`
	RetryCount *int  `json:"retryCount"`
}

func (r webhookResponse) webhook(sessionID, expectedID string) (*Webhook, error) {
	if !validSegment(r.ID) || (expectedID != "" && r.ID != expectedID) || r.SessionID != sessionID ||
		!validHTTPURL(r.URL) || r.Events == nil || r.Active == nil || r.RetryCount == nil ||
		r.CreatedAt.IsZero() || r.UpdatedAt.IsZero() {
		return nil, ErrInvalidResponse
	}

	r.Webhook.Active = *r.Active
	r.Webhook.RetryCount = *r.RetryCount
	return &r.Webhook, nil
}
