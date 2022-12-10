package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildNetwork(t *testing.T) {
	n := 200
	network := buildNetwork(n)

	assert.Equal(t, n, len(network))
}
