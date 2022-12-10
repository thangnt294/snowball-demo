package main

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))
}

func main() {
	network := buildNetwork(200)
	startWebServer(network)
}
