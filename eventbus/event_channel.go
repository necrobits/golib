package eventbus

type EventChannel chan Event
type EventChannels []EventChannel

func NewEventChannel() EventChannel {
	return make(EventChannel)
}
