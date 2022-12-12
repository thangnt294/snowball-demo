package network

import (
	"ava/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomNodesToask(t *testing.T) {
	sampleSize := 5
	neighbors := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	nodesToAsk := randomNodesToAsk(sampleSize, neighbors)
	assert.Equal(t, sampleSize, len(nodesToAsk))
}

func TestSnowball(t *testing.T) {
	network := BuildNetwork(config.GlobalConfig.NetworkConf.Size)
	testNode := network[0]

	type TestCase struct {
		Data     int
		Expected bool
	}
	var testCases = []TestCase{
		{Data: 5, Expected: false},  // invalid for all nodes
		{Data: 22, Expected: false}, // valid for malicious nodes
		{Data: 101, Expected: true}, // valid for all nodes
	}

	for _, tc := range testCases {
		var pref bool
		if tc.Data >= testNode.ValidTxThreshold {
			pref = true
		}
		txValidation := TxValidation{
			Node: &testNode,
			Tx:   tc.Data,
			Pref: pref,
		}
		decision := txValidation.snowBall()
		assert.Equal(t, tc.Expected, decision)
	}
}
