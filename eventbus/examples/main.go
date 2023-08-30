package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/necrobits/x/eventbus"
)

const (
	Topic1 eventbus.Topic = "topic1"
	Topic2 eventbus.Topic = "topic2"
)

func printDataEvent(ch string, event eventbus.Event) {
	fmt.Printf("Channel: %s; Topic: %s; DataEvent: %v\n", ch, event.Topic(), event.Data())
}

func main() {
	eb := eventbus.NewEventBus()

	ch1 := eventbus.NewEventChannel()
	ch2 := eventbus.NewEventChannel()

	eb.Subscribe(Topic1, ch1)
	eb.Subscribe(Topic2, ch2)

	publisTo := func(topic eventbus.Topic, data string) {
		for {
			eb.Publish(topic, data)
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
	}

	go publisTo(Topic1, "Hello to topic1")
	go publisTo(Topic2, "Goodbye from topic2")

	for {
		select {
		case d := <-ch1:
			printDataEvent("ch1", d)
		case d := <-ch2:
			printDataEvent("ch2", d)
		}
	}
}
