package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type TxValidationRequest struct {
	Tx int `json:"tx"`
}

type TxValidationResponse struct {
	Pref bool `json:"pref"`
}

type CreateTxRequest struct {
	Tx int `json:"tx"`
}

type myHTTPHandler func(w http.ResponseWriter, r *http.Request, node *Node)

func (node *Node) handle(handleF myHTTPHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleF(w, r, node)
	}
}

func readJSONRequest(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}
	return nil
}

func readJSONResponse(r *http.Response, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}
	return nil
}

func writeJSONResponse(w http.ResponseWriter, r *http.Request, v interface{}) {
	resJson, err := json.Marshal(v)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Cannot write response", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resJson)
}
