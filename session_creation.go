package openwa

type CreateSessionInput struct {
	Name   string         `json:"name"`
	Config *SessionConfig `json:"config,omitempty"`
}

type SessionConfig struct {
	AutoRejectCalls      *bool `json:"autoRejectCalls,omitempty"`
	MaxReconnectAttempts *int  `json:"maxReconnectAttempts,omitempty"`
	ReconnectBaseDelay   *int  `json:"reconnectBaseDelay,omitempty"`
}

type ListSessionsOptions struct {
	Limit  int
	Offset int
}
