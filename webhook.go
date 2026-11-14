package openwa

import "time"

// Webhook describes a subscription. The gateway treats signing secrets and
// custom headers as write-only.
type Webhook struct {
	ID              string          `json:"id"`
	SessionID       string          `json:"sessionId"`
	URL             string          `json:"url"`
	Events          []string        `json:"events"`
	Filters         *WebhookFilters `json:"filters,omitempty"`
	Active          bool            `json:"active"`
	RetryCount      int             `json:"retryCount"`
	LastTriggeredAt *time.Time      `json:"lastTriggeredAt"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// CreateWebhookInput registers a subscription for a session. Use a signing
// secret when the receiving application verifies webhook deliveries.
type CreateWebhookInput struct {
	URL        string            `json:"url"`
	Events     []string          `json:"events,omitempty"`
	Secret     string            `json:"secret,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Filters    *WebhookFilters   `json:"filters,omitempty"`
	RetryCount *int              `json:"retryCount,omitempty"`
}

// UpdateWebhookInput changes supplied fields. A nonnil empty Secret clears the
// secret, an empty Headers map clears custom headers, and a nonnil ConfigValue
// with nil Value clears filters. Nil fields preserve the current settings.
type UpdateWebhookInput struct {
	URL        *string                      `json:"url,omitempty"`
	Events     []string                     `json:"events,omitempty"`
	Secret     *string                      `json:"secret,omitempty"`
	Headers    *map[string]string           `json:"headers,omitempty"`
	Filters    *ConfigValue[WebhookFilters] `json:"filters,omitempty"`
	Active     *bool                        `json:"active,omitempty"`
	RetryCount *int                         `json:"retryCount,omitempty"`
}

// WebhookFilters applies all conditions to a delivery.
type WebhookFilters struct {
	Conditions []WebhookFilterCondition `json:"conditions"`
}

// WebhookFilterCondition accepts a string, []string, or bool as its Value.
// The gateway validates which fields, operators and values work together.
type WebhookFilterCondition struct {
	Field         string `json:"field"`
	Operator      string `json:"operator"`
	Value         any    `json:"value"`
	CaseSensitive bool   `json:"caseSensitive,omitempty"`
}
