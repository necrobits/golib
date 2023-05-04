package flow

import "fmt"

type typedActionHandler[A Action] func(data FlowData, a A) (Event, FlowData, error)
type ActionRoutes map[ActionType]ActionHandler

// TypedHandler is a helper function to create a typed action handler.
// It is useful when you know the type of the action beforehand and you want to avoid type assertions.
// For example:
// Instead of writing:
//
//	func (f *OrderFlowCreator) HandleCancelation(state flow.FlowData, a Action) (flow.Event, flow.FlowData, error) {
//		if a, ok := a.(PaymentAction); ok {
//			//Do something...
//		}
//	}
//
// You can write like this and access the action directly without type assertions:
//
//	func (f *OrderFlowCreator) HandlePayment(state flow.FlowData, a PaymentAction) (flow.Event, flow.FlowData, error) {
//		//Do something...
//	}
func TypedHandler[A Action](f typedActionHandler[A]) ActionHandler {
	return func(data FlowData, a Action) (Event, FlowData, error) {
		if a, ok := a.(A); ok {
			return f(data, a)
		}
		return "", data, fmt.Errorf("invalid action type: %T", a)
	}
}

// ActionRouter is a helper to route actions to handlers.
// It is useful when you want to handle multiple action types in a single handler.
// You can create multiple typed handlers and add them to the router, then use the router as a single handler.
// For example:
//
//	router := flow.NewRouter(flow.ActionRoutes{
//					PayForOrder: flow.TypedHandler(f.HandlePayment),
//					CancelOrder: flow.TypedHandler(f.HandleCancelation),
//				})
//	table := flow.TransitionTable{
//		AwaitingPayment: flow.StateConfig{
//			Handler: router.ToHandler(),
//			Transitions: flow.Transitions{
//				OrderPaid:     AwaitingShipping,
//				OrderCanceled: Canceled,
//			},
//		},
//	}
type ActionRouter struct {
	routes ActionRoutes
}

// NewRouter creates a new ActionRouter.
// You can also add routes to the router later using AddRoute or AddRoutes.
func NewRouter(routes ActionRoutes) *ActionRouter {
	return &ActionRouter{routes: routes}
}

// Handle handles an action using the router.
// You should not call this method directly, instead you should use the router as a handler.
func (r *ActionRouter) Handle(data FlowData, a Action) (Event, FlowData, error) {
	if handler, ok := r.routes[a.Type()]; ok {
		return handler(data, a)
	}
	return "", data, fmt.Errorf("no handler for action type: %s", a.Type())
}

// AddRoute adds a route to the router.
func (r *ActionRouter) AddRoute(actionType ActionType, handler ActionHandler) {
	r.routes[actionType] = handler
}

// AddRoutes adds multiple routes to the router.
func (r *ActionRouter) AddRoutes(routes ActionRoutes) {
	for actionType, handler := range routes {
		r.routes[actionType] = handler
	}
}

// ToHandler converts the router to a single handler.
// You should call this method after adding all the routes to convert the router to a handler.
func (r *ActionRouter) ToHandler() ActionHandler {
	return func(data FlowData, a Action) (Event, FlowData, error) {
		return r.Handle(data, a)
	}
}
