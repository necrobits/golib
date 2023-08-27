package eventbus

type Event struct {
	topic string
	data  interface{}
}

func (e Event) Topic() string {
	return e.topic
}

func (e Event) Data() interface{} {
	return e.data
}
