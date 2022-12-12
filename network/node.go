package network

import (
	"ava/api"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Node struct {
	Addr              int               // Address of the node
	Neighbors         []int             // Addresses of the neighbor nodes
	Chain             []int             // Local chain of the node
	ValidTxThreshold  int               // Threshold to decide if a transaction is valid or not
	DecisionMap       sync.Map          // Map of already-decided transactions
	DecisionChan      chan DecisionOnTx // Channel to listen to decisions made on tx
	SampleSize        int               // Number of nodes to ask each time
	QuorumSize        int               // Number of nodes required to have the same answers for consensus
	DecisionThreshold int               // Number of consecutive successes needed to arrive to a decision
}

type myHTTPHandler func(w http.ResponseWriter, r *http.Request, node *Node)

func (node *Node) Handle(handleF myHTTPHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleF(w, r, node)
	}
}

// start starts the node.
func (node *Node) start() {
	// fmt.Printf("Node %d up and running\n", node.Addr)

	go node.handleDecisionOnTx()

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", node.Handle(validateHandler))
	mux.HandleFunc("/createTx", node.Handle(createTxHandler))
	mux.HandleFunc("/listChain", node.Handle(listChainHandler))
	mux.HandleFunc("/neighbors", node.Handle(listNeighborsHandler))
	http.ListenAndServe(fmt.Sprintf(":%d", node.Addr), mux)
}

func listNeighborsHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	fmt.Printf("Node %d neighbors: %v\n", node.Addr, node.Neighbors)
	w.Write([]byte("OK"))
}

// listChainHandler handles list chain request.
func listChainHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	fmt.Printf("Node %d: %v\n", node.Addr, node.Chain)
	w.Write([]byte("OK"))
}

// createTxHandler handles create tx request.
func createTxHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	var req api.CreateTxRequest
	err := api.ReadJSONRequest(r, &req)
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

	v, ok := node.DecisionMap.Load(req.Tx)
	if ok && v.(int) == DECISION_INVALID {
		fmt.Println("False transaction. Will not propagate.")
	} else if ok && v.(int) == DECISION_WAITING {
		fmt.Println("Transaction being validated. Will not propagate.")
	} else {
		fmt.Println("New transaction. Asking around for decision.")
		node.DecisionChan <- DecisionOnTx{Tx: req.Tx, Decision: DECISION_WAITING}
		txValidation := TxValidation{
			Node: node,
			Tx:   req.Tx,
			Pref: pref,
		}
		go txValidation.DecideOnTx()
	}

	w.Write([]byte("OK"))
}

// validateHandler handles a tx validation request from a node.
func validateHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	var req api.TxValidationRequest
	err := api.ReadJSONRequest(r, &req)
	if err != nil {
		log.Printf("Error reading request: %v", err)
		http.Error(w, "Cannot read request", http.StatusBadRequest)
		return
	}

	v, ok := node.DecisionMap.Load(req.Tx)
	if ok && v.(int) != DECISION_NEW && v.(int) != DECISION_WAITING { // already decided
		var decision bool
		if v.(int) == DECISION_VALID {
			decision = true
		}

		res := api.TxValidationResponse{
			Pref: decision,
		}
		api.WriteJSONResponse(w, r, res)
		return
	}

	// build an initial preference
	var pref bool
	if req.Tx >= node.ValidTxThreshold {
		pref = true
	} else {
		pref = false
	}

	if !ok { // first encounter, need to ask around
		node.DecisionChan <- DecisionOnTx{Tx: req.Tx, Decision: DECISION_WAITING}
		txValidation := TxValidation{
			Node: node,
			Tx:   req.Tx,
			Pref: pref,
		}
		go txValidation.DecideOnTx()
	}

	res := api.TxValidationResponse{
		Pref: pref,
	}
	api.WriteJSONResponse(w, r, res)
}

// handleDecisionOnTx handles the decision signals and updates the decision map accordingly.
func (node *Node) handleDecisionOnTx() {
	for decisionOnTx := range node.DecisionChan {
		v, ok := node.DecisionMap.Load(decisionOnTx.Tx)
		if !ok || (ok && v.(int) != decisionOnTx.Decision) {
			node.DecisionMap.Store(decisionOnTx.Tx, decisionOnTx.Decision)

			// add to chain if valid
			if decisionOnTx.Decision == DECISION_VALID {
				node.Chain = append(node.Chain, decisionOnTx.Tx)
			}
		}
	}
}
