package common

import (
	"encoding/json"
	"net/http"
)

// Response :
type Response struct {
	Error   interface{} `json:"error,omitempty"`
	Content interface{} `json:"content,omitempty"`
}

//APIResponse : to send response in request
func APIResponse(w http.ResponseWriter, status int, output interface{}) {
	var objResponce Response
	if status == http.StatusOK {
		objResponce.Content = output
	} else {
		objResponce.Error = output
	}
	finalOutput, _ := json.Marshal(objResponce)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(finalOutput)
	return
}
