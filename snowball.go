package main

import "math/rand"

type DecisionReply struct {
	Decision bool
	Tx       int
}

type TxValidation struct {
	Addrs             []chan NodeMsg // Addresses of all nodes in the network to ask
	Resp              chan bool      // Channel to receive responses
	Tx                int            // Transaction to be validated
	SampleSize        int            // Number of nodes to ask
	QuorumSize        int            // Minimum number of same responses needed to adopt a new preference
	DecisionThreshold int            // Number of consecutive quorum needed to make a final decision
	NodeId            int            // ID of the node trying to validate the tx
	Pref              bool           // Current preference
}

const (
	NEW = iota
	WAITING_FOR_DECISION
	DECIDED_FALSE
	DECIDED_TRUE
)

// decideOnTx decides if the Tx is valid or not, then sends the decision back to replyChan.
func (txValidation *TxValidation) decideOnTx(replyChan chan DecisionReply) {
	decision := txValidation.snowBall()
	replyChan <- DecisionReply{Decision: decision, Tx: txValidation.Tx}
}

// snowBall implements the snowball algorithm.
func (txValidation *TxValidation) snowBall() bool {
	var decision bool
	myMsg := NodeMsg{txValidation.NodeId, txValidation.Tx, txValidation.Resp}

	decided := false
	consecutiveSuccesses := 0
	for !decided {
		// ask random nodes
		nodesToAsk := randomNodesToAsk(txValidation.NodeId, len(txValidation.Addrs), txValidation.SampleSize)
		for _, id := range nodesToAsk {
			txValidation.Addrs[id] <- myMsg
		}

		// collect responses
		countT := 0
		countF := 0
		for i := 0; i < txValidation.SampleSize; i++ {
			ans := <-txValidation.Resp
			if ans {
				countT++
			} else {
				countF++
			}
		}

		if countT >= txValidation.QuorumSize {
			newPref := true
			if newPref == txValidation.Pref {
				consecutiveSuccesses++
			} else {
				consecutiveSuccesses = 1
			}
			txValidation.Pref = newPref
		} else if countF >= txValidation.QuorumSize {
			newPref := false
			if newPref == txValidation.Pref {
				consecutiveSuccesses++
			} else {
				consecutiveSuccesses = 1
			}
			txValidation.Pref = newPref
		} else {
			consecutiveSuccesses = 0
		}

		if consecutiveSuccesses >= txValidation.DecisionThreshold { // decided
			decided = true
			decision = txValidation.Pref
		}
	}

	return decision
}

// randomNodesToAsk randomizes and returns a list of nodes in the network to ask.
func randomNodesToAsk(nodeId, n, sampleSize int) []int {
	nodesToAsk := make([]int, 0) // pool of k nodes to ask

	nodeSet := make(map[int]struct{})
	for i := 0; i < sampleSize; i++ {
		r := rand.Intn(n)
		for _, ok := nodeSet[r]; ok || r == nodeId; _, ok = nodeSet[r] { // retrying until getting a new node
			r = rand.Intn(n)
		}
		nodeSet[r] = struct{}{}
	}
	for node := range nodeSet {
		nodesToAsk = append(nodesToAsk, node)
	}

	return nodesToAsk
}
