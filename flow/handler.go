package flow

import "fmt"

type typedActionHandler[A Action] func(data FlowData, a A) (Event, FlowData, error)
type ActionRoutes map[ActionType]ActionHandler

func TypedHandler[A Action](f typedActionHandler[A]) ActionHandler {
	return func(data FlowData, a Action) (Event, FlowData, error) {
		if a, ok := a.(A); ok {
			return f(data, a)
		}
		return "", data, fmt.Errorf("invalid action type: %T", a)
	}
}

type ActionRouter struct {
	routes ActionRoutes
}

func NewRouter(routes ActionRoutes) *ActionRouter {
	return &ActionRouter{routes: routes}
}

func (r *ActionRouter) Handle(data FlowData, a Action) (Event, FlowData, error) {
	if handler, ok := r.routes[a.Type()]; ok {
		return handler(data, a)
	}
	return "", data, fmt.Errorf("no handler for action type: %s", a.Type())
}

func (r *ActionRouter) AddRoute(actionType ActionType, handler ActionHandler) {
	r.routes[actionType] = handler
}

func (r *ActionRouter) AddRoutes(routes ActionRoutes) {
	for actionType, handler := range routes {
		r.routes[actionType] = handler
	}
}

func (r *ActionRouter) ToHandler() ActionHandler {
	return func(data FlowData, a Action) (Event, FlowData, error) {
		return r.Handle(data, a)
	}
}
