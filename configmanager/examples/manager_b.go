package main

import (
	"fmt"

	"github.com/necrobits/x/configmanager"
)

type ManagerBDeps struct {
	ServerConfig ServerConfig
	SystemConfig SystemConfig
	CfgManager   *configmanager.Manager
}

type ManagerB struct {
	serverConfig ServerConfig
	systemConfig SystemConfig
}

func NewManagerB(deps *ManagerBDeps) *ManagerB {
	cfgMng := deps.CfgManager

	serverCfgCh := cfgMng.Register(ServerConfig{})
	systemCfgCh := cfgMng.Register(SystemConfig{})

	m := &ManagerB{
		serverConfig: deps.ServerConfig,
		systemConfig: deps.SystemConfig,
	}

	go func() {
		for {
			select {
			case event := <-serverCfgCh:
				m.serverConfig = event.Data().(ServerConfig)
			case event := <-systemCfgCh:
				m.systemConfig = event.Data().(SystemConfig)
			}
		}
	}()

	return m
}

func (m *ManagerB) PrintConfig() {
	fmt.Printf("Manager B Server config:\n\tHost: %s\n\tPort: %d\n", m.serverConfig.Host, m.serverConfig.Port)
	fmt.Printf("Manager B System config:\n\tName: %s\n", m.systemConfig.Name)
	if len(m.systemConfig.SupportedOS) != 0 {
		for i, os := range m.systemConfig.SupportedOS {
			fmt.Printf("Manager B System config:\nSupportedOS %d: %+v\n", i, os)
		}
	}
	fmt.Printf("Manager B System config:\n\tDatabase 1:\n\t\tHost: %s\n\t\tPort: %d\n\t\tName: %s\n", m.systemConfig.Databasse["db1"].Host, m.systemConfig.Databasse["db1"].Port, m.systemConfig.Databasse["db1"].Name)
	fmt.Printf("Manager B System config:\n\tDatabase 2:\n\t\tHost: %s\n\t\tPort: %d\n\t\tName: %s\n", m.systemConfig.Databasse["db2"].Host, m.systemConfig.Databasse["db2"].Port, m.systemConfig.Databasse["db2"].Name)
}
