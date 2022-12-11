package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	NetworkConf NetworkConf `toml:"network"`
	NodeConf    NodeConf    `toml:"node"`
}

type NetworkConf struct {
	Nodes     int `toml:"nodes"`
	Neighbors int `toml:"neighbors"`
}

type NodeConf struct {
	SampleSize        int `toml:"sampleSize"`
	QuorumSize        int `toml:"quorumSize"`
	DecisionThreshold int `toml:"decisionThreshold"`
}

func readConfig(configFile string, v interface{}) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println("Error: Cannot read config", err)
		os.Exit(1)
	}
	_, err = toml.Decode(string(file), v)
	if err != nil {
		fmt.Println("Error: Cannot read config", err)
		os.Exit(1)
	}
}
