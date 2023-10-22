package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/necrobits/x/event"
)

const (
	Topic1 event.Topic = "topic1"
	Topic2 event.Topic = "topic2"
)

func printDataEvent(ch string, event event.Event) {
	fmt.Printf("Channel: %s; Topic: %s; DataEvent: %v\n", ch, event.Topic(), event.Data())
}

func main() {
	eb := event.NewEventBus()

	ch1 := event.NewEventChannel()
	ch2 := event.NewEventChannel()

	eb.Subscribe(Topic1, ch1)
	eb.Subscribe(Topic2, ch2)

	publisTo := func(topic event.Topic, data string) {
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
