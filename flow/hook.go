package flow

import (
	"context"
	"fmt"
)

// preTransitionHookTable is a map of pre-transition hooks.
type preTransitionHookTable map[State]preTransitionHook

// preTransitionHook is the function that is called before the flow changes its state.
// It is used to perform some actions before the state changes.
// If the error is not nil, the flow will remain in the same state.
type preTransitionHook func(ctx context.Context, data FlowData) error

// RegisterHook registers a pre-transition hook for a state.
// The hook will be called before the flow transitions to the state.
// If the hook returns an error, the transition will be aborted.
// If the hook returns nil, the transition will continue.
// If the state does not exist, the hook will not be registered.
func (f *Flow) RegisterHook(state State, hook preTransitionHook) {
	if f.hookTable == nil {
		f.hookTable = make(map[State]preTransitionHook)
	}
	if _, ok := f.hookTable[state]; !ok {
		f.hookTable[state] = hook
	}
}

func (f *Flow) getPretransitionHook(state State) preTransitionHook {
	if f.hookTable == nil {
		return nil
	}
	return f.hookTable[state]
}

// TypedHook creates a pre-transition hook for a specific data type.
func TypedHook[D FlowData](hook func(D) error) preTransitionHook {
	return func(ctx context.Context, data FlowData) error {
		var castedData D
		var ok bool
		if castedData, ok = data.(D); !ok {
			return fmt.Errorf("invalid data type: %T", data)
		}
		return hook(castedData)
	}
}
