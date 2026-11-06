package openwa

import "encoding/json"

// ConfigValue distinguishes an explicit null (Value=nil) from an omitted patch
// field (the *ConfigValue itself is nil). This also preserves false and zero.
type ConfigValue[T any] struct{ Value *T }

func (v ConfigValue[T]) MarshalJSON() ([]byte, error) { return json.Marshal(v.Value) }

// SessionConfigPatch merges supplied fields. A nonnil ConfigValue with nil Value
// restores the gateway default. Reconnect changes apply on the next start.
type SessionConfigPatch struct {
	AutoRejectCalls      *ConfigValue[bool] `json:"autoRejectCalls,omitempty"`
	MaxReconnectAttempts *ConfigValue[int]  `json:"maxReconnectAttempts,omitempty"`
	ReconnectBaseDelay   *ConfigValue[int]  `json:"reconnectBaseDelay,omitempty"`
}

// EffectiveSessionConfig is the gateway's current tunable configuration.
// Nil MaxReconnectAttempts means unlimited; zero disables reconnection.
type EffectiveSessionConfig struct {
	AutoRejectCalls      bool `json:"autoRejectCalls"`
	MaxReconnectAttempts *int `json:"maxReconnectAttempts"`
	ReconnectBaseDelay   int  `json:"reconnectBaseDelay"`
}

func decodeSessionConfig(raw json.RawMessage) (*EffectiveSessionConfig, error) {
	var wire struct {
		AutoRejectCalls      *bool           `json:"autoRejectCalls"`
		MaxReconnectAttempts json.RawMessage `json:"maxReconnectAttempts"`
		ReconnectBaseDelay   *int            `json:"reconnectBaseDelay"`
	}
	if json.Unmarshal(raw, &wire) != nil || wire.AutoRejectCalls == nil || len(wire.MaxReconnectAttempts) == 0 || wire.ReconnectBaseDelay == nil {
		return nil, ErrInvalidResponse
	}
	var cap *int
	if json.Unmarshal(wire.MaxReconnectAttempts, &cap) != nil || cap != nil && (*cap < 0 || *cap > 20) || *wire.ReconnectBaseDelay < 1000 || *wire.ReconnectBaseDelay > 300000 {
		return nil, ErrInvalidResponse
	}
	return &EffectiveSessionConfig{AutoRejectCalls: *wire.AutoRejectCalls, MaxReconnectAttempts: cap, ReconnectBaseDelay: *wire.ReconnectBaseDelay}, nil
}
