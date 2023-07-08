package flow

var (
	globalHookRegistry = newHookRegistry()
)

type hookRegistryType[T any] map[FlowType]map[State][]T
type preTransitionRegistry hookRegistryType[hookFn]
type postTransitionRegistry hookRegistryType[silentHookFn]
type hydrationRegistry map[FlowType][]hydrationHookFn
type completionRegistry map[FlowType][]silentHookFn
type hookRegistry struct {
	hydrationHooks      hydrationRegistry
	preTransitionHooks  preTransitionRegistry
	postTransitionHooks postTransitionRegistry
	completionHooks     completionRegistry
}

func newHookRegistry() *hookRegistry {
	return &hookRegistry{
		hydrationHooks:      make(hydrationRegistry),
		preTransitionHooks:  make(preTransitionRegistry),
		postTransitionHooks: make(postTransitionRegistry),
		completionHooks:     make(completionRegistry),
	}
}

func (r *hookRegistry) composeHydrationHooks(flowType FlowType) hydrationHookFn {
	if _, ok := r.hydrationHooks[flowType]; !ok {
		return nil
	}
	return composeHydrationHooks(r.hydrationHooks[flowType])
}

func (r *hookRegistry) composePreTransitions(flowType FlowType, state State) hookFn {
	if _, ok := r.preTransitionHooks[flowType]; !ok {
		return nil
	}
	if _, ok := r.preTransitionHooks[flowType][state]; !ok {
		return nil
	}
	return composePreTransitionHooks(r.preTransitionHooks[flowType][state])
}

func (r *hookRegistry) composePostTransitionHooks(flowType FlowType, state State) silentHookFn {
	if _, ok := r.postTransitionHooks[flowType]; !ok {
		return nil
	}
	if _, ok := r.postTransitionHooks[flowType][state]; !ok {
		return nil
	}
	return composeSilentHooks(r.postTransitionHooks[flowType][state])
}

func (r *hookRegistry) composeCompletionHooks(flowType FlowType) silentHookFn {
	if _, ok := r.completionHooks[flowType]; !ok {
		return nil
	}
	return composeSilentHooks(r.completionHooks[flowType])
}

func (r *hookRegistry) RegisterHydration(flowType FlowType, hook hydrationHookFn) {
	if _, ok := r.hydrationHooks[flowType]; !ok {
		r.hydrationHooks[flowType] = []hydrationHookFn{}
	}
	r.hydrationHooks[flowType] = append(r.hydrationHooks[flowType], hook)
}

func (r *hookRegistry) RegisterPreTransition(flowType FlowType, state State, hook hookFn) {
	r.preTransitionHooks = addHookToRegistry[hookFn](r.preTransitionHooks, flowType, state, hook)
}

func (r *hookRegistry) RegisterPostTransition(flowType FlowType, state State, hook silentHookFn) {
	r.postTransitionHooks = addHookToRegistry[silentHookFn](r.postTransitionHooks, flowType, state, hook)
}

func (r *hookRegistry) RegisterCompletion(flowType FlowType, hook silentHookFn) {
	if _, ok := r.completionHooks[flowType]; !ok {
		r.completionHooks[flowType] = []silentHookFn{}
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
