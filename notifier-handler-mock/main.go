package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

/*
 From official RabbitMQ tutorial:
 https://www.rabbitmq.com/tutorials/tutorial-one-go.html
*/

func main() {
	time.Sleep(time.Second * 15) // waiting until rabbit is alive
	connStr := os.Getenv("RABBIT_MQ_CONN_STRING")
	mqConn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("Can't connect to RabbitMQ on %s, got an error: %s", connStr, err.Error())
		return
	}
	defer mqConn.Close()
	ch, err := mqConn.Channel()
	if err != nil {
		log.Fatalf("Can't create rabbit channel, error: %s", err.Error())
		return
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
		log.Fatalf("Can't declare channel, got an error: %s", err.Error())
		return
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		log.Fatalf("Can't create go chan to consume messages repeatadly, got an error: %s", err.Error())
		return
	}

	forever := make(chan bool)

	go func() {
		fmt.Println("Started receiving messages ...")
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
