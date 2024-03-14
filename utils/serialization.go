package utils

import (
	"encoding/json"
	"net/http"
)

func Send(w http.ResponseWriter, data interface{}) {
	protoData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Error marshaling to json", http.StatusInternalServerError)
		return
	}

	// Send the Protobuf data as the HTTP response
	w.Header().Set("Content-Type", "application/json")
	w.Write(protoData)
}

func Read(r *http.Request, models interface{}) error {
	// Create an instance of YourStruct to decode the JSON into

	// Decode the JSON body into the struct
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(models)
	if err != nil {
		return err
	}

	return nil
}
