package flow

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

const (
	// NoEvent is a special event, which is used to indicate that no event is sent to the state machine.
	NoEvent Event = ""
)

var (
	isDebugging = false
)

// Flow is the state machine.
// It contains the internal data of the flow, and the current state.
// It also contains the transition table, which describes the state machine.
type Flow struct {
	id              string
	flowType        FlowType
	data            FlowData
	currentState    State
	states          TransitionTable
	defaultHandler  ActionHandler
	expiresAt       time.Time
	completed       bool
	hookTable       preTransitionHookTable
	postHookTable   silentHookTable
	completionHooks []silentHookFn
}

// FlowData is the internal data of the flow. This can be anything.
type FlowData interface{}
type FlowType string

type TransitionTable map[State]StateConfig

// Action is the interface that all actions must implement.
type Action interface {
	Type() ActionType
}

// ActionType is the type of the action. It is used to identify the action in the TransitionTable.
type ActionType string

// AutoPassAction is a special action, which is used to handle automatic transitions.
type autopassAction struct {
}

func (a autopassAction) Type() ActionType {
	return "Autopass"
}

type nilAction struct {
}

func (a nilAction) Type() ActionType {
	return "Nil"
}

// NilAction is a special action, which can be used when you don't have any specific action.
func NilAction() Action {
	return nilAction{}
}

// Event is the input to the state machine. It is the output of the action handler.
// Hence, it is only a string. The action handler is responsible for changing the data of the flow.
// It causes the state machine to change its state.
type Event string

// State is just a name for a state in the state machine.
type State string

// CreateFlowOpts is used to create a new Flow
// Both ID and Type are used to identify the flow and restore it from a snapshot.
type CreateFlowOpts struct {
	// ID is the unique identifier for the flow.
	ID string
	// Type is the type of the flow. This is used to identify the flow and restore it from a snapshot.
	Type FlowType
	// Data is the initial internal data for the flow.
	Data FlowData
	// InitialState is the initial state for the flow. It must be one of the states in TransitionTable.
	InitialState State
	// TransitionTable contains the description of the state machine for the flow.
	TransitionTable TransitionTable
	// Handler is the default handler for the flow. If a state does not have a handler, this handler is used.
	Handler ActionHandler
	// ExpireAt is the time at which the flow expires.
	ExpireAt time.Time
	// ExpireIn is the duration after which the flow expires.
	// If both ExpireAt and ExpireIn are set, ExpireIn is used.
	ExpireIn time.Duration
}

// Snapshot is used to persist the flow, and restore it later.
type Snapshot struct {
	// ID of the original flow.
	ID string `json:"id"`
	// Type of the original flow.
	Type string `json:"type"`
	// Marshalled data of the flow.
	EncodedData json.RawMessage `json:"data"`
	// Data is the decoded data of the flow. Must be marshallable
	Data FlowData `json:"-"`
	// CurrentState of the original flow.
	CurrentState State `json:"current_state"`
	// ExpiresAt is the time at which the flow expires.
	ExpiresAt sql.NullTime `json:"expire_at"`
	// IsCompleted indicates whether the flow is completed or not.
	IsCompleted bool `json:"is_completed"`
}

// ActionHandler is the function that handles an action.
// It returns:
//   - an Event, which is the input to the state machine.
//   - the new internal data of the flow, which replaces the old one.
//
// If the error is not nil, no event will be sent to the state machine, and the flow will remain in the same state.
// If the error is nil, the flow will change its state according to the the transition table.
type ActionHandler func(ctx context.Context, data FlowData, a Action) (Event, FlowData, error)

// Transitions is a map, describing the transition from a state to the next state.
type Transitions map[Event]State

// StateConfig is the configuration for a state in the state machine.
// In addition to just a name, a State contains a handler function, and a map of transitions.
// When the flow is in a state, and an action is received, the handler function of that state is called.
// If you are going to handle one action in a state, you may consider using the [TypedHandler] function to avoid type assertion.
// If you are going to handle multiple actions in a state, you may consider using the [NewRouter] function create an action router.
type StateConfig struct {
	// Handler is the handler function for the state. It receives the current internal data of the flow, and the action.
	Handler ActionHandler
	// Transitions is a map, describing the transition from a state to the next state when an event occurs.
	Transitions Transitions
	// Final indicates whether the state is a final state or not. If the Flow reachs a final state, it is completed.
	Final bool
	// Autopass indicates the state will automatically transition to the next state without any action.
	// The handler function will be called with an [AutopassAction].
	Autopass bool
}

// HandleAction handles an action for the flow.
// Everytime an action is handled, the flow may change its state.
// This function is the only way to change the state of the flow.
func (f *Flow) HandleAction(ctx context.Context, a Action) error {
	if f.completed {
		return fmt.Errorf("flow is completed")
	}
	if f.IsExpired() {
		return fmt.Errorf("flow expired")
	}

	actionType := a.Type()
	stateConfig, ok := f.states[f.currentState]
	if !ok {
		return fmt.Errorf("illegal state: %s", f.currentState)
	}
	actionHandler := stateConfig.Handler

	if actionHandler == nil {
		if f.defaultHandler == nil {
			return fmt.Errorf("no handler for state: %s and no default handler found. Did you forget to set one of them?", f.currentState)
		}
		actionHandler = f.defaultHandler
	}
	f.logf("Incoming action: %s\n", actionType)
	inputEvent, nextData, err := actionHandler(ctx, f.data, a)
	if err != nil {
		f.logf("Error: %v\n", err)
		return err
	}
	var nextState State
	if inputEvent == NoEvent {
		f.logf("<Action>%s -> No event\n", actionType)
		nextState = f.currentState
	} else {
		nextState, ok = stateConfig.Transitions[inputEvent]
		if !ok {
			return fmt.Errorf("no transition found for event: %s", actionType)
		}
		f.logf("<Action>%s -> <Event>%s\n", actionType, inputEvent)

		err := f.runPreTransitionHooks(ctx, nextData, nextState)
		if err != nil {
			return err
		}
		f.logf("Transition: %s -> %s\n", f.currentState, nextState)
	}

	f.data = nextData
	f.currentState = nextState
	f.runPostTransitionHooks(ctx, nextData, nextState)

	if nextStateConfig, ok := f.states[nextState]; ok {
		if nextStateConfig.Final {
			f.completed = true
			f.runCompletionHooks(ctx, nextData)
			f.logf("Flow completed\n")
		}
	}

	if nextStateConfig, ok := f.states[nextState]; ok && nextStateConfig.Autopass {
		f.logf("Reached an autopass state: %s\n", nextState)
		return f.HandleAction(ctx, autopassAction{})
	}
	return nil
}

func (f *Flow) runPreTransitionHooks(ctx context.Context, data FlowData, nextState State) error {
	hook := f.composePreTransitionHooks(nextState)
	if hook != nil {
		f.logf("Calling pre-transition hook for state: %s\n", nextState)
		if err := hook(ctx, data); err != nil {
			f.logf("Error during pre-transition hook: %v\n", err)
			return err
		}
	}
	registryHook := globalHookRegistry.composePreTransitions(f.flowType, nextState)
	if registryHook != nil {
		f.logf("Calling pre-transition hook from global registry for state: %s\n", nextState)
		if err := registryHook(ctx, data); err != nil {
			f.logf("Error during pre-transition hook from registry: %v\n", err)
			return err
		}
	}
	return nil
}

func (f *Flow) runPostTransitionHooks(ctx context.Context, data FlowData, nextState State) {
	hook := f.composePostTransitionHooks(nextState)
	if hook != nil {
		f.logf("Calling post-transition hook for state: %s\n", nextState)
		hook(ctx, data)
	}
	registryHook := globalHookRegistry.composePostTransitionHooks(f.flowType, nextState)
	if registryHook != nil {
		f.logf("Calling post-transition hook from global registry for state: %s\n", nextState)
		registryHook(ctx, data)
	}
}

func (f *Flow) runCompletionHooks(ctx context.Context, data FlowData) {
	hook := f.composeCompletionHooks()
	if hook != nil {
		f.logf("Calling completion hook\n")
		hook(ctx, data)
	}
	registryHook := globalHookRegistry.composeCompletionHooks(f.flowType)
	if registryHook != nil {
		f.logf("Calling completion hook from global registry\n")
		registryHook(ctx, data)
	}
}

// New creates a new Flow.
func New(opts CreateFlowOpts) *Flow {
	if opts.TransitionTable == nil {
		panic("TransitionTable cannot be nil")
	}
	if opts.InitialState == "" {
		panic("InitialState cannot be empty")
	}
	logf("Creating a new Flow(%s), ID=%s\n", opts.Type, opts.ID)
	if opts.ExpireIn > 0 {
		opts.ExpireAt = time.Now().Add(opts.ExpireIn)
	}

	return &Flow{
		id:             opts.ID,
		flowType:       opts.Type,
		data:           opts.Data,
		currentState:   opts.InitialState,
		states:         opts.TransitionTable,
		defaultHandler: opts.Handler,
		expiresAt:      opts.ExpireAt,
	}
}

// ToSnapshot converts the flow to a Snapshot to be persisted.
func (f *Flow) ToSnapshot() (*Snapshot, error) {
	dataJson, err := json.Marshal(f.data)
	if err != nil {
		return nil, err
	}
	return &Snapshot{
		ID:           f.id,
		Type:         string(f.flowType),
		Data:         f.data,
		EncodedData:  dataJson,
		CurrentState: f.currentState,
		ExpiresAt: sql.NullTime{
			Time:  f.expiresAt,
			Valid: !f.expiresAt.IsZero(),
		},
		IsCompleted: f.completed,
	}, nil
}

func FromSnapShot(ctx context.Context, s *Snapshot, stateMap TransitionTable) (*Flow, error) {
	f := UnhydratedFromSnapshot(s, stateMap)
	hydrationFn := HookRegistry().composeHydrationHooks(f.flowType)
	if hydrationFn != nil {
		var err error
		if f.data, err = hydrationFn(ctx, f.data); err != nil {
			return nil, err
		}
	}
	return f, nil
}

// FromSnapshot restores a flow from a Snapshot.
func UnhydratedFromSnapshot(s *Snapshot, stateMap TransitionTable) *Flow {
	flow := Flow{
		id:           s.ID,
		flowType:     FlowType(s.Type),
		data:         s.Data,
		currentState: s.CurrentState,
		states:       stateMap,
		completed:    s.IsCompleted,
	}
	if s.ExpiresAt.Valid {
		flow.expiresAt = s.ExpiresAt.Time
	}
	return &flow
}

// WithDefaultActionHandler sets the default action handler for the flow.
// If a state does not have a handler, the default handler will be used.
// This is useful when you want to have a central place to handle all actions.
func (f *Flow) WithDefaultActionHandler(handler ActionHandler) *Flow {
	f.defaultHandler = handler
	return f
}

func (f *Flow) ID() string {
	return f.id
}

func (f *Flow) Type() FlowType {
	return f.flowType
}

func (f *Flow) CurrentState() State {
	return f.currentState
}

func (f *Flow) Data() FlowData {
	return f.data
}

func (f *Flow) TransitionTable() TransitionTable {
	return f.states
}

func (f *Flow) IsCompleted() bool {
	return f.completed
}

func (f *Flow) IsExpired() bool {
	return !f.expiresAt.IsZero() && time.Now().After(f.expiresAt)
}

func (f *Flow) ExpiresAt() time.Time {
	return f.expiresAt
}

func (f *Flow) SetExpirationAt(t time.Time) {
	f.expiresAt = t
}

func (f *Flow) SetExpirationIn(t time.Duration) {
	f.expiresAt = time.Now().Add(t)
}

func (f *Flow) logf(format string, a ...any) {
	if isDebugging {
		fmt.Printf("[%s %s]: ", f.flowType, f.id)
		fmt.Printf(format, a...)
	}
}

func logf(format string, a ...any) {
	if isDebugging {
		fmt.Print("[Flow]: ")
		fmt.Printf(format, a...)
	}
}

// DebugMode enables or disables the debug mode.
// When debug mode is enabled, the flow will print out debug messages.
func DebugMode(enabled bool) {
	isDebugging = enabled
}

// MustCast casts the data to the specified type.
func MustCast[D FlowData](data interface{}) D {
	if d, ok := data.(D); ok {
		return d
	}
	panic(fmt.Sprintf("cannot cast data to type: %T", data))
}
