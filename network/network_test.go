package network

import (
	"os"
	"testing"

	"ava/config"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) { // test setup
	config.ReadConfig("../ava.conf", &config.GlobalConfig)
	code := m.Run()
	os.Exit(code)
}

func TestBuildNetwork(t *testing.T) {
	n := 200
	network := BuildNetwork(n)

	assert.Equal(t, n, len(network))
}

func TestBuildNodeConn(t *testing.T) {
	config.GlobalConfig.NetworkConf.Size = 10
	config.GlobalConfig.NetworkConf.Neighbors = 4
	nodeConn := buildNodeConn(config.GlobalConfig.NetworkConf.Size)

	assert.Equal(t, config.GlobalConfig.NetworkConf.Size, len(nodeConn))
	for _, conn := range nodeConn {
		assert.GreaterOrEqual(t, len(conn), config.GlobalConfig.NetworkConf.Neighbors)
	}
}
