// amqprocessor.go - RabbitMQ goflow-processor

package main

import (
	"github.com/streadway/amqp"
)

// AMQ is needed for RabbitMQ
type AMQ struct {
	URI             string
	ClientTag       string
	EventsQueue     string // consume opsgenie-reader events
	TasksQueue      string // consume goflows-scheduler task
	HistoryExchange string // push logs
}

var amqConnection *amqp.Connection // used for goflows-processor to receive events, tasks, etc.
var amqChannel *amqp.Channel       // used for goflows-processor to receive events, tasks, etc.

var amqEventsQueue amqp.Queue              // used for goflows-processor to receive tasks from opsgenie-reader
var amqEventsConsumed <-chan amqp.Delivery // go channel for messages received from opsgenie-reader

var amqTasksQueue amqp.Queue              // used for goflows-processor to receive tasks from goflows-scheduler
var amqTasksConsumed <-chan amqp.Delivery // go channel for messages received from goflows-scheduler

var amqHistoryChannel *amqp.Channel // used for goflows-processor to publish logs (see logger.go)

// setup connection to RabbitMQ
func initRabbitMQ() error {
	var err error

	amqConnection, err = amqp.Dial(cfg.AMQprocessor.URI)
	if err != nil {
		return err
	}

	amqChannel, err = amqConnection.Channel()
	if err != nil {
		return err
	}

	// opsgenie-reader
	amqEventsQueue, err = amqChannel.QueueDeclare(
		cfg.AMQprocessor.EventsQueue, // queue name
		true,                         // durable
		false,                        // delete when unused
		false,                        // exclusive
		false,                        // no-wait
		nil,                          // arguments
	)
	if err != nil {
		return err
	}

	amqEventsConsumed, err = amqChannel.Consume(
		amqEventsQueue.Name,                  // queue name
		cfg.AMQprocessor.ClientTag+"-events", // consumer
		true,                                 // auto-ack
		false,                                // exclusive
		false,                                // no-local
		false,                                // no-wait
		nil,                                  // arguments
	)
	if err != nil {
		return err
	}

	// goflows-scheduler
	amqTasksQueue, err = amqChannel.QueueDeclare(
		cfg.AMQprocessor.TasksQueue, // queue name
		true,                        // durable
		false,                       // delete when unused
		false,                       // exclusive
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		return err
	}

	amqTasksConsumed, err = amqChannel.Consume(
		amqTasksQueue.Name,                  // queue name
		cfg.AMQprocessor.ClientTag+"-tasks", // consumer
		true,                                // auto-ack
		false,                               // exclusive
		false,                               // no-local
		false,                               // no-wait
		nil,                                 // arguments
	)
	if err != nil {
		return err
	}

	// logs (i.e., history)
	amqHistoryChannel, err = amqConnection.Channel()
	if err != nil {
		return err
	}

	if err = amqHistoryChannel.ExchangeDeclare(
		cfg.AMQprocessor.HistoryExchange, // name of exchange
		"fanout",                         // type
		true,                             // durable
		false,                            // delete when finished
		false,                            // internal
		false,                            // no-wait
		nil,                              // arguments
	); err != nil {
		return err
	}

	return nil
}

// setup RabbitMQ during CLI (for history)
func initRabbitMQCLI() error {
	var err error

	amqConnection, err = amqp.Dial(cfg.AMQprocessor.URI)
	if err != nil {
		return err
	}

	amqChannel, err = amqConnection.Channel()
	if err != nil {
		return err
	}

	// logs (i.e., history)
	amqHistoryChannel, err = amqConnection.Channel()
	if err != nil {
		return err
	}

	if err = amqHistoryChannel.ExchangeDeclare(
		cfg.AMQprocessor.HistoryExchange+"-cli", // name of exchange
		"fanout", // type
		true,     // durable
		true,     // delete when finished
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	); err != nil {
		return err
	}

	// bind the cli exchange to the daemon exchange
	amqHistoryChannel.ExchangeBind(
		cfg.AMQprocessor.HistoryExchange, // destination exchange
		"",                               // routing key
		cfg.AMQprocessor.HistoryExchange+"-cli", // source exchange
		false,
		nil)

	cliRMQ = true
	return nil
}

// closeRabbitMQ closes theRabbitMQ connection
func closeRabbitMQ() {
	amqChannel.Close()
	amqConnection.Close()
}

// publish logs
func publishLog(b []byte) error {
	return amqHistoryChannel.Publish(
		cfg.AMQprocessor.HistoryExchange, // name of exchange
		"",                               // routing key
		false,                            // mandatory
		false,                            // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        b,
		},
	)
}
