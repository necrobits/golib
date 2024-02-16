package configmanager

type ConfigWatcher[T Config] struct {
	ch  chan Config
	cfg T
}

func NewConfigWatcher[T Config](cfg T, ch chan Config) *ConfigWatcher[T] {
	return &ConfigWatcher[T]{
		ch:  ch,
		cfg: cfg,
	}
}

func (w *ConfigWatcher[T]) Config() T {
	return w.cfg
}

func (w *ConfigWatcher[T]) Listen(callbackFn func(T) error) {
	go func() {
		for cfg := range w.ch {
			castedCfg := cfg.(T)
			w.cfg = castedCfg
			if callbackFn != nil {
				callbackFn(castedCfg)
			}
		}	
	}()
}
