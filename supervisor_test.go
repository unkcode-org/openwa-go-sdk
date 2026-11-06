package openwa

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type supervisionFake struct {
	get   func(context.Context, string) (*Session, error)
	start func(context.Context, string) (*Session, error)
}

func (f supervisionFake) GetSession(c context.Context, id string) (*Session, error) {
	return f.get(c, id)
}
func (f supervisionFake) StartSession(c context.Context, id string) (*Session, error) {
	return f.start(c, id)
}
func TestReconcileStateAndEngineOwnership(t *testing.T) {
	for _, tc := range []struct {
		status      SessionStatus
		loaded      bool
		restriction bool
		start       bool
		state       ConnectionState
	}{
		{SessionStatusReady, true, false, false, ConnectionReady},
		{SessionStatusDisconnected, true, false, false, ConnectionRecovering},
		{SessionStatusInitializing, true, false, false, ConnectionRecovering},
		{SessionStatusDisconnected, false, false, true, ConnectionRecovering},
		{SessionStatusFailed, false, false, true, ConnectionRecovering},
		{SessionStatusCreated, false, false, true, ConnectionRecovering},
		{SessionStatusQRReady, true, false, false, ConnectionNeedsAttention},
		{SessionStatusActionRequired, false, false, false, ConnectionNeedsAttention},
		{SessionStatusAuthenticating, false, false, false, ConnectionRecovering},
		{SessionStatusReady, true, true, false, ConnectionNeedsAttention},
		{SessionStatus("future"), false, false, false, ConnectionNeedsAttention},
	} {
		t.Run(string(tc.status)+time.Duration(boolInt(tc.loaded)).String(), func(t *testing.T) {
			starts := 0
			f := supervisionFake{get: func(context.Context, string) (*Session, error) {
				s := &Session{ID: "s", Status: tc.status, EngineLoaded: tc.loaded}
				if tc.restriction {
					s.Restriction = &AccountRestriction{Kind: "blocked"}
				}
				return s, nil
			}, start: func(context.Context, string) (*Session, error) {
				starts++
				return &Session{ID: "s", Status: SessionStatusInitializing, EngineLoaded: true}, nil
			}}
			supervisor, _ := NewSupervisor(f, SupervisorConfig{})
			result, e := supervisor.Reconcile(context.Background(), "s")
			if e != nil || result.State != tc.state || starts != boolInt(tc.start) || result.Started != tc.start {
				t.Fatalf("result=%+v starts=%d err=%v", result, starts, e)
			}
		})
	}
}
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
func TestSupervisorMissingConflictAndNoRecreation(t *testing.T) {
	calls := 0
	starts := 0
	f := supervisionFake{get: func(context.Context, string) (*Session, error) {
		calls++
		if calls == 1 {
			return &Session{ID: "s", Status: SessionStatusDisconnected}, nil
		}
		return &Session{ID: "s", Status: SessionStatusReady, EngineLoaded: true}, nil
	}, start: func(context.Context, string) (*Session, error) { starts++; return nil, &APIError{StatusCode: 409} }}
	s, _ := NewSupervisor(f, SupervisorConfig{})
	r, e := s.Reconcile(context.Background(), "s")
	if e != nil || r.State != ConnectionReady || starts != 1 || calls != 2 {
		t.Fatalf("r=%v e=%v", r, e)
	}
	f.get = func(context.Context, string) (*Session, error) { return nil, &APIError{StatusCode: 404} }
	s, _ = NewSupervisor(f, SupervisorConfig{})
	r, e = s.Reconcile(context.Background(), "s")
	if e == nil || r.State != ConnectionMissing || starts != 1 {
		t.Fatal("missing session recreated")
	}
}
func TestRunCancellationAndDuplicateOwnership(t *testing.T) {
	entered := make(chan struct{})
	var once sync.Once
	f := supervisionFake{get: func(ctx context.Context, id string) (*Session, error) {
		once.Do(func() { close(entered) })
		<-ctx.Done()
		return nil, ctx.Err()
	}}
	s, _ := NewSupervisor(f, SupervisorConfig{})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx, "s") }()
	<-entered
	if e := s.Run(ctx, "s"); !errors.Is(e, ErrAlreadySupervised) {
		t.Fatal(e)
	}
	cancel()
	select {
	case e := <-done:
		if !errors.Is(e, context.Canceled) {
			t.Fatal(e)
		}
	case <-time.After(time.Second):
		t.Fatal("Run leaked")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.locks) != 0 || len(s.active) != 0 {
		t.Fatal("ownership leaked")
	}
}
func TestSupervisorSerializesSameSessionAndBoundsConcurrency(t *testing.T) {
	var current, max atomic.Int32
	f := supervisionFake{get: func(ctx context.Context, id string) (*Session, error) {
		n := current.Add(1)
		defer current.Add(-1)
		for old := max.Load(); n > old && !max.CompareAndSwap(old, n); old = max.Load() {
		}
		time.Sleep(time.Millisecond)
		return &Session{ID: id, Status: SessionStatusReady, EngineLoaded: true}, nil
	}}
	s, _ := NewSupervisor(f, SupervisorConfig{MaxConcurrent: 2})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = s.Reconcile(context.Background(), "s") }()
	}
	wg.Wait()
	if max.Load() != 1 {
		t.Fatal("same session raced")
	}
}
func TestRunObservesTransitionsAndRecoversWithoutMessages(t *testing.T) {
	var gets atomic.Int32
	f := supervisionFake{get: func(context.Context, string) (*Session, error) {
		if gets.Add(1) == 1 {
			return &Session{ID: "s", Status: SessionStatusDisconnected}, nil
		}
		return &Session{ID: "s", Status: SessionStatusReady, EngineLoaded: true}, nil
	}, start: func(context.Context, string) (*Session, error) {
		return &Session{ID: "s", Status: SessionStatusInitializing, EngineLoaded: true}, nil
	}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	states := []ConnectionState{}
	s, _ := NewSupervisor(f, SupervisorConfig{PollInterval: time.Millisecond, MaxBackoff: 10 * time.Millisecond, OnChange: func(r SupervisionResult) {
		states = append(states, r.State)
		if r.State == ConnectionReady {
			cancel()
		}
	}})
	if e := s.Run(ctx, "s"); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if len(states) != 2 || states[0] != ConnectionRecovering || states[1] != ConnectionReady {
		t.Fatal(states)
	}
}

func TestSupervisorRequestTimeoutAndInvalidInput(t *testing.T) {
	f := supervisionFake{get: func(ctx context.Context, id string) (*Session, error) { <-ctx.Done(); return nil, ctx.Err() }}
	s, e := NewSupervisor(f, SupervisorConfig{RequestTimeout: time.Millisecond})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Reconcile(context.Background(), "s"); !errors.Is(e, context.DeadlineExceeded) {
		t.Fatal(e)
	}
	if _, e = s.Reconcile(nil, "s"); !errors.Is(e, ErrInvalidInput) {
		t.Fatal(e)
	}
	for _, cfg := range []SupervisorConfig{{PollInterval: -1}, {MaxConcurrent: -1}, {RequestTimeout: -1}, {PollInterval: time.Hour, MaxBackoff: time.Second}} {
		if _, e = NewSupervisor(f, cfg); !errors.Is(e, ErrInvalidConfig) {
			t.Fatal(e)
		}
	}
}
