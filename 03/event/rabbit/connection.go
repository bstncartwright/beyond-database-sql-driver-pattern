package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/bstncartwright/beyond-database-sql-driver-pattern/03/event/driver"
	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbit struct {
	cfg  config
	conn *amqp.Connection
	ch   *amqp.Channel
}

type route struct {
	resource string
	action   string
}

func (r *rabbit) Push(ctx context.Context, topic driver.Topic, m driver.Message) error {
	defer r.ch.Close()
	defer r.conn.Close()
	// assert it has the right type
	messageType := reflect.TypeOf(m)
	topicType := reflect.TypeOf(topic.Type)

	if reflect.TypeOf(m) != reflect.TypeOf(topic.Type) {
		return fmt.Errorf("message type: %s does not match topic type: %s", messageType, topicType)
	}

	confirms := make(chan amqp.Confirmation)
	errs := make(chan error, 1)

	err := r.ch.Confirm(false)
	if err != nil {
		close(confirms)
		return fmt.Errorf("channel could not be put into confirm mode: %w", err)
	}

	r.ch.NotifyPublish(confirms)
	// TODO: this channel is essential and needs to be improved so unacknowledged messages are handled better
	// ideally through the error return on the push
	go func() {
		for confirm := range confirms {
			if confirm.Ack {
				// code when messages is confirmed
				errs <- nil
			} else {
				// code when messages is nack-ed
				log.Printf("Nacked")
				errs <- fmt.Errorf("unable to acknowledge the message on rabbit mq")
			}
		}
	}()

	body, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = r.ch.Publish(
		fmt.Sprintf("%s%s", os.Getenv("BUS_PREFIX"), topic.Exchange), // exchange
		fmt.Sprintf("%s%s", topic.Name, "*"),                         // routing key
		true,                                                         // mandatory
		false,                                                        // immediate
		amqp.Publishing{
			Timestamp:    time.Now(),
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: 2, // persistent
		})
	if err != nil {
		return err
	}

	if err := <-errs; err != nil {
		return fmt.Errorf("confirmation failed: %w", err)
	}

	return nil
}

func (r *rabbit) Subscribe(ctx context.Context, topics []driver.Topic) error {
	defer r.ch.Close()
	defer r.conn.Close()
	// create the exchanges
	err := r.declareExchange(topics)
	if err != nil {
		return fmt.Errorf("unable to create exchanges: %w", err)
	}

	// ensure our queue exists, using configuration name
	_, err = r.declareQueue(r.cfg.Name)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	// bind topics to our queue
	for _, topic := range topics {
		innerErr := r.ch.QueueBind(
			fmt.Sprintf("%s%s", os.Getenv("BUS_PREFIX"), r.cfg.Name),
			topic.Name,
			fmt.Sprintf("%s%s", os.Getenv("BUS_PREFIX"), topic.Exchange),
			false, // no wait
			nil,   // args
		)
		if innerErr != nil {
			return fmt.Errorf(
				"unable to bind queue %q with exchange %q for topic %q: err: %w",
				fmt.Sprintf("%s%s", os.Getenv("BUS_PREFIX"), r.cfg.Name),
				fmt.Sprintf("%s%s", os.Getenv("BUS_PREFIX"), topic.Exchange),
				topic.Name,
				innerErr,
			)
		}
	}

	msgs, err := r.ch.Consume(
		fmt.Sprintf("%s%s", os.Getenv("BUS_PREFIX"), r.cfg.Name), // queue
		"",    // consumer
		false, // auto awk
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("unable to consume message from queue %s: %w", r.cfg.Name, err)
	}

	for msg := range msgs {
		go func(msg amqp.Delivery) {
			// split the routing key, should be 3 parts, into a standard struct.
			// log an error if it fails.
			key, err := routingKeySplit(msg.RoutingKey)
			if err != nil {
				// log our error
				log.Print(err)

				// sent no acknowledgement back, and requeue the message (this will blow up data dog intentionally!)
				err = msg.Nack(false, true)
				if err != nil {
					log.Printf("rabbit acknowledgement unsuccessful: %s", err)
				}

				return
			}

			t := key.match(topics)
			// there are some arguments that we should always pass the delivery and not just the body
			// so we can act on other things too... hypotheticals though so I'm not adding it
			err = t.Consumer(msg.Body)
			if err != nil {
				log.Printf(
					"consumer had an issue processing an event message: %s, err: %s",
					msg.Body,
					err,
				)

				err = msg.Nack(false, true)
				if err != nil {
					log.Printf("rabbit acknowledgement unsuccessful: %s", err)
				}

				return
			}
			err = msg.Ack(false)
			if err != nil {
				log.Printf("rabbit acknowledgement unsuccessful: %s", err)
			}
		}(msg)
	}
	return nil
}

const (
	// ExpiresTime sets the time that if no consumers are interacting with the
	// queue, the queue will be removed in that time. Currently set to 3 days.
	ExpiresTime = 259200000
)

// QueueType defines the queue type
// each queue is in a quorum to add resiliency and speed of recovery this
// must be set by client (and cannot be set by a policy) see:
// https://www.rabbitmq.com/quorum-queues.html#declaring
var QueueType = "quorum"

func (r *rabbit) declareQueue(name string) (amqp.Queue, error) {
	// CHANGING ANY OF THE BELOW WILL CAUSE YOUR SERVICE TO NOT START IF AN EXISTING QUEUE IS DECLARED WITH DIFFERENT VALUES
	args := make(amqp.Table)
	args["x-expires"] = ExpiresTime
	args["x-queue-type"] = QueueType

	// QueueDeclarePassive can also be used to check for the existence of a queue
	// if one is found we could just connect to it and use whatever params are currently existing.
	// this may be an appealing option for the operations team, as we can change params live without causing breaks
	// TODO: make a pro con list lawls
	queue, err := r.ch.QueueDeclare(
		// if name is left empty, rabbit will decide one randomly, it's important we always define one
		fmt.Sprintf("%s%s", os.Getenv("BUS_PREFIX"), name), // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // arguments
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("unable to create queue: %w", err)
	}

	return queue, nil
}

func (r *rabbit) declareExchange(topics []driver.Topic) error {
	for _, t := range topics {
		err := r.ch.ExchangeDeclare(
			fmt.Sprintf("%s%s", os.Getenv("BUS_PREFIX"), t.Exchange), // exchange name
			"topic", // type (always topic)
			true,    // durable
			false,   // auto-deleted
			false,   // internal
			false,   // no-wait
			nil,     // arguments
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func routingKeySplit(key string) (route, error) {
	r := route{}

	s := strings.Split(key, ".")
	if len(s) < 3 {
		return r, fmt.Errorf("routing key improperly formatted: %s cannot process message", key)
	}

	r.resource = s[0]
	r.action = s[1]

	return r, nil
}

func (r route) match(topics []driver.Topic) driver.Topic {
	for _, t := range topics {
		// do the resources match, or does our topic want all
		if t.Resource() == r.resource || t.Resource() == "*" {
			// do the verbs match, or does our topic want all
			if t.Action() == r.action || t.Action() == "*" {
				return t
			}
		}
	}
	return driver.Topic{}
}
