package configmanager

import (
	"testing"
)

func TestConfigWatcherListen(t *testing.T) {
	type testConfig struct {
		value string
	}

	t.Run("ListenWithCallback", func(t *testing.T) {
		var testVal string
		cfg := testConfig{value: "value1"}
		ch := make(chan Config)

		w := NewConfigWatcher(cfg, ch)
		w.Listen(func(cfg testConfig) error {
			testVal = cfg.value
			return nil
		})

		ch <- testConfig{value: "value2"}
		close(ch)

		if testVal != "value2" {
			t.Errorf("unexpected data: %v", testVal)
		}
	})

	t.Run("ListenWithoutCallback", func(t *testing.T) {
		ch := make(chan Config)
		cfg := testConfig{value: "value1"}

		w := NewConfigWatcher(cfg, ch)
		w.Listen(nil)

		ch <- testConfig{value: "value2"}
		close(ch)

		if w.Config().value != "value2" {
			t.Errorf("unexpected data: %v", w.Config())
		}
	})
}
