package main

import (
	"ava/config"
	"ava/network"
	"ava/web"
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

func main() {
	config.ReadConfig("ava.conf", &config.GlobalConfig)
	fmt.Printf("Read the following config: %#v\n", config.GlobalConfig)
	network := network.BuildNetwork(config.GlobalConfig.NetworkConf.Size)
	web.StartWebServer(network)
}
