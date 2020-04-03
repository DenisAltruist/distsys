package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type clientResponse struct {
	Text string
	Code int
}

func SendBodyResponse(w http.ResponseWriter, text string, code int) {
	resp := clientResponse{
		Text: text,
		Code: code,
	}
	encodedJson, _ := json.Marshal(&resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, "%s\n", string(encodedJson))
}

func SendError(w http.ResponseWriter, code int, formatMsg string, args ...interface{}) {
	msg := formatMsg
	if len(args) != 0 {
		msg = fmt.Sprintf(formatMsg, args)
	}
	log.Printf(msg)
	SendBodyResponse(w, msg, code)
}
