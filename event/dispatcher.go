package event

type Dispatcher struct {
	listeners []Listener
}

func NewDispatcher(listeners ...Listener) *Dispatcher {
	return &Dispatcher{
		listeners: listeners,
	}
}

func (d *Dispatcher) Dispatch(event Event) error {
	for _, listener := range d.listeners {
		if err := listener.HandleEvent(event); err != nil {
			return err
		}
	}
	return nil
}
