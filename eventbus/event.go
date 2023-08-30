package eventbus

type Topic string

type Event struct {
	topic Topic
	data  interface{}
}

func (e Event) Topic() Topic {
	return e.topic
}

func (e Event) Data() interface{} {
	return e.data
}
