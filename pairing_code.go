package openwa

type PairingCodeInput struct {
	PhoneNumber string `json:"phoneNumber"`
}

// PairingCode is authentication material; never log or expose it publicly.
type PairingCode struct {
	PairingCode string        `json:"pairingCode"`
	Status      SessionStatus `json:"status"`
}
