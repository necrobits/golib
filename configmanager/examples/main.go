package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/necrobits/x/configmanager"
	"github.com/necrobits/x/kvstore"
)

func main() {
	cfg := Config{
		System: SystemConfig{
			Name: "Test System",
			Databasse: DatabaseConfigMap{
				"db1": {
					Host: "localhost",
					Port: 3306,
					Name: "db1",
				},
				"db2": {
					Host: "localhost2",
					Port: 3308,
					Name: "db2",
				},
			},
			SupportedOS: OSConfigs{
				{
					Name:    "linux",
					Version: "li1",
				},
				{
					Name:    "windows",
					Version: "10",
				},
			},
		},
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	dottedCfg := configmanager.FlatConfig(cfg, "cfg")
	data := make(map[string]kvstore.Data)
	for k, v := range dottedCfg {
		data[k] = v
	}

	memStore := &MemStore{
		data: data,
	}

	cfgMng, err := configmanager.NewManager(&configmanager.ManagerOpts{
		Store:   memStore,
		RootCfg: cfg,
		TagKey:  "cfg",
	})
	if err != nil {
		panic(err)
	}
	mngA := NewManagerA(&ManagerADeps{
		Name:           cfg.System.Name,
		DatabaseConfig: cfg.System.Databasse,
		ServerConfig:   cfg.Server,
		CfgManager:     cfgMng,
	})
	mngB := NewManagerB(&ManagerBDeps{
		ServerConfig: cfg.Server,
		SystemConfig: cfg.System,
		CfgManager:   cfgMng,
	})

	printAllConfigs := func(n int) {
		fmt.Printf("Config %d:\n", n)
		mngA.PrintConfig()
		mngB.PrintConfig()
		fmt.Println()
	}

	printMemStore := func(n int) {
		fmt.Printf("\nMemstore (%d):\n", n)
		for k, v := range memStore.data {
			fmt.Printf("%s: %+v\n", k, v)
		}
	}

	updateSystemName := func() {
		fmt.Println("\nUpdating system name")
		err := cfgMng.Update(map[string]interface{}{
			"system.name": SystemName(fmt.Sprintf("New system %d", rand.Intn(100))),
		})
		if err != nil {
			fmt.Println("Error updating system name: ", err)
		} else {
			fmt.Println("Updated system name")
		}
	}
	updateOSConfig := func() {
		fmt.Println("\nUpdating OS config")
		err := cfgMng.Update(map[string]interface{}{
			// "system.supported_os.1.name":    "new windows",
			// "system.supported_os.1.version": "11",
			// "system.supported_os.2": OSConfig{
			// 	Name:    "mac",
			// 	Version: "10.1",
			// },
			"system.supported_os": OSConfigs{
				{
					Name:    "linux",
					Version: "kernel x",
				},
				{
					Name:    "windows",
					Version: "10",
				},
			},
		})
		if err != nil {
			fmt.Println("Error updating OS config: ", err)
		} else {
			fmt.Println("Updated OS config")
		}
	}
	updateDb1Config := func() {
		port := rand.Intn(1000) + 3000
		fmt.Println("\nUpdating database 1 config with port ", port)
		err := cfgMng.Update(map[string]interface{}{
			"system.database.db1.port": port,
			"system.database.db1.name": "newdb1",
		})
		if err != nil {
			fmt.Println("Error updating database 1 config: ", err)
		} else {
			fmt.Println("Updated database 1 config")
		}
	}
	updateDb2Config := func() {
		port := rand.Intn(1000) + 3000
		fmt.Println("\nUpdating database 2 config with port ", port)
		err := cfgMng.Update(map[string]interface{}{
			"system.database.db2": DatabaseConfig{
				Host: "localhostdb2",
				Port: port,
				Name: "Db2",
			},
		})
		if err != nil {
			fmt.Println("Error updating database 2 config: ", err)
		} else {
			fmt.Println("Updated database 2 config")
		}
	}
	updateServerConfig := func() {
		port := rand.Intn(1000) + 8000
		testInt := rand.Intn(1000)
		fmt.Println("\nUpdating server config with port ", port)
		err := cfgMng.Update(map[string]interface{}{
			"server.host":      "newserverhost",
			"server.port":      port,
			"server.test.name": "newtest" + strconv.Itoa(testInt),
		})
		if err != nil {
			fmt.Println("Error updating server config: ", err)
		} else {
			fmt.Println("Updated server config")
		}
	}

	go func() {
		for {
			time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)
			cfgToChange := rand.Intn(5)
			switch cfgToChange {
			case 0:
				updateServerConfig()
			case 1:
				updateOSConfig()
			case 2:
				updateDb1Config()
			case 3:
				updateDb2Config()
			case 4:
				updateSystemName()
			}
		}
	}()

	go func() {
		i := 0
		for {
			time.Sleep(time.Duration(1000+rand.Intn(1000)) * time.Millisecond)
			printAllConfigs(i)
			printMemStore(i)
			i++
		}
	}()

	time.Sleep(5 * time.Second)
}
