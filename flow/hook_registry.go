package flow

var (
	globalHookRegistry = newHookRegistry()
)

type hookRegistryType[T any] map[FlowType]map[State][]T
type preTransitionRegistry hookRegistryType[preTransitionHook]
type postTransitionRegistry hookRegistryType[silentHook]
type completionRegistry map[FlowType][]silentHook
type hookRegistry struct {
	preTransitionHooks  preTransitionRegistry
	postTransitionHooks postTransitionRegistry
	completionHooks     completionRegistry
}

func newHookRegistry() *hookRegistry {
	return &hookRegistry{
		preTransitionHooks:  make(preTransitionRegistry),
		postTransitionHooks: make(postTransitionRegistry),
		completionHooks:     make(completionRegistry),
	}
}

func (r *hookRegistry) composePreTransitions(flowType FlowType, state State) preTransitionHook {
	if _, ok := r.preTransitionHooks[flowType]; !ok {
		return nil
	}
	if _, ok := r.preTransitionHooks[flowType][state]; !ok {
		return nil
	}
	return composePreTransitionHooks(r.preTransitionHooks[flowType][state])
}

func (r *hookRegistry) composePostTransitionHooks(flowType FlowType, state State) silentHook {
	if _, ok := r.postTransitionHooks[flowType]; !ok {
		return nil
	}
	if _, ok := r.postTransitionHooks[flowType][state]; !ok {
		return nil
	}
	return composeSilentHooks(r.postTransitionHooks[flowType][state])
}

func (r *hookRegistry) composeCompletionHooks(flowType FlowType) silentHook {
	if _, ok := r.completionHooks[flowType]; !ok {
		return nil
	}
	return composeSilentHooks(r.completionHooks[flowType])
}

func (r *hookRegistry) RegisterPreTransition(flowType FlowType, state State, hook preTransitionHook) {
	r.preTransitionHooks = addHookToRegistry[preTransitionHook](r.preTransitionHooks, flowType, state, hook)
}

func (r *hookRegistry) RegisterPostTransition(flowType FlowType, state State, hook silentHook) {
	r.postTransitionHooks = addHookToRegistry[silentHook](r.postTransitionHooks, flowType, state, hook)
}

func (r *hookRegistry) RegisterCompletion(flowType FlowType, state State, hook silentHook) {
	if _, ok := r.completionHooks[flowType]; !ok {
		r.completionHooks[flowType] = []silentHook{}
	}
	r.completionHooks[flowType] = append(r.completionHooks[flowType], hook)
}

func HookRegistry() *hookRegistry {
	return globalHookRegistry
}

func addHookToRegistry[T any](registry map[FlowType]map[State][]T, flowType FlowType, state State, hook T) map[FlowType]map[State][]T {
	if registry == nil {
		registry = make(map[FlowType]map[State][]T)
	}
	if _, ok := registry[flowType]; !ok {
		registry[flowType] = make(map[State][]T)
	}
	if _, ok := registry[flowType][state]; !ok {
		registry[flowType][state] = []T{}
	}
	registry[flowType][state] = append(registry[flowType][state], hook)
	return registry
}
