// Autors: Irene Vera Barea, Reinaldo Silva Mejía
package main

//CLIENT
import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func randomOperacion(min int, max int) int {
	num := rand.Intn(max-min) + min
	if num == 0 {
		num = num + 1
	}
	return num
}

func operacionsRPC(n int) (res string, err error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	corrId := randomString(32)

	err = ch.Publish(
		"",          // exchange
		"rpc_queue", // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			ReplyTo:       q.Name,
			Body:          []byte(strconv.Itoa(n)),
		})
	failOnError(err, "Failed to publish a message")

	for d := range msgs {
		if corrId == d.CorrelationId {
			res = (string(d.Body))
			break
		}
	}

	return
}

func main() {

	p := bodyFrom(os.Args)
	s := fmt.Sprintf("Hola, el meu nom és: %s ", p)
	log.Printf(s)
	rand.Seed(time.Now().UnixNano())
	numOp := randomOperacion(1, 6)
	log.Printf(" %s vol fer %d operacions", p, numOp)
	for j := 1; j <= numOp; j++ {
		rand.Seed(time.Now().UnixNano())
		n := randomOperacion(-5, 15)
		log.Printf("%s operació %d: %d", p, j, n)
		log.Printf("Operació sol·licitada!")
		res, err := operacionsRPC(n)
		failOnError(err, "Failed to handle RPC request")
		log.Printf(res)
		log.Printf("%d ----------------------", j)
		time.Sleep(2 * time.Second)
	}
}

func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "Pepe"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}
