package web

import (
	"ava/config"
	"ava/network"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) { // test setup
	config.ReadConfig("../ava.conf", &config.GlobalConfig)
	code := m.Run()
	os.Exit(code)
}

func TestWebServer(t *testing.T) {
	testNetwork := network.BuildNetwork(config.GlobalConfig.NetworkConf.Size)
	go StartWebServer(testNetwork)

	res, err := http.Get(fmt.Sprintf("http://localhost:%d/chains", config.GlobalConfig.NetworkConf.WebPort))
	assert.NoError(t, err)

	resp, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "OK", string(resp))

	res, err = http.Get(fmt.Sprintf("http://localhost:%d/chain/%d", config.GlobalConfig.NetworkConf.WebPort, config.GlobalConfig.NetworkConf.StartPort))
	assert.NoError(t, err)

	resp, err = io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "OK", string(resp))

	res, err = http.Get(fmt.Sprintf("http://localhost:%d/neighbors/%d", config.GlobalConfig.NetworkConf.WebPort, config.GlobalConfig.NetworkConf.StartPort))
	assert.NoError(t, err)

	resp, err = io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "OK", string(resp))

	res, err = http.Post(fmt.Sprintf("http://localhost:%d/createTx/1", config.GlobalConfig.NetworkConf.WebPort), "text/plain; charset=utf-8", nil)
	assert.NoError(t, err)

	resp, err = io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "OK", string(resp))
}
