package main

import (
	"fmt"
	"math/rand"
	"time"
)

var config Config

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

func main() {
	readConfig("ava.conf", &config)
	fmt.Printf("Read the following config: %#v\n", config)
	network := buildNetwork(config.NetworkConf.Nodes)
	startWebServer(network)
}
