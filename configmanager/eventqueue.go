package configmanager

import "github.com/necrobits/x/eventbus"

type EventQueue []eventbus.Event

func (q *EventQueue) add(topic Topic, data interface{}) {
	*q = append(*q, eventbus.NewEvent(topic, data))
}
