package main

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebServer(t *testing.T) {
	network := buildNetwork(10)
	go startWebServer(network)

	res, err := http.Get("http://localhost:3000/chain")
	assert.NoError(t, err)

	resp, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "OK", string(resp))

	res, err = http.Post("http://localhost:3000/createTx/1", "text/plain; charset=utf-8", nil)
	assert.NoError(t, err)

	resp, err = io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "OK", string(resp))
}
