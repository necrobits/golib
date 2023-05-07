package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/necrobits/golib/flow"
	"github.com/necrobits/golib/flowviz"
)

const (
	AwaitingPayment  flow.State = "AwaitingPayment"
	AwaitingShipping flow.State = "AwaitingShipping"
	OrderFulfilled   flow.State = "OrderFulfilled"
	Canceled         flow.State = "Canceled"

	PayForOrder flow.ActionType = "PayForOrder"
	ShipOrder   flow.ActionType = "ShipOrder"
	CancelOrder flow.ActionType = "CancelOrder"

	OrderPaid     flow.Event = "OrderPaid"
	OrderShipped  flow.Event = "OrderShipped"
	OrderCanceled flow.Event = "OrderCanceled"
)

type OrderInternalState struct {
	OrderID     string
	TotalAmount int
	Paid        bool
	CanceledAt  int64
}

type PaymentAction struct {
	Amount int
}

type CancelAction struct{}

type ShipOrderAction struct{}

func (p ShipOrderAction) Type() flow.ActionType {
	return ShipOrder
}

func (p CancelAction) Type() flow.ActionType {
	return CancelOrder
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
			Handler: flow.NewRouter(flow.ActionRoutes{
				PayForOrder: flow.TypedHandler(f.HandlePayment),
				CancelOrder: flow.TypedHandler(f.HandleCancelation),
			}).ToHandler(),
			Transitions: flow.Transitions{
				OrderPaid:     AwaitingShipping,
				OrderCanceled: Canceled,
			},
		},
		AwaitingShipping: flow.StateConfig{
			Handler: flow.TypedHandler(f.HandleShipping),
			Transitions: flow.Transitions{
				OrderShipped: OrderFulfilled,
			},
		},
		OrderFulfilled: flow.StateConfig{
			Final: true,
		},
	}
	return f
}

func (f *OrderFlowCreator) NewFlow(orderId string, amount int) *flow.Flow {
	return flow.New(flow.CreateFlowOpts{
		ID:              "abc123",
		Type:            "OrderFlow",
		Data:            &OrderInternalState{OrderID: orderId, TotalAmount: amount},
		InitialState:    AwaitingPayment,
		TransitionTable: f.transTable,
	})
}

func (f *OrderFlowCreator) NewFlowFromSnapshot(s *flow.Snapshot) *flow.Flow {
	return flow.FromSnapshot(s, f.transTable)
}

func (f *OrderFlowCreator) HandleCancelation(state *OrderInternalState, a CancelAction) (flow.Event, *OrderInternalState, error) {
	state.CanceledAt = time.Now().Unix()
	return OrderCanceled, state, nil
}

func (f *OrderFlowCreator) HandlePayment(state *OrderInternalState, payment PaymentAction) (flow.Event, *OrderInternalState, error) {
	if payment.Amount != state.TotalAmount {
		return flow.NoEvent, nil, fmt.Errorf("payment amount does not match order total")
	}
	state.Paid = true
	return OrderPaid, state, nil
}

func (f *OrderFlowCreator) HandleShipping(state *OrderInternalState, a ShipOrderAction) (flow.Event, *OrderInternalState, error) {
	actionType := a.Type()
	if actionType == ShipOrder {
		return OrderShipped, state, nil
	}
	return flow.NoEvent, nil, fmt.Errorf("invalid action")
}

func main() {
	flow.DebugMode(true)
	orderFlowCreator := NewOrderFlowCreator()

	orderFlow := orderFlowCreator.NewFlow("123", 100)

	err := orderFlow.HandleAction(PaymentAction{Amount: 100})
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	err = orderFlow.HandleAction(ShipOrderAction{})
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	fmt.Println("data", flow.MustCast[*OrderInternalState](orderFlow.Data()))

	b, err := json.Marshal(orderFlow.ToSnapshot())
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Printf("Snapshot: %s\n", string(b))
	var buf bytes.Buffer
	flowviz.CreateGraphvizForFlow(orderFlowCreator.transTable, flowviz.VizFormatPNG, &buf)
	os.WriteFile("flow.png", buf.Bytes(), 0644)
	//fmt.Printf("Graphviz:\n%s\n", buf.String())
}
