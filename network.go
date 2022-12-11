package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

type Node struct {
	Addr              int               // Address of the node
	Neighbors         []int             // Addresses of the neighbor nodes
	Chain             []int             // Local chain of the node
	ValidTxThreshold  int               // Threshold to decide if a transaction is valid or not
	DecisionMap       map[int]int       // Map of already-decided transactions
	DecisionChan      chan DecisionOnTx // Channel to listen to decisions made on tx
	SampleSize        int               // Number of nodes to ask each time
	QuorumSize        int               // Number of nodes required to have the same answers for consensus
	DecisionThreshold int               // Number of consecutive successes needed to arrive to a decision
}

const (
	SERVER_HOST = "localhost"
	START_PORT  = 9000
	SERVER_TYPE = "tcp"
)

// start starts the node.
func (node *Node) start() {
	fmt.Printf("Node %d up and running\n", node.Addr)

	// algorithm-related variables
	node.SampleSize = config.NodeConf.SampleSize
	node.QuorumSize = config.NodeConf.QuorumSize
	node.DecisionThreshold = config.NodeConf.DecisionThreshold

	node.DecisionMap = make(map[int]int)
	node.DecisionChan = make(chan DecisionOnTx)

	go node.handleDecisionOnTx()

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", node.handle(validateHandler))
	mux.HandleFunc("/createTx", node.handle(createTxHandler))
	mux.HandleFunc("/listChain", node.handle(listChainHandler))
	http.ListenAndServe(fmt.Sprintf(":%d", node.Addr), mux)
}

// listChainHandler handles list chain request.
func listChainHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	fmt.Printf("Node %d: %v\n", node.Addr, node.Chain)
	w.Write([]byte("OK"))
}

// createTxHandler handles create tx request.
func createTxHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	var req CreateTxRequest
	err := readJSONRequest(r, &req)
	if err != nil {
		log.Printf("Error reading request: %v", err)
		http.Error(w, "Cannot read request", http.StatusBadRequest)
		return
	}
	fmt.Printf("Node %d received request to create tx %d\n", node.Addr, req.Tx)

	var pref bool
	if req.Tx >= node.ValidTxThreshold {
		pref = true
	} else {
		pref = false
	}
	if node.DecisionMap[req.Tx] == INVALID {
		fmt.Println("False transaction. Will not propagate.")
	} else if node.DecisionMap[req.Tx] == WAITING {
		fmt.Println("Transaction being validated. Will not propagate.")
	} else {
		fmt.Println("New transaction. Asking around for decision.")
		node.DecisionChan <- DecisionOnTx{Tx: req.Tx, Decision: WAITING}
		txValidation := TxValidation{
			Node: node,
			Tx:   req.Tx,
			Pref: pref,
		}
		go txValidation.decideOnTx()
	}

	w.Write([]byte("OK"))
}

// validateHandler handles a tx validation request from a node.
func validateHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	var req TxValidationRequest
	err := readJSONRequest(r, &req)
	if err != nil {
		log.Printf("Error reading request: %v", err)
		http.Error(w, "Cannot read request", http.StatusBadRequest)
		return
	}

	currentDecision := node.DecisionMap[req.Tx]
	if currentDecision != NEW && currentDecision != WAITING { // already decided
		var decision bool
		if currentDecision == VALID {
			decision = true
		}

		res := TxValidationResponse{
			Pref: decision,
		}

		writeJSONResponse(w, r, res)
	} else {
		// build an initial preference
		var pref bool
		if req.Tx >= node.ValidTxThreshold {
			pref = true
		} else {
			pref = false
		}

		if currentDecision == NEW { // first encounter, need to ask around
			node.DecisionChan <- DecisionOnTx{Tx: req.Tx, Decision: WAITING}
			txValidation := TxValidation{
				Node: node,
				Tx:   req.Tx,
				Pref: pref,
			}
			go txValidation.decideOnTx()
		}

		res := TxValidationResponse{
			Pref: pref,
		}
		writeJSONResponse(w, r, res)
	}
}

// askToValidateTx sends a request to validate a tx to a node.
func askToValidateTx(nodeAddr int, tx int, resChan chan<- bool, errChan chan<- error) {
	myMsg := TxValidationRequest{Tx: tx}
	jsonBody, err := json.Marshal(myMsg)
	if err != nil {
		errChan <- err
	}
	url := fmt.Sprintf("http://localhost:%d/validate", nodeAddr)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		errChan <- err
	}
	var res TxValidationResponse
	err = readJSONResponse(resp, &res)
	if err != nil {
		errChan <- err
	}

	resChan <- res.Pref
}

// handleDecisionOnTx handles the decision signals and updates the decision map accordingly.
func (node *Node) handleDecisionOnTx() {
	for decisionOnTx := range node.DecisionChan {
		node.DecisionMap[decisionOnTx.Tx] = decisionOnTx.Decision

		// add to chain if valid
		if decisionOnTx.Decision == VALID {
			node.Chain = append(node.Chain, decisionOnTx.Tx)
		}
	}
}

// buildNetwork builds the test network with n nodes.
func buildNetwork(n int) []Node {
	network := make([]Node, 0)

	// build a map of node connections first
	nodeConn := buildNodeConn(n)

	// spin up the nodes
	for i := 0; i < n; i++ {
		var node Node
		nodeAddr := START_PORT + i

		if i > 0 && i%5 == 0 { // every 5 nodes, spin up a malicious node
			// malicious node
			node = Node{
				Addr:             nodeAddr,
				Neighbors:        nodeConn[nodeAddr],
				Chain:            make([]int, 0),
				ValidTxThreshold: 20, // malicious node trying to approve tx bigger than or equal to 20
			}
		} else {
			// valid node
			node = Node{
				Addr:             nodeAddr,
				Neighbors:        nodeConn[nodeAddr],
				Chain:            make([]int, 0),
				ValidTxThreshold: 100, // valid node only approve tx bigger than or equal to 100
			}
		}
		network = append(network, node)
		go node.start()
	}

	return network
}

// buildNodeConn builds and returns a map of node connections.
func buildNodeConn(n int) map[int][]int {
	// each node will connect to at least [neighbors] node. The number of minimum neighbors is written in the config file.
	// there are nodes that might connect to more nodes than necessary.
	nodeConn := make(map[int][]int)
	for i := 0; i < n; i++ {
		nodeAddr := START_PORT + i

		if _, ok := nodeConn[nodeAddr]; !ok {
			nodeConn[nodeAddr] = make([]int, 0)
		}

		neighborSet := make(map[int]struct{})
		for _, neighbor := range nodeConn[nodeAddr] { // check already-formed neighbors
			neighborSet[neighbor] = struct{}{}
		}

		numberOfNeighborToAdd := config.NetworkConf.Neighbors - len(neighborSet)
		for k := 0; k < numberOfNeighborToAdd; k++ {
			neighbor := START_PORT + rand.Intn(n)
			for _, ok := neighborSet[neighbor]; ok || neighbor == nodeAddr; _, ok = neighborSet[neighbor] { // retry until getting a new node
				neighbor = START_PORT + rand.Intn(n)
			}
			neighborSet[neighbor] = struct{}{}
		}
		for neighbor := range neighborSet { // form connection between 2 nodes
			nodeConn[nodeAddr] = append(nodeConn[nodeAddr], neighbor)

			if _, ok := nodeConn[neighbor]; !ok {
				nodeConn[neighbor] = make([]int, 0)
			}
			nodeConn[neighbor] = append(nodeConn[neighbor], nodeAddr)
		}
	}

	return nodeConn
}
