package main

import (
	"fmt"

	"github.com/necrobits/x/configmanager"
)

type ManagerADeps struct {
	Name           SystemName
	DatabaseConfig DatabaseConfigMap
	ServerConfig   ServerConfig
	CfgManager     *configmanager.Manager
}

type ManagerA struct {
	name           SystemName
	databaseConfig DatabaseConfigMap
	serverConfig   ServerConfig
}

func NewManagerA(deps *ManagerADeps) *ManagerA {
	cfgMng := deps.CfgManager

	dbCfgCh := cfgMng.Register(DatabaseConfigMap{})
	srvCfgCh := cfgMng.Register(ServerConfig{})
	systemNameCh := cfgMng.Register(SystemName(""))

	m := &ManagerA{
		name:           deps.Name,
		databaseConfig: deps.DatabaseConfig,
		serverConfig:   deps.ServerConfig,
	}

	go func() {
		for {
			select {
			case event := <-dbCfgCh:
				m.databaseConfig = event.Data().(DatabaseConfigMap)
			case cfg := <-srvCfgCh:
				m.serverConfig = cfg.Data().(ServerConfig)
			case cfg := <-systemNameCh:
				m.name = cfg.Data().(SystemName)
			}
		}
	}()

	return m
}

func (m *ManagerA) PrintConfig() {
	fmt.Printf("Manager A System name: %s\n", m.name)
	if m.databaseConfig != nil {
		fmt.Printf("Manager A System database 1:\n\t\tHost: %s\n\t\tPort: %d\n\t\tName: %s\n", m.databaseConfig["db1"].Host, m.databaseConfig["db1"].Port, m.databaseConfig["db1"].Name)
		fmt.Printf("Manager A System database 2:\n\t\tHost: %s\n\t\tPort: %d\n\t\tName: %s\n", m.databaseConfig["db2"].Host, m.databaseConfig["db2"].Port, m.databaseConfig["db2"].Name)
	}
	fmt.Printf("Manager A Server config:\n\tHost: %s\n\tPort: %d\n", m.serverConfig.Host, m.serverConfig.Port)
}
