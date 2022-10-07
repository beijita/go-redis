package main

import (
	"fmt"
	"go-redis/config"
	"go-redis/transport/redis/server"
	"go-redis/transport/tcp"
	"os"
)

var banner = `
   ______          ___
  / ____/___  ____/ (_)____
 / / __/ __ \/ __  / / ___/
/ /_/ / /_/ / /_/ / (__  )
\____/\____/\__,_/_/____/
`

func main() {
	fmt.Println("hello go-redis")
	print(banner)

	configFilename := os.Getenv("CONFIG")
	if configFilename != "" {
		config.SetupConfig(configFilename)
	} else if fileExists("redis.config") {
		config.SetupConfig("redis.config")
	} else {
		config.Properties = defaultProperties
	}
	address := fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port)
	tcp.ListenAndServeWithSignal(&tcp.Config{Address: address}, server.NewHandler())
}

var defaultProperties = &config.ServerProperties{
	Bind:           "127.0.0.1",
	Port:           6379,
	AppendOnly:     false,
	AppendFilename: "",
	MaxClients:     128,
	RequirePass:    "",
	DataBases:      0,
	RDBFilename:    "",
	Peers:          nil,
	Self:           "",
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}
