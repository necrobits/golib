package old_configmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager(t *testing.T) {
	type TestConfig struct {
		Host string `cfg:"host"`
		Port int    `cfg:"port"`
	}

	cfg := &TestConfig{
		Host: "localhost",
		Port: 8080,
	}
	m := NewManager(&NewManagerOpts{
		Cfg:    cfg,
		CfgKey: "test",
	})

	assert.Equal(t, "test", m.ConfigKey())
	assert.Equal(t, cfg, m.Config())
}

func TestRegisterNode(t *testing.T) {
	type TestConfig struct {
		Host string `cfg:"host"`
		Port int    `cfg:"port"`
	}

	cfg := &TestConfig{
		Host: "localhost",
		Port: 8080,
	}
	m := NewManager(&NewManagerOpts{
		Cfg:    cfg,
		CfgKey: "test",
	})

	m.RegisterNode("child", NewManager(&NewManagerOpts{
		Cfg:    cfg,
		CfgKey: "child",
	}))

	assert.Equal(t, "test", m.ConfigKey())
	assert.Equal(t, cfg, m.Config())
	assert.Equal(t, "child", m.nodes["child"].ConfigKey())
	assert.Equal(t, cfg, m.nodes["child"].Config())
}
