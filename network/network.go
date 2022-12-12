package network

import (
	"ava/config"
	"fmt"
	"math/rand"
)

// BuildNetwork builds the test network with n nodes.
func BuildNetwork(n int) []Node {
	network := make([]Node, 0)

	// build a map of node connections first
	nodeConn := buildNodeConn(n)

	// spin up the nodes
	for i := 0; i < n; i++ {
		var node Node
		nodeAddr := config.GlobalConfig.NetworkConf.StartPort + i

		if i > 0 && i%5 == 0 { // every 5 nodes, spin up a malicious node
			// malicious node
			node = Node{
				Addr:              nodeAddr,
				Neighbors:         nodeConn[nodeAddr],
				Chain:             make([]int, 0),
				ValidTxThreshold:  20, // malicious node trying to approve tx bigger than or equal to 20
				SampleSize:        config.GlobalConfig.SnowballConf.SampleSize,
				QuorumSize:        config.GlobalConfig.SnowballConf.QuorumSize,
				DecisionThreshold: config.GlobalConfig.SnowballConf.DecisionThreshold,
				DecisionChan:      make(chan DecisionOnTx),
			}
		} else {
			// valid node
			node = Node{
				Addr:              nodeAddr,
				Neighbors:         nodeConn[nodeAddr],
				Chain:             make([]int, 0),
				ValidTxThreshold:  100, // valid node only approve tx bigger than or equal to 100
				SampleSize:        config.GlobalConfig.SnowballConf.SampleSize,
				QuorumSize:        config.GlobalConfig.SnowballConf.QuorumSize,
				DecisionThreshold: config.GlobalConfig.SnowballConf.DecisionThreshold,
				DecisionChan:      make(chan DecisionOnTx),
			}
		}
		network = append(network, node)
		go node.start()
	}

	return network
}

// buildNodeConn builds and returns a map of node connections for n nodes.
func buildNodeConn(n int) map[int][]int {
	// each node will connect to at least [neighbors] node. The number of minimum neighbors is written in the config file.
	// there are nodes that might connect to more nodes than necessary.
	nodeConn := make(map[int][]int)
	for i := 0; i < n; i++ {
		nodeAddr := config.GlobalConfig.NetworkConf.StartPort + i

		if _, ok := nodeConn[nodeAddr]; !ok {
			nodeConn[nodeAddr] = make([]int, 0)
		}

		exisingNeighborSet := make(map[int]struct{})
		for _, neighbor := range nodeConn[nodeAddr] { // check already-formed neighbors
			exisingNeighborSet[neighbor] = struct{}{}
		}

		numberOfNeighborToAdd := config.GlobalConfig.NetworkConf.Neighbors - len(exisingNeighborSet)
		if numberOfNeighborToAdd <= 0 { // enough neighbors
			continue
		}

		newNeighbors := make([]int, numberOfNeighborToAdd)
		for k := 0; k < numberOfNeighborToAdd; k++ {
			neighbor := config.GlobalConfig.NetworkConf.StartPort + rand.Intn(n)
			for _, ok := exisingNeighborSet[neighbor]; ok || neighbor == nodeAddr; _, ok = exisingNeighborSet[neighbor] { // retry until getting a new node
				neighbor = config.GlobalConfig.NetworkConf.StartPort + rand.Intn(n)
			}
			exisingNeighborSet[neighbor] = struct{}{} // add to a set to avoid duplication
			newNeighbors[k] = neighbor
		}

		for _, neighbor := range newNeighbors { // form connection between 2 nodes
			nodeConn[nodeAddr] = append(nodeConn[nodeAddr], neighbor)

			if _, ok := nodeConn[neighbor]; !ok {
				nodeConn[neighbor] = make([]int, 0)
			}
			nodeConn[neighbor] = append(nodeConn[neighbor], nodeAddr)
		}

	}

	for k, v := range nodeConn {
		if len(v) == 0 {
			fmt.Println("Conn missing:", k, v)
		}
	}
	return nodeConn
}
