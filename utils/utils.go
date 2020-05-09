package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

type ClientResponse struct {
	Text string
	Code int
}

func SendBodyResponse(w http.ResponseWriter, text string, code int) {
	resp := ClientResponse{
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

func RandStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
