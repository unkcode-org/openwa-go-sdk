package openwa

import (
	"context"
	"errors"
	"sync"
	"time"
)

// SessionClient is the narrow lifecycle contract needed by the supervisor.
type SessionClient interface {
	GetSession(context.Context, string) (*Session, error)
	StartSession(context.Context, string) (*Session, error)
}
type ConnectionState string

const (
	ConnectionReady          ConnectionState = "ready"
	ConnectionRecovering     ConnectionState = "recovering"
	ConnectionNeedsAttention ConnectionState = "needs_attention"
	ConnectionMissing        ConnectionState = "missing"
	ConnectionUnavailable    ConnectionState = "unavailable"
)

var ErrAlreadySupervised = errors.New("openwa: session already supervised")

// SupervisionResult omits QR, phone, upstream errors and account details.
// A ready result is an observation of the gateway, not proof of delivery.
type SupervisionResult struct {
	SessionID string
	State     ConnectionState
	Started   bool
}
type SupervisorConfig struct {
	PollInterval   time.Duration
	RequestTimeout time.Duration
	MaxBackoff     time.Duration
	MaxConcurrent  int
	// OnChange runs synchronously from Run, outside lifecycle locks, and must
	// return promptly. It receives no provider bodies, QR or credentials.
	OnChange func(SupervisionResult)
}

// Supervisor never sends messages, destroys credentials or persists state.
// Applications own session scope, durable desired state and distributed leases.
// Stop a Run and wait for its return before manually stopping/unlinking a session.
type Supervisor struct {
	client SessionClient
	config SupervisorConfig
	slots  chan struct{}
	mu     sync.Mutex
	active map[string]bool
	locks  map[string]*sessionGate
}
type sessionGate struct {
	ch   chan struct{}
	refs int
}

func NewSupervisor(client SessionClient, cfg SupervisorConfig) (*Supervisor, error) {
	if client == nil || cfg.PollInterval < 0 || cfg.RequestTimeout < 0 || cfg.MaxBackoff < 0 || cfg.MaxConcurrent < 0 {
		return nil, ErrInvalidConfig
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 30 * time.Second
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 15 * time.Second
	}
	if cfg.MaxBackoff == 0 {
		cfg.MaxBackoff = 5 * time.Minute
	}
	if cfg.MaxBackoff < cfg.PollInterval {
		return nil, ErrInvalidConfig
	}
	if cfg.MaxConcurrent == 0 {
		cfg.MaxConcurrent = 4
	}
	return &Supervisor{client: client, config: cfg, slots: make(chan struct{}, cfg.MaxConcurrent), active: make(map[string]bool), locks: make(map[string]*sessionGate)}, nil
}
func (s *Supervisor) acquire(ctx context.Context, id string) (func(), error) {
	s.mu.Lock()
	g := s.locks[id]
	if g == nil {
		g = &sessionGate{ch: make(chan struct{}, 1)}
		s.locks[id] = g
	}
	g.refs++
	s.mu.Unlock()
	drop := func() {
		s.mu.Lock()
		g.refs--
		if g.refs == 0 {
			delete(s.locks, id)
		}
		s.mu.Unlock()
	}
	select {
	case g.ch <- struct{}{}:
	case <-ctx.Done():
		drop()
		return nil, ctx.Err()
	}
	select {
	case s.slots <- struct{}{}:
	case <-ctx.Done():
		<-g.ch
		drop()
		return nil, ctx.Err()
	}
	return func() { <-s.slots; <-g.ch; drop() }, nil
}
func connectionState(session *Session) ConnectionState {
	if session == nil {
		return ConnectionUnavailable
	}
	if session.Restriction != nil || session.Status == SessionStatusQRReady || session.Status == SessionStatusActionRequired {
		return ConnectionNeedsAttention
	}
	if session.Status == SessionStatusReady && session.EngineLoaded {
		return ConnectionReady
	}
	switch session.Status {
	case SessionStatusCreated, SessionStatusInitializing, SessionStatusAuthenticating, SessionStatusReady, SessionStatusDisconnected, SessionStatusFailed:
		return ConnectionRecovering
	default:
		return ConnectionNeedsAttention
	}
}
