package main

import (
	"github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
}

// NewProducer establishes a connection to RabbitMQ and declares the queue
func NewProducer(url string) (*Producer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Declare the queue
	q, err := ch.QueueDeclare(
		"scan_tasks", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return nil, err
	}

	return &Producer{conn: conn, ch: ch, q: q}, nil
}

// Publish sends a message to the queue
func (p *Producer) Publish(body []byte) error {
	return p.ch.Publish(
		"",       // exchange
		p.q.Name, // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

// Close closes the channel and connection
func (p *Producer) Close() {
	if p.ch != nil {
		p.ch.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
