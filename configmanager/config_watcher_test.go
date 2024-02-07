package configmanager

import (
	"testing"
)

func TestConfigWatcherListen(t *testing.T) {
	type testConfig struct {
		value string
	}

	ch := make(chan Config)
	cfg := testConfig{value: "value1"}

	var testVal string

	w := NewConfigWatcher(cfg, ch)
	w.Listen(func(cfg testConfig) error {
		testVal = cfg.value
		return nil
	})

	ch <- testConfig{value: "value2"}
	close(ch)

	if w.Config().value != "value2" {
		t.Errorf("unexpected data: %v", w.Config())
	}
	if testVal != "value2" {
		t.Errorf("unexpected data: %v", testVal)
	}
}
