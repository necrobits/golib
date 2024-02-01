package old_configmanager

const TagKey = "cfg"

type Config interface{}

type ConfigValidateFunc func(cfg Config) error
