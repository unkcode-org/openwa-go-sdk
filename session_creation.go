package openwa

type CreateSessionInput struct {
	Name      string         `json:"name"`
	Config    *SessionConfig `json:"config,omitempty"`
	ProxyURL  string         `json:"proxyUrl,omitempty"`
	ProxyType ProxyType      `json:"proxyType,omitempty"`
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

// ProxyType is the gateway's per-session egress proxy protocol.
type ProxyType string

const (
	ProxyTypeHTTP   ProxyType = "http"
	ProxyTypeHTTPS  ProxyType = "https"
	ProxyTypeSOCKS4 ProxyType = "socks4"
	ProxyTypeSOCKS5 ProxyType = "socks5"
)
