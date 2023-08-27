package eventbus

import "sync"

type EventBus struct {
	rm          sync.RWMutex
	subscribers map[string]EventChannels
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string]EventChannels),
	}
}

func (eb *EventBus) Subscribe(topic string, ch EventChannel) {
	eb.rm.Lock()
	defer eb.rm.Unlock()
	if _, found := eb.subscribers[topic]; !found {
		eb.subscribers[topic] = EventChannels{ch}
	} else {
		eb.subscribers[topic] = append(eb.subscribers[topic], ch)
	}
}

func (eb *EventBus) Publish(topic string, data interface{}) {
	eb.rm.RLock()
	defer eb.rm.RUnlock()
	if chans, found := eb.subscribers[topic]; found {
		channels := make(EventChannels, len(chans))
		copy(channels, chans)
		go func(channels EventChannels, data interface{}) {
			for _, ch := range channels {
				ch <- Event{topic, data}
			}
		}(channels, data)
	}
}
