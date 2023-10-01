package configmanager

import "github.com/necrobits/x/eventbus"

type Topic = eventbus.Topic

type ValidatableConfig interface {
	Validate() error
}

type RegistrableConfig interface {
	Topic() Topic
}
