package flow

import (
	"context"
	"time"
)

type StateMachine interface {
	HandleAction(ctx context.Context, a Action) error
	TransitionTable() TransitionTable

	ID() string
	Type() FlowType
	CurrentState() State
	Data() FlowData
	IsCompleted() bool
	IsExpired() bool
	ExpiresAt() time.Time
	SetExpirationAt(t time.Time)
	SetExpirationIn(t time.Duration)
	ToSnapshot() (*Snapshot, error)

	RegisterPreTransition(state State, hook hookFn)
	RegisterPostTransition(state State, hook silentHookFn)
	RegisterCompletionHook(state State, hook silentHookFn)
}
