// Autors: Irene Vera Barea, Reinaldo Silva Mejía
package main

import (
	"log"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	chLadron, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer chLadron.Close()

	cola, err := chLadron.QueueDeclare(
		"lladre", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := chLadron.Consume(
		cola.Name, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("El botí és %s", d.Body)
			log.Printf("M'en vaig correns")
		}
	}()

	log.Printf(" Hola, el meu nom és: Mak the Knife")
	log.Printf(" Això és un atracament, per menys de 20 no m'hi pos")
	<-forever

}
