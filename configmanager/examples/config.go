package main

import (
	"fmt"
	"math/rand"

	"github.com/necrobits/x/configmanager"
)

type Topic = configmanager.Topic

type Config struct {
	Server ServerConfig `cfg:"server"`
	System SystemConfig `cfg:"system"`
}

type ServerConfig struct {
	Host string `cfg:"host"`
	Port int    `cfg:"port"`
}

func (c ServerConfig) Topic() Topic {
	return "server"
}

type SystemConfig struct {
	Name        SystemName        `cfg:"name"`
	Databasse   DatabaseConfigMap `cfg:"database"`
	SupportedOS OSConfigs         `cfg:"supported_os"`
}

func (c SystemConfig) Topic() Topic {
	return "system"
}

type SystemName string

func (c SystemName) Topic() Topic {
	return "system:name"
}

type DatabaseConfig struct {
	Host string `cfg:"host"`
	Port int    `cfg:"port"`
	Name string `cfg:"name"`
}

func (c DatabaseConfig) Topic() Topic {
	return configmanager.Topic(fmt.Sprintf("database:%s", c.Name))
}

func (c *DatabaseConfig) Validate() error {
	if c.Port%2 == 0 {
		return nil
	}
	return fmt.Errorf("invalid port %d", c.Port)
}

type DatabaseConfigMap map[string]DatabaseConfig

func (c DatabaseConfigMap) Topic() Topic {
	return "databases"
}

type OSConfig struct {
	Name    string `cfg:"name"`
	Version string `cfg:"version"`
}

func (c OSConfig) Topic() Topic {
	return configmanager.Topic(fmt.Sprintf("os:%s", c.Name))
}

type OSConfigs []OSConfig

func (c OSConfigs) Topic() Topic {
	return "supported_os"
}

func (c OSConfigs) Validate() error {
	r := rand.Intn(2)
	if r == 0 {
		return fmt.Errorf("invalid OS config")
	}
	return nil
}
