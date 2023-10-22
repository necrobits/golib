package event

type Topic string

type Event struct {
	topic Topic
	data  interface{}
}

func NewEvent(topic Topic, data interface{}) Event {
	return Event{topic: topic, data: data}
}

func (e Event) Topic() Topic {
	return e.topic
}

func (e Event) Data() interface{} {
	return e.data
}
