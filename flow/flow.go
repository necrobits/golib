package flow

import (
	"database/sql"
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
	id             string
	flowType       string
	data           FlowData
	currentState   State
	states         TransitionTable
	defaultHandler ActionHandler
	expireAt       time.Time
	isCompleted    bool
}

// FlowData is the internal data of the flow. This can be anything.
type FlowData interface{}

type TransitionTable map[State]StateConfig

// Action is the interface that all actions must implement.
type Action interface {
	Type() ActionType
}

// ActionType is the type of the action. It is used to identify the action in the TransitionTable.
type ActionType string

// Event is the input to the state machine. It is the output of the action handler.
// Hence, it is only a string. The action handler is responsible for changing the data of the flow.
// It causes the state machine to change its state.
type Event string

// State is just a name for a state in the state machine.
type State string

// CreateFlowOpts is used to create a new Flow
// ID is the unique identifier for the flow.
// Type is the type of the flow.
// Both ID and Type are used to identify the flow and restore it from a snapshot.
//
// Data is the initial internal data for the flow.
// InitialState is the initial state for the flow. It must be one of the states in TransitionTable.
// TransitionTable contains the description of the state machine for the flow.
// Handler is the default handler for the flow. If a state does not have a handler, this handler is used.
// ExpireAt is the time at which the flow expires.
// ExpireIn is the duration after which the flow expires.
// If both ExpireAt and ExpireIn are set, ExpireIn is used.
type CreateFlowOpts struct {
	ID              string
	Type            string
	Data            FlowData
	InitialState    State
	TransitionTable TransitionTable
	Handler         ActionHandler
	ExpireAt        time.Time
	ExpireIn        time.Duration
}

// Snapshot is used to persist the flow, and restore it later.
type Snapshot struct {
	ID           string       `json:"id"`
	Type         string       `json:"type"`
	Data         FlowData     `json:"data"`
	CurrentState State        `json:"current_state"`
	ExpireAt     sql.NullTime `json:"expire_at"`
	IsCompleted  bool         `json:"is_completed"`
}

// ActionHandler is the function that handles an action.
// It returns:
//   - an Event, which is the input to the state machine.
//   - the new internal data of the flow, which replaces the old one.
//
// If the error is not nil, no event will be sent to the state machine, and the flow will remain in the same state.
// If the error is nil, the flow will change its state according to the the transition table.
type ActionHandler func(data FlowData, a Action) (Event, FlowData, error)

// Transitions is a map, describing the transition from a state to the next state.
type Transitions map[Event]State

// StateConfig is the configuration for a state in the state machine.
// In addition to just a name, a State contains a handler function, and a map of transitions.
// When the flow is in a state, and an action is received, the handler function of that state is called.
// If you are going to handle one action in a state, you may consider using the [TypedHandler] function to avoid type assertion.
// If you are going to handle multiple actions in a state, you may consider using the [NewRouter] function create an action router.
type StateConfig struct {
	Handler     ActionHandler
	Transitions Transitions
	Final       bool
}

// HandleAction handles an action for the flow.
// Everytime an action is handled, the flow may change its state.
// This function is the only way to change the state of the flow.
func (f *Flow) HandleAction(a Action) error {
	if f.isCompleted {
		return fmt.Errorf("flow is completed")
	}
	if f.IsExpired() {
		return fmt.Errorf("flow already expired")
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
	inputEvent, nextData, err := actionHandler(f.data, a)
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
		f.logf("Transition: %s -> %s\n", f.currentState, nextState)
	}
	if nextStateConfig, ok := f.states[nextState]; ok {
		if nextStateConfig.Final {
			f.isCompleted = true
			f.logf("Flow completed\n")
		}
	}
	f.data = nextData
	f.currentState = nextState
	return nil
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
		expireAt:       opts.ExpireAt,
	}
}

// ToSnapshot converts the flow to a Snapshot to be persisted.
func (f *Flow) ToSnapshot() Snapshot {
	return Snapshot{
		ID:           f.id,
		Type:         f.flowType,
		Data:         f.data,
		CurrentState: f.currentState,
		ExpireAt: sql.NullTime{
			Time:  f.expireAt,
			Valid: !f.expireAt.IsZero(),
		},
		IsCompleted: f.isCompleted,
	}
}

// FromSnapshot restores a flow from a Snapshot.
func FromSnapshot(s *Snapshot, stateMap TransitionTable) *Flow {
	flow := Flow{
		id:           s.ID,
		flowType:     s.Type,
		data:         s.Data,
		currentState: s.CurrentState,
		states:       stateMap,
		isCompleted:  s.IsCompleted,
	}
	if s.ExpireAt.Valid {
		flow.expireAt = s.ExpireAt.Time
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

func (f *Flow) WithExpiration(t time.Duration) *Flow {
	f.expireAt = time.Now().Add(t)
	return f
}

func (f *Flow) ID() string {
	return f.id
}

func (f *Flow) Type() string {
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
	return f.isCompleted
}

func (f *Flow) IsExpired() bool {
	return !f.expireAt.IsZero() && time.Now().After(f.expireAt)
}

func (f *Flow) ExpiresAt() time.Time {
	return f.expireAt
}

func (f *Flow) logf(format string, a ...any) {
	if isDebugging {
		fmt.Printf("[%s %s]: ", f.flowType, f.id)
		fmt.Printf(format, a...)
	}
}

func (f *Flow) log(msg string) {
	if isDebugging {
		fmt.Printf("[%s %s]: ", f.flowType, f.id)
		fmt.Println(msg)
	}
}

func log(msg string) {
	if isDebugging {
		fmt.Print("[Flow]: ")
		fmt.Println(msg)
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
