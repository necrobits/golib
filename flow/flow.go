package flow

import (
	"fmt"
)

type Flow struct {
	id           string
	flowType     string
	data         FlowData
	currentState State
	states       TransitionTable
}

// FlowData is the internal data of the flow. This can be anything.
type FlowData interface{}

type TransitionTable map[State]StateConfig

// Action is the interface that all actions must implement.
type Action interface {
	Type() ActionType
}

// ActionType is the type of the action. It is used to identify the action in the FSMMap.
type ActionType string

// CreateFlowOpts is used to create a new Flow
// ID is the unique identifier for the flow.
// Type is the type of the flow.
// Both ID and Type are used to identify the flow and restore it from a snapshot.
//
// Data is the initial internal data for the flow.
// InitialState is the initial state for the flow. It must be one of the states in FSMMap.
// FSMMap contains the description of the state machine for the flow.
type CreateFlowOpts struct {
	ID              string
	Type            string
	Data            FlowData
	InitialState    State
	TransitionTable TransitionTable
}

// Snapshot is used to persist the flow, and restore it later.
type Snapshot struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Data         FlowData `json:"data"`
	CurrentState State    `json:"current_state"`
}

// ActionHandler is the function that handles an action. It can modify the internal data of the flow.
// It returns the new internal data, and an error if any.
// If the error is not nil, the flow will not change its state.
// If the error is nil, the flow will change its state according to the transitions in FSMMap.
type ActionHandler func(data FlowData, a Action) (FlowData, error)

// State is just a name for a state in the state machine.
type State string

// Transitions is a map, describing the transition from a state to the next state.
type Transitions map[ActionType]State

// StateConfig is the configuration for a state in the state machine.
// In addition to just a name, a State contains a handler function, and a map of transitions.
// When the flow is in a state, and an action is received, the handler function of that state is called.
type StateConfig struct {
	Handler     ActionHandler
	Transitions Transitions
}

// HandleAction handles an action for the flow.
// Everytime an action is handled, the flow may change its state.
// This function is the only way to change the state of the flow.
func (f *Flow) HandleAction(a Action) error {
	actionType := a.Type()
	stateConfig, ok := f.states[f.currentState]
	if !ok {
		return fmt.Errorf("illegal state: %s", f.currentState)
	}
	actionHandler := stateConfig.Handler
	nextState, ok := stateConfig.Transitions[actionType]
	if !ok {
		return fmt.Errorf("no transition found for event: %s", actionType)
	}
	if actionHandler != nil {
		newData, err := actionHandler(f.data, a)
		if err != nil {
			return err
		}
		f.data = newData
	}
	f.currentState = nextState
	return nil
}

// New creates a new Flow.
func New(opts CreateFlowOpts) *Flow {
	if opts.TransitionTable == nil {
		panic("FSMMap cannot be nil")
	}
	if opts.InitialState == "" {
		panic("InitialState cannot be empty")
	}
	return &Flow{
		id:           opts.ID,
		flowType:     opts.Type,
		data:         opts.Data,
		currentState: opts.InitialState,
		states:       opts.TransitionTable,
	}
}

// ToSnapshot converts the flow to a Snapshot to be persisted.
func (f *Flow) ToSnapshot() Snapshot {
	return Snapshot{
		ID:           f.id,
		Type:         f.flowType,
		Data:         f.data,
		CurrentState: f.currentState,
	}
}

// FromSnapshot restores a flow from a Snapshot.
func FromSnapshot(s *Snapshot, stateMap TransitionTable) *Flow {
	return &Flow{
		id:           s.ID,
		flowType:     s.Type,
		data:         s.Data,
		currentState: s.CurrentState,
		states:       stateMap,
	}
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
