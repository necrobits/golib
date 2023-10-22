package event

type Listener interface {
	HandleEvent(Event) error
}
