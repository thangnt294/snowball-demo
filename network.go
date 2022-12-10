package main

import (
	"fmt"
	"strconv"
)

type Node struct {
	Id               int            // ID of the node
	Addrs            []chan NodeMsg // List of all node addresses in the chain
	ApiChan          chan ApiMsg    // Channel to listen to API call
	Chain            []int          // Local chain of the node
	ValidTxThreshold int            // Threshold to decide if a transaction is valid or not
}

type NodeMsg struct {
	nodeId    int
	tx        int
	responseC chan bool
}

// buildNetwork builds the test network with n nodes.
func buildNetwork(n int) []Node {
	addrs := make([]chan NodeMsg, 0)
	network := make([]Node, 0)

	// make a channel to receive msg for each node in the network
	for i := 0; i < n; i++ {
		addrs = append(addrs, make(chan NodeMsg))
	}

	// spin up the nodes
	for i := 0; i < n; i++ {
		var node Node
		if i > 0 && i%5 == 0 { // every 5 nodes, spin up a malicious node
			// malicious node
			node = Node{
				Id:               i,
				Addrs:            addrs,
				ApiChan:          make(chan ApiMsg),
				Chain:            make([]int, 0),
				ValidTxThreshold: 20,
			}
		} else {
			// valid node
			node = Node{
				Id:               i,
				Addrs:            addrs,
				ApiChan:          make(chan ApiMsg),
				Chain:            make([]int, 0),
				ValidTxThreshold: 100,
			}
		}
		network = append(network, node)
		go node.start()
	}

	return network
}

// start starts the node, listening for incoming messages and API requests.
func (node *Node) start() {
	fmt.Printf("Node %d up and running\n", node.Id)

	// Algorithm-related variables
	sampleSize := 10       // Number of nodes to ask each time
	quorumSize := 7        // Number of nodes required to have the same answers for consensus
	decisionThreshold := 5 // Number of consecutive successes needed to arrive to a decision

	decisionMap := make(map[int]int)              // map of already-decided transactions
	decisionReplyChan := make(chan DecisionReply) // channel to receive final decisions
	respC := make(chan bool, sampleSize)          // channel to receive answers from other nodes

	for {
		select {
		case ApiMsg := <-node.ApiChan:
			switch ApiMsg.Type {
			case 1: // create Tx
				tx, _ := strconv.Atoi(ApiMsg.Data)
				var pref bool
				if tx >= node.ValidTxThreshold {
					pref = true
				} else {
					pref = false
				}
				if decisionMap[tx] == DECIDED_FALSE {
					fmt.Println("False transaction. Will not propagate.")
				} else if decisionMap[tx] == WAITING_FOR_DECISION {
					fmt.Println("Validating transaction. Will not propagate.")
				} else {
					fmt.Println("New transaction. Asking around for decision...")
					decisionMap[tx] = WAITING_FOR_DECISION
					txValidation := TxValidation{
						Addrs:             node.Addrs,
						Resp:              respC,
						Tx:                tx,
						SampleSize:        sampleSize,
						QuorumSize:        quorumSize,
						DecisionThreshold: decisionThreshold,
						NodeId:            node.Id,
						Pref:              pref,
					}
					go txValidation.decideOnTx(decisionReplyChan)
				}

			case 2: // list chain
				fmt.Printf("Node %d: %v\n", node.Id, node.Chain)
			}

		case msg := <-node.Addrs[node.Id]: // receive questions from other nodes
			currentDecision := decisionMap[msg.tx]
			if currentDecision != NEW && currentDecision != WAITING_FOR_DECISION { // already decided
				var decision bool
				if currentDecision == DECIDED_TRUE {
					decision = true
				}
				msg.responseC <- decision
			} else {
				// build an initial preference
				var pref bool
				if msg.tx >= node.ValidTxThreshold {
					pref = true
				} else {
					pref = false
				}
				msg.responseC <- pref

				if currentDecision == NEW {
					decisionMap[msg.tx] = WAITING_FOR_DECISION
					txValidation := TxValidation{
						Addrs:             node.Addrs,
						Resp:              respC,
						Tx:                msg.tx,
						SampleSize:        sampleSize,
						QuorumSize:        quorumSize,
						DecisionThreshold: decisionThreshold,
						NodeId:            node.Id,
						Pref:              pref,
					}
					go txValidation.decideOnTx(decisionReplyChan)
				}
			}

		case reply := <-decisionReplyChan: // finally came to a decision on a tx
			decision := DECIDED_FALSE
			if reply.Decision {
				decision = DECIDED_TRUE
			}
			decisionMap[reply.Tx] = decision
			if reply.Decision {
				node.Chain = append(node.Chain, reply.Tx)
			}
		}
	}
}
