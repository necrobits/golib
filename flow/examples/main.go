package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/necrobits/golib/flow"
	"github.com/necrobits/golib/flowviz"
)

const (
	AwaitingPayment  flow.State = "AwaitingPayment"
	AwaitingShipping flow.State = "AwaitingShipping"
	OrderFulfilled   flow.State = "OrderFulfilled"

	PayForOrder flow.ActionType = "PayForOrder"
	OrderPaid   flow.Event      = "OrderPaid"
)

type OrderInternalState struct {
	OrderID     string
	TotalAmount int
	Paid        bool
}

type PaymentAction struct {
	Amount int
}

func (p PaymentAction) Type() flow.ActionType {
	return PayForOrder
}

type OrderFlowCreator struct {
	transTable flow.TransitionTable
}

func NewOrderFlowCreator() *OrderFlowCreator {
	f := &OrderFlowCreator{}
	f.transTable = flow.TransitionTable{
		AwaitingPayment: flow.StateConfig{
			Handler: f.HandlePayment,
			Transitions: flow.Transitions{
				OrderPaid: AwaitingShipping,
			},
		},
	}
	return f
}

func (f *OrderFlowCreator) NewFlow(orderId string, amount int) *flow.Flow {
	return flow.New(flow.CreateFlowOpts{
		ID:              "flowID123",
		Type:            "OrderFlow",
		Data:            OrderInternalState{OrderID: orderId, TotalAmount: amount},
		InitialState:    AwaitingPayment,
		TransitionTable: f.transTable,
	})
}

func (f *OrderFlowCreator) NewFlowFromSnapshot(s *flow.Snapshot) *flow.Flow {
	return flow.FromSnapshot(s, f.transTable)
}

func (f *OrderFlowCreator) HandlePayment(state flow.FlowData, a flow.Action) (flow.Event, flow.FlowData, error) {
	state = state.(OrderInternalState)
	payment := a.(PaymentAction)
	if payment.Amount != state.(OrderInternalState).TotalAmount {
		return flow.NoEvent, nil, fmt.Errorf("payment amount does not match order total")
	}
	newState := state.(OrderInternalState)
	newState.Paid = true
	return OrderPaid, newState, nil
}

func main() {
	orderFlowCreator := NewOrderFlowCreator()

	orderFlow := orderFlowCreator.NewFlow("123", 100)

	fmt.Printf("Current State: %s\n", orderFlow.CurrentState())

	err := orderFlow.HandleAction(PaymentAction{Amount: 100})
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Printf("Current State: %s\n", orderFlow.CurrentState())
	b, err := json.Marshal(orderFlow.ToSnapshot())
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Printf("Snapshot: %s\n", string(b))
	var buf bytes.Buffer
	flowviz.CreateGraphvizForFlow(orderFlowCreator.transTable, flowviz.VizFormatDot, &buf)
	fmt.Printf("Graphviz:\n%s\n", buf.String())
}
