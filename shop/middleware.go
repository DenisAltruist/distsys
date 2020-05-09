package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DenisAltruist/distsys/utils"

	"github.com/gorilla/mux"
)

func authMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" { // no need to authorize GET requests
				next.ServeHTTP(w, r)
				return
			}
			authToken, err := getAuthToken(r)
			if err != nil {
				utils.SendError(w, http.StatusUnauthorized, "Can't parse auth token, got an error: %s", err.Error())
				return
			}
			req, err := http.NewRequest("GET", os.Getenv("AUTH_VALIDATION_ROUTE"), nil)
			if err != nil {
				utils.SendError(w, http.StatusUnauthorized, "Can't create auth request: %s", err.Error())
				return
			}
			q := req.URL.Query()
			q.Add("token", authToken)
			req.URL.RawQuery = q.Encode()
			client := http.Client{
				Timeout: 5 * time.Second,
			}
			resp, err := client.Do(req)
			if err != nil {
				utils.SendError(w, http.StatusUnauthorized, "Can't make validate GET request: %s", err.Error())
				return
			}
			if resp.StatusCode == http.StatusUnauthorized {
				utils.SendError(w, http.StatusUnauthorized, "Not authorized")
				return
			}
			defer resp.Body.Close()
			var respJson utils.ClientResponse
			message, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				utils.SendError(w, http.StatusUnauthorized, "Can't read bytes from body of validation request: %s", err.Error())
				return
			}
			log.Printf("%s\n", string(message))
			err = json.Unmarshal([]byte(string(message)), &respJson)
			if err != nil {
				utils.SendError(w, http.StatusUnauthorized, "Can't convert bytes from body of validation request to JSON: %s", err.Error())
				return
			}
			next.ServeHTTP(w, r)
			return
		})
	}
}

func getAuthToken(r *http.Request) (string, error) {
	authString := r.Header.Get("Authorization")
	splitAuth := strings.Split(authString, " ")
	if len(splitAuth) != 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("Can't retrieve Bearer from Authorization")
	}
	return splitAuth[1], nil
}
