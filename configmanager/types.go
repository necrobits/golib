package configmanager

type Config interface{}

type ValidateFunc func(cfg Config) error
