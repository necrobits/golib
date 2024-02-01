package old_configmanager

import (
	"github.com/necrobits/x/event"
	"github.com/necrobits/x/kvstore"
)

type Manager struct {
	store      kvstore.KvStore
	parent     *Manager
	cfg        Config
	cfgKey     string
	nodes      map[string]*Manager
	validators map[string]ConfigValidateFunc
	eb         *event.EventBus
}

type NewManagerOpts struct {
	Store  kvstore.KvStore
	Parent *Manager
	Cfg    Config
	CfgKey string
}

func NewManager(opts *NewManagerOpts) *Manager {
	return &Manager{
		store:  opts.Store,
		parent: opts.Parent,
		cfg:    opts.Cfg,
		cfgKey: opts.CfgKey,
		nodes:  make(map[string]*Manager),
		eb:     event.NewEventBus(),
	}
}

func (m *Manager) ConfigKey() string {
	return m.cfgKey
}

func (m *Manager) Config() Config {
	return m.cfg
}

func (m *Manager) RegisterNode(key string, node *Manager) {
	node.parent = m
	m.nodes[key] = node
}

func (m *Manager) RegisterValidator(key string, validator ConfigValidateFunc) {
	m.validators[key] = validator
}

func (m *Manager) SubscribeConfig(key string) event.EventChannel {
	cfgCh := event.NewEventChannel()
	m.eb.Subscribe(event.Topic(key), cfgCh)
	return cfgCh
}
