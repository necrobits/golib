package flow

import (
	"context"
	"fmt"
)

type typedActionHandler[D FlowData, A Action] func(ctx context.Context, data D, a A) (Event, D, error)
type ActionRoutes map[ActionType]ActionHandler

// TypedHandler is a helper function to create a typed action handler.
// It is useful when you know the type of the action beforehand and you want to avoid type assertions.
// For example:
// Instead of writing:
//
//	func (f *OrderFlowCreator) HandleCancelation(ctx context.Context, state flow.FlowData, a Action) (flow.Event, flow.FlowData, error) {
//		if a, ok := a.(PaymentAction); ok {
//			//Do something...
//		}
//	}
//
// You can write like this and access the action directly without type assertions:
//
//	func (f *OrderFlowCreator) HandlePayment(ctx context.Context, state flow.FlowData, a PaymentAction) (flow.Event, flow.FlowData, error) {
//		//Do something...
//	}
func TypedHandler[D FlowData, A Action](f typedActionHandler[D, A]) ActionHandler {
	return func(ctx context.Context, data FlowData, a Action) (Event, FlowData, error) {
		var castedA A
		var castedData D
		var ok bool
		if castedData, ok = data.(D); !ok {
			return "", data, fmt.Errorf("invalid data type: %T", data)
		}
		if castedA, ok = a.(A); !ok {
			return "", data, fmt.Errorf("invalid action type: %T", a)
		}
		return f(ctx, castedData, castedA)
	}
}

// actionRouter is a helper to route actions to handlers.
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
type actionRouter struct {
	routes ActionRoutes
}

// NewRouter creates a new ActionRouter.
// You can also add routes to the router later using AddRoute or AddRoutes.
func NewRouter(routes ActionRoutes) *actionRouter {
	return &actionRouter{routes: routes}
}

// Handle handles an action using the router.
// You should not call this method directly, instead you should use the router as a handler.
func (r *actionRouter) Handle(ctx context.Context, data FlowData, a Action) (Event, FlowData, error) {
	if handler, ok := r.routes[a.Type()]; ok {
		return handler(ctx, data, a)
	}
	return "", data, fmt.Errorf("no handler for action type: %s", a.Type())
}

// AddRoute adds a route to the router.
func (r *actionRouter) AddRoute(actionType ActionType, handler ActionHandler) {
	r.routes[actionType] = handler
}

// AddRoutes adds multiple routes to the router.
func (r *actionRouter) AddRoutes(routes ActionRoutes) {
	for actionType, handler := range routes {
		r.routes[actionType] = handler
	}
}

// ToHandler converts the router to a single handler.
// You should call this method after adding all the routes to convert the router to a handler.
func (r *actionRouter) ToHandler() ActionHandler {
	return func(ctx context.Context, data FlowData, a Action) (Event, FlowData, error) {
		return r.Handle(ctx, data, a)
	}
}
