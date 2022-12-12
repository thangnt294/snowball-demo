package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	NetworkConf  NetworkConf  `toml:"network"`
	SnowballConf SnowballConf `toml:"snowball"`
}

type NetworkConf struct {
	Size      int `toml:"size"`
	Neighbors int `toml:"neighbors"`
	StartPort int `toml:"startPort"`
	WebPort   int `toml:"webPort"`
}

type SnowballConf struct {
	SampleSize        int `toml:"sampleSize"`
	QuorumSize        int `toml:"quorumSize"`
	DecisionThreshold int `toml:"decisionThreshold"`
}

var GlobalConfig Config

func ReadConfig(configFile string, v interface{}) {
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
