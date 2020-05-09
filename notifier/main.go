package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/DenisAltruist/distsys/utils"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

func sendNotifyOutside(w http.ResponseWriter, encdodedReq []byte) bool {
	mqConn, err := amqp.Dial(os.Getenv("RABBIT_MQ_CONN_STRING"))
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't connect to MQ service: %s", err.Error())
		return false
	}
	defer mqConn.Close()
	ch, err := mqConn.Channel()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't create channel from MQ connection: %s", err.Error())
		return false
	}
	q, err := ch.QueueDeclare(
		os.Getenv("NOTIFIER_CHAN_NAME"), // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't declare queue for notifications: %s", err.Error())
		return false
	}
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        encdodedReq,
		})
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Can't publish message to queue: %s", err.Error())
		return false
	}
	return true
}

func notify(w http.ResponseWriter, r *http.Request) {
	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Can't parse request body, got error: %s", err.Error())
		return
	}
	ok := sendNotifyOutside(w, contents)
	if !ok {
		return
	}
	utils.SendBodyResponse(w, "Success", http.StatusOK)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/notify", notify).Methods("PUT")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("INTERNAL_LISTEN_PORT")), router))
}
