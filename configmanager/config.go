package configmanager

import "github.com/necrobits/x/event"

type Topic = event.Topic

type ValidatableConfig interface {
	Validate() error
}

type RegistrableConfig interface {
	Topic() Topic
}
