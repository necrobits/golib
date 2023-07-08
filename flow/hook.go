package flow

import (
	"context"
	"fmt"
)

// preTransitionHookTable is a map of pre-transition hooks.
type preTransitionHookTable map[State][]hookFn
type silentHookTable map[State][]silentHookFn

func composePreTransitionHooks(hooks []hookFn) hookFn {
	return func(ctx context.Context, data FlowData) error {
		for _, hook := range hooks {
			if err := hook(ctx, data); err != nil {
				return err
			}
		}
		return nil
	}
}

func composeSilentHooks(hooks []silentHookFn) silentHookFn {
	return func(ctx context.Context, data FlowData) {
		for _, hook := range hooks {
			hook(ctx, data)
		}
	}
}

func composeHydrationHooks(hooks []hydrationHookFn) hydrationHookFn {
	return func(ctx context.Context, data FlowData) (FlowData, error) {
		for _, hook := range hooks {
			var err error
			if data, err = hook(ctx, data); err != nil {
				return data, err
			}
		}
		return data, nil
	}
}

type hydrationHookFn func(ctx context.Context, data FlowData) (FlowData, error)

// hookFn is the function that is called before the flow changes its state.
// It is used to perform some actions before the state changes.
// If the error is not nil, the flow will remain in the same state.
type hookFn func(ctx context.Context, data FlowData) error
type silentHookFn func(ctx context.Context, data FlowData)

// RegisterHook registers a pre-transition hook for a state.
// The hook will be called before the flow transitions to the state.
// If the hook returns an error, the transition will be aborted.
// If the hook returns nil, the transition will continue.
// If the state does not exist, the hook will not be registered.
func (f *Flow) RegisterPreTransition(state State, hook hookFn) {
	if f.hookTable == nil {
		f.hookTable = make(preTransitionHookTable)
	}
	if _, ok := f.hookTable[state]; !ok {
		f.hookTable[state] = []hookFn{}
	}
	f.hookTable[state] = append(f.hookTable[state], hook)
}

func (f *Flow) RegisterPostTransition(state State, hook silentHookFn) {
	if f.postHookTable == nil {
		f.postHookTable = make(silentHookTable)
	}
	if _, ok := f.postHookTable[state]; !ok {
		f.postHookTable[state] = []silentHookFn{}
	}
	f.postHookTable[state] = append(f.postHookTable[state], hook)
}

func (f *Flow) RegisterCompletionHook(state State, hook silentHookFn) {
	if f.completionHooks == nil {
		f.completionHooks = make([]silentHookFn, 0)
	}
	f.completionHooks = append(f.completionHooks, hook)
}

func (f *Flow) composePreTransitionHooks(state State) hookFn {
	if f.hookTable == nil {
		return nil
	}
	hooks := f.hookTable[state]
	if len(hooks) == 0 {
		return nil
	}
	return composePreTransitionHooks(hooks)
}

func (f *Flow) composePostTransitionHooks(state State) silentHookFn {
	if f.postHookTable == nil {
		return nil
	}
	hooks := f.postHookTable[state]
	if len(hooks) == 0 {
		return nil
	}
	return composeSilentHooks(hooks)
}

func (f *Flow) composeCompletionHooks() silentHookFn {
	if f.completionHooks == nil {
		return nil
	}
	return composeSilentHooks(f.completionHooks)
}

// TypedPreHook creates a pre-transition hook for a specific data type.
func TypedPreHook[D FlowData](hook func(D) error) hookFn {
	return func(ctx context.Context, data FlowData) error {
		var castedData D
		var ok bool
		if castedData, ok = data.(D); !ok {
			return fmt.Errorf("invalid data type: %T", data)
		}
		return hook(castedData)
	}
}

func TypedHook[D FlowData](hook func(D)) silentHookFn {
	return func(ctx context.Context, data FlowData) {
		var castedData D
		var ok bool
		if castedData, ok = data.(D); !ok {
			panic(fmt.Errorf("invalid data type: %T", data))
		}
		hook(castedData)
	}
}
