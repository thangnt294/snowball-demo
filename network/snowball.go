package network

import (
	"ava/api"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

type TxValidation struct {
	Node *Node // The node trying to validate tx
	Tx   int   // Transaction to be validated
	Pref bool  // Current preference
}

type DecisionOnTx struct {
	Decision int
	Tx       int
}

const (
	DECISION_NEW = iota
	DECISION_WAITING
	DECISION_INVALID
	DECISION_VALID
)

// decideOnTx decides if the Tx is valid or not, then sends the decision back to decisionChan.
func (txValidation *TxValidation) DecideOnTx() {
	var decision int
	isValid := txValidation.snowBall()
	if isValid {
		decision = DECISION_VALID
	} else {
		decision = DECISION_INVALID
	}
	txValidation.Node.DecisionChan <- DecisionOnTx{Decision: decision, Tx: txValidation.Tx}
}

// snowBall implements the snowball algorithm.
func (txValidation *TxValidation) snowBall() bool {
	var decision bool

	decided := false
	consecutiveSuccesses := 0
	for !decided {
		resChan := make(chan bool, txValidation.Node.SampleSize)
		errChan := make(chan error, txValidation.Node.SampleSize)
		// ask random nodes
		nodesToAsk := randomNodesToAsk(txValidation.Node.SampleSize, txValidation.Node.Neighbors)
		for _, node := range nodesToAsk {
			go askToValidateTx(node, txValidation.Tx, resChan, errChan)
		}

		// collect responses
		countT := 0
		countF := 0
		for i := 0; i < txValidation.Node.SampleSize; i++ {
			select {
			case pref := <-resChan:
				if pref {
					countT++
				} else {
					countF++
				}
			case <-errChan: // Can log errors
			}
		}

		if countT >= txValidation.Node.QuorumSize {
			newPref := true
			if newPref == txValidation.Pref {
				consecutiveSuccesses++
			} else {
				consecutiveSuccesses = 1
			}
			txValidation.Pref = newPref
		} else if countF >= txValidation.Node.QuorumSize {
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

		if consecutiveSuccesses >= txValidation.Node.DecisionThreshold { // decided
			decided = true
			decision = txValidation.Pref
		}
	}

	return decision
}

// askToValidateTx sends a request to validate a tx to a node.
func askToValidateTx(nodeAddr int, tx int, resChan chan<- bool, errChan chan<- error) {
	myMsg := api.TxValidationRequest{Tx: tx}
	jsonBody, err := json.Marshal(myMsg)
	if err != nil {
		errChan <- err
	}
	url := fmt.Sprintf("http://localhost:%d/validate", nodeAddr)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		errChan <- err
	}
	var res api.TxValidationResponse
	err = api.ReadJSONResponse(resp, &res)
	if err != nil {
		errChan <- err
		return
	}

	resChan <- res.Pref
}

// randomNodesToAsk randomizes and returns a list of nodes from the neighbor list to ask.
func randomNodesToAsk(sampleSize int, neighbors []int) []int {
	nodesToAsk := make([]int, 0) // pool of k nodes to ask

	nodeSet := make(map[int]struct{})
	for i := 0; i < sampleSize; i++ {
		nodeToAsk := neighbors[rand.Intn(len(neighbors))]
		for _, ok := nodeSet[nodeToAsk]; ok; _, ok = nodeSet[nodeToAsk] { // retrying until getting a new node
			nodeToAsk = neighbors[rand.Intn(len(neighbors))]
		}
		nodeSet[nodeToAsk] = struct{}{}
	}
	for node := range nodeSet {
		nodesToAsk = append(nodesToAsk, node)
	}

	return nodesToAsk
}
