package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomNodesToask(t *testing.T) {
	nodeId := 1
	n := 20
	sampleSize := 10
	nodesToAsk := randomNodesToAsk(nodeId, n, sampleSize)
	assert.Equal(t, sampleSize, len(nodesToAsk))
	assert.NotContains(t, nodesToAsk, nodeId) // obviously shouldn't ask yourself
}

func TestSnowball(t *testing.T) {
	network := buildNetwork(10)
	testNode := network[0]
	sampleSize := 4        // Number of nodes to ask each time
	quorumSize := 3        // Number of nodes required to have the same answers for consensus
	decisionThreshold := 3 // Number of consecutive successes needed to arrive to a decision

	respC := make(chan bool, sampleSize) // channel to receive answers from other nodes

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
			Addrs:             testNode.Addrs,
			Resp:              respC,
			Tx:                testCase.Tx,
			SampleSize:        sampleSize,
			QuorumSize:        quorumSize,
			DecisionThreshold: decisionThreshold,
			NodeId:            testNode.Id,
			Pref:              pref,
		}
		decision := txValidation.snowBall()
		assert.Equal(t, decision, testCase.Valid)
	}
}
