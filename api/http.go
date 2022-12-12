package api

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

func ReadJSONRequest(r *http.Request, v interface{}) error {
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

func ReadJSONResponse(r *http.Response, v interface{}) error {
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

func WriteJSONResponse(w http.ResponseWriter, r *http.Request, v interface{}) {
	resJson, err := json.Marshal(v)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Cannot write response", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resJson)
}
