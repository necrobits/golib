package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/necrobits/x/flow"
	"github.com/necrobits/x/flow/flowregistry"
	"github.com/necrobits/x/flowviz"
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
			Autopass: true,
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
		ExpireIn:        -time.Hour * 24 * 7,
	})
}

func (f *OrderFlowCreator) NewFlowFromSnapshot(s *flow.Snapshot) *flow.Flow {
	return flow.FromSnapshot(s, f.transTable)
}

func (f *OrderFlowCreator) HandleCancelation(ctx context.Context, state *OrderInternalState, a CancelAction) (flow.Event, *OrderInternalState, error) {
	state.CanceledAt = time.Now().Unix()
	return OrderCanceled, state, nil
}

func (f *OrderFlowCreator) HandlePayment(ctx context.Context, state *OrderInternalState, payment PaymentAction) (flow.Event, *OrderInternalState, error) {
	if payment.Amount != state.TotalAmount {
		return flow.NoEvent, nil, fmt.Errorf("payment amount does not match order total")
	}
	state.Paid = true
	return OrderPaid, state, nil
}

func (f *OrderFlowCreator) HandleShipping(ctx context.Context, state *OrderInternalState, a flow.Action) (flow.Event, *OrderInternalState, error) {
	return OrderShipped, state, nil
}

func main() {
	flow.DebugMode(true)
	ctx := context.Background()
	orderFlowCreator := NewOrderFlowCreator()

	orderFlow := orderFlowCreator.NewFlow("123", 100)
	orderFlow.RegisterPreTransition(AwaitingShipping, flow.TypedPreHook(func(data *OrderInternalState) error {
		fmt.Printf("[PRE HOOK] Paid, awaiting shipping: %s\n", data.OrderID)
		return nil
	}))
	orderFlow.RegisterPostTransition(OrderFulfilled, flow.TypedHook(func(data *OrderInternalState) {
		fmt.Printf("[POST HOOK] Order fulfilled: %s\n", data.OrderID)
	}))
	orderFlow.RegisterCompletionHook(OrderFulfilled, flow.TypedHook(func(data *OrderInternalState) {
		fmt.Printf("[COMPLETION HOOK] Order fulfilled: %s\n", data.OrderID)
	}))
	flowregistry.Global().Register("OrderFlow", OrderInternalState{})
	flow.HookRegistry().RegisterPostTransition("OrderFlow", AwaitingShipping, flow.TypedHook(func(data *OrderInternalState) {
		fmt.Printf("[GLOBAL POST HOOK] Paid, awaiting shipping: %s\n", data.OrderID)
	}))
	flow.HookRegistry().RegisterCompletion("OrderFlow", OrderFulfilled, flow.TypedHook(func(data *OrderInternalState) {
		fmt.Printf("[GLOBAL COMPLETION HOOK] Order fulfilled: %s\n", data.OrderID)
	}))
	err := orderFlow.HandleAction(ctx, PaymentAction{Amount: 100})
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	// Dont need to call this, since it's autopass
	//err = orderFlow.HandleAction(ShipOrderAction{})
	//if err != nil {
	//	fmt.Printf("Error: %s\n", err)
	//}
	fmt.Println("data", flow.MustCast[*OrderInternalState](orderFlow.Data()))
	fmt.Printf("Is completed: %t\n", orderFlow.IsCompleted())

	snapshot, err := orderFlow.ToSnapshot()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Printf("Snapshot: %s\n", string(snapshot.EncodedData))
	var buf bytes.Buffer
	flowviz.CreateGraphvizForFlow(orderFlowCreator.transTable, flowviz.VizFormatPNG, &buf)
	os.WriteFile("flow.png", buf.Bytes(), 0644)

	snapshot2, err := flowregistry.Global().DecodeSnapshot(snapshot)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Printf("Snapshot2: %+v\n", snapshot2.Data.(*OrderInternalState))
	//fmt.Printf("Graphviz:\n%s\n", buf.String())
}
