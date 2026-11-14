package openwa

import "time"

type SessionStatus string

const (
	SessionStatusCreated        SessionStatus = "created"
	SessionStatusInitializing   SessionStatus = "initializing"
	SessionStatusQRReady        SessionStatus = "qr_ready"
	SessionStatusAuthenticating SessionStatus = "authenticating"
	SessionStatusReady          SessionStatus = "ready"
	SessionStatusDisconnected   SessionStatus = "disconnected"
	SessionStatusActionRequired SessionStatus = "action_required"
	SessionStatusFailed         SessionStatus = "failed"
)

// Session is an OpenWA session. Status alone does not establish whether the
// engine is loaded: a disconnected session may be reconnecting automatically.
// Unknown, nonempty statuses are retained for forward compatibility.
type Session struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Status       SessionStatus       `json:"status"`
	Phone        *string             `json:"phone"`
	PushName     *string             `json:"pushName"`
	ConnectedAt  *time.Time          `json:"connectedAt"`
	LastActive   *time.Time          `json:"lastActive"`
	CreatedAt    time.Time           `json:"createdAt"`
	UpdatedAt    time.Time           `json:"updatedAt"`
	LastError    *string             `json:"lastError"`
	Restriction  *AccountRestriction `json:"restriction"`
	EngineLoaded bool                `json:"engineLoaded"`
}

// RestrictionPasskeyRequired signals that WhatsApp requires a passkey to finish linking.
// Unknown restriction kinds are retained as strings for forward compatibility.
const RestrictionPasskeyRequired = "passkey_required"

type AccountRestriction struct {
	Kind      string     `json:"kind"`
	Code      string     `json:"code"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// QRCode contains an image data URL. Treat it as a credential: do not log it or
// expose it through an unauthenticated endpoint.
type QRCode struct {
	QRCode string        `json:"qrCode"`
	Status SessionStatus `json:"status"`
}

type NumberCheckResult struct {
	Number     string  `json:"number"`
	Exists     bool    `json:"exists"`
	WhatsAppID *string `json:"whatsappId"`
}
