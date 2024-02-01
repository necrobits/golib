package old_configmanager

import "github.com/necrobits/x/event"

type EventQueue []event.Event

func (q *EventQueue) add(topic event.Topic, data interface{}) {
	*q = append(*q, event.NewEvent(topic, data))
}
