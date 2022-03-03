// amqstatus.go - update status to RabbitMQ

package goflows

import (
	"fmt"

	"github.com/streadway/amqp"
)

// AMQ is needed for status queue
type AMQ struct {
	URI                string
	JobStatusExchange  string
	JobStatusClientTag string
}

type internalAMQ struct {
	configInfo          AMQ
	jobStatusConnection *amqp.Connection
	jobStatusChannel    *amqp.Channel
}

// setup connection to RabbitMQ
func (gf *GoFlow) initJobStatusQ() error {
	var err error

	// NOTE: closing connection is not deferred, it is closed at the end of the goflow "run"
	gf.statusQ.jobStatusConnection, err = amqp.Dial(gf.statusQ.configInfo.URI)
	if err != nil {
		return fmt.Errorf("error: failed to connect to RabbitMQ - %v", err.Error())
	}

	// NOTE: closing channel is not deferred, it is closed at the end of the goflow "run"
	gf.statusQ.jobStatusChannel, err = gf.statusQ.jobStatusConnection.Channel()
	if err != nil {
		return fmt.Errorf("error: failed to open channel - %v", err.Error())
	}

	if err = gf.statusQ.jobStatusChannel.ExchangeDeclare(
		gf.statusQ.configInfo.JobStatusExchange, // name of exchange
		"fanout",                                // type
		true,                                    // durable
		false,                                   // delete when finished
		false,                                   // internal
		false,                                   // no-wait
		nil,                                     // arguments
	); err != nil {
		return fmt.Errorf("error: failed to declare exchange - %v", err.Error())
	}

	return nil
}

func (gf *GoFlow) updateJobStatusQ(b []byte) error {
	return gf.statusQ.jobStatusChannel.Publish(
		gf.statusQ.configInfo.JobStatusExchange, // name of exchange
		"",                                      // routing key
		false,                                   // mandatory
		false,                                   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        b,
		},
	)
}

// CloseJobsStatusQ closes a RabbitMQ connection that used for a GoFlow "run"
func (gf *GoFlow) CloseJobsStatusQ() {
	gf.statusQ.jobStatusChannel.Close()
	gf.statusQ.jobStatusConnection.Close()
}
