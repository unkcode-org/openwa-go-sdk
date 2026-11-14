package openwa

type CustomLinkPreview struct {
	URL             string `json:"url"`
	Title           string `json:"title"`
	Description     string `json:"description,omitempty"`
	ThumbnailURL    string `json:"thumbnailUrl,omitempty"`
	ThumbnailBase64 string `json:"thumbnailBase64,omitempty"`
}

// TextMessage uses the provider's canonical chat ID, not a regional phone number.
// Link preview support and defaults depend on the gateway engine.
type TextMessage struct {
	ChatID            string             `json:"chatId"`
	Text              string             `json:"text"`
	Mentions          []string           `json:"mentions,omitempty"`
	LinkPreview       *bool              `json:"linkPreview,omitempty"`
	CustomLinkPreview *CustomLinkPreview `json:"customLinkPreview,omitempty"`
	QuotedMessageID   string             `json:"quotedMessageId,omitempty"`
}

// DocumentMessage requires exactly one of Base64 (standard encoded bytes) or URL.
// MIMEType is required for Base64. URL is fetched by the gateway, not the SDK;
// applications should prefer Base64 for private documents and restrict any URLs
// they accept from users to prevent upstream SSRF.
type DocumentMessage struct {
	ChatID          string   `json:"chatId"`
	URL             string   `json:"url,omitempty"`
	Base64          string   `json:"base64,omitempty"`
	MIMEType        string   `json:"mimetype,omitempty"`
	Filename        string   `json:"filename,omitempty"`
	Caption         string   `json:"caption,omitempty"`
	Mentions        []string `json:"mentions,omitempty"`
	QuotedMessageID string   `json:"quotedMessageId,omitempty"`
}

// ImageMessage requires exactly one of Base64 (standard encoded bytes) or URL.
// MIMEType is required for Base64. The gateway fetches URLs, so applications
// should restrict URLs they accept from users.
type ImageMessage struct {
	ChatID          string   `json:"chatId"`
	URL             string   `json:"url,omitempty"`
	Base64          string   `json:"base64,omitempty"`
	MIMEType        string   `json:"mimetype,omitempty"`
	Caption         string   `json:"caption,omitempty"`
	Mentions        []string `json:"mentions,omitempty"`
	QuotedMessageID string   `json:"quotedMessageId,omitempty"`
}

// MessageResult confirms gateway acceptance only, not delivery or read status.
type MessageResult struct {
	MessageID string  `json:"messageId"`
	Timestamp float64 `json:"timestamp"`
}
