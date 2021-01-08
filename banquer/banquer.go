// Autors: Irene Vera Barea, Reinaldo Silva Mejía
package main

//BANQUER
import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

var cuenta int                                                                     // Variable global para todos los clientes

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func retornarSaldo(n int) string {                                                 // Función para devolver el saldo y mostrarlo por consola   
	aux := cuenta + n                                                                // Utiliza una variable auxiliar para comprobar que no deja en negativo la cuenta antes de modificarla
	if aux >= 0 {                                                                    // Si la operación es posible, modifica la variable global "cuenta"
		cuenta = aux  
		s := fmt.Sprintf("Operació feta! \n Balanç actual: %d", cuenta)
		return s
	} else {
		s := fmt.Sprintf("NO PERMESA, NO HI HA FONS\n Balanç actual: %d", cuenta)       // Sino avisa de que no está permitido
		return s
	}
	return "error"
}

func main() {
	cuenta = 0                                                                         // Inicializa la cuenta a 0
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rpc_queue", // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

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

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			n, err := strconv.Atoi(string(d.Body))                                          // Aquí le dan la operación
			failOnError(err, "Failed to convert body to integer")
			log.Printf(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
			log.Printf(" Operació rebuda: (%d)", n)
			cuentaaux := n + cuenta
			if cuentaaux < 0 {
				log.Printf("NO PERMESA, NO HI HA FONS")
			}
			if cuenta > 20 {                                                                // Avisa al ladrón
				err = ch.Publish(
					"",        // exchange
					cola.Name, // routing key
					false,     // mandatory
					false,     // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(strconv.Itoa(cuenta)),
					})
				failOnError(err, "Failed to publish a message")
			} else {                                                                        // Devuelve el saldo a los clientes
				response := retornarSaldo(n)
				log.Printf("Balanç: %d", cuenta)
				log.Printf("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
				err = ch.Publish(
					"",        // exchange
					d.ReplyTo, // routing key
					false,     // mandatory
					false,     // immediate
					amqp.Publishing{
						ContentType:   "text/plain",
						CorrelationId: d.CorrelationId,
						Body:          []byte((response)),
					})
				failOnError(err, "Failed to publish a message")

				d.Ack(false)
			}
		}
	}()

	log.Printf(" El banc obri")
	currentTime := time.Now()
	str1 := currentTime.Format("2006-01-02 3:4:5")
	str2 := "[*] Esperant clients. Per sortir pitja CTRL+C"
	log.Printf(str1 + str2)

	<-forever
}
