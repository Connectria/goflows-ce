// amqprocessor.go - RabbitMQ start/stop for goflows-scheduler

package main

import (
	"github.com/streadway/amqp"
)

// AMQ is needed for RabbitMQ
type AMQ struct {
	URI             string
	ClientTag       string
	TasksQueue      string // consume goflows-scheduler task
	HistoryExchange string // push logs
}

var amqConnection *amqp.Connection // used for goflows-processor to receive events, tasks, etc.

var amqChannel *amqp.Channel        // used for goflows-processor to receive events, tasks, etc.
var amqHistoryChannel *amqp.Channel // used for goflows-scheduler to publish logs (see logger.go)

var amqTasksQueue amqp.Queue // used for goflows-processor to receive tasks from goflows-scheduler

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

// send info to goflow-processor
func publishRabbitMQ(b []byte) error {
	err := amqChannel.Publish(
		"",                 // exchange
		amqTasksQueue.Name, // routing name
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        b,
		})

	if err != nil {
		return err
	}

	return nil
}

// closeRabbitMQ closes theRabbitMQ connection
func closeRabbitMQ() {
	amqChannel.Close()
	amqHistoryChannel.Close()
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
