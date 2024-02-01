package event

import "sync"

type EventChannel chan Event
type EventChannels []EventChannel

func NewEventChannel() EventChannel {
	return make(EventChannel)
}

type EventBus struct {
	rm            sync.RWMutex
	subscribtions map[Topic]EventChannels
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribtions: make(map[Topic]EventChannels),
	}
}

func (eb *EventBus) Subscribe(topic Topic, ch EventChannel) {
	eb.rm.Lock()
	defer eb.rm.Unlock()
	if _, found := eb.subscribtions[topic]; !found {
		eb.subscribtions[topic] = EventChannels{ch}
	} else {
		eb.subscribtions[topic] = append(eb.subscribtions[topic], ch)
	}
}

func (eb *EventBus) Publish(topic Topic, data interface{}) {
	eb.rm.RLock()
	defer eb.rm.RUnlock()
	if chans, found := eb.subscribtions[topic]; found {
		channels := make(EventChannels, len(chans))
		copy(channels, chans)
		go func(channels EventChannels, data interface{}) {
			for _, ch := range channels {
				ch <- Event{topic, data}
			}
		}(channels, data)
	}
}

func (eb *EventBus) HasSubscribers(topic Topic) bool {
	_, found := eb.subscribtions[topic]
	return found
}
