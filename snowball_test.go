package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomNodesToask(t *testing.T) {
	sampleSize := 5
	neighbors := []int{2, 3, 4, 5, 6, 7, 8, 9, 10}
	nodesToAsk := randomNodesToAsk(sampleSize, neighbors)
	assert.Equal(t, sampleSize, len(nodesToAsk))
}

// TODO: fix test
func TestSnowball(t *testing.T) {
	readConfig("ava.conf", &config)
	network := buildNetwork(config.NetworkConf.Nodes)
	assert.Equal(t, config.NetworkConf.Nodes, len(network))
	testNode := network[0]

	type TestCase struct {
		Tx    int
		Valid bool
	}
	var testCases = []TestCase{
		{Tx: 5, Valid: false},  // invalid for all nodes
		{Tx: 22, Valid: false}, // valid for malicious nodes
		{Tx: 101, Valid: true}, // valid for all nodes
	}

	for _, testCase := range testCases {
		var pref bool
		if testCase.Tx >= testNode.ValidTxThreshold {
			pref = true
		}
		txValidation := TxValidation{
			Node: &testNode,
			Tx:   testCase.Tx,
			Pref: pref,
		}
		decision := txValidation.snowBall()
		assert.Equal(t, testCase.Valid, decision)
	}
}
