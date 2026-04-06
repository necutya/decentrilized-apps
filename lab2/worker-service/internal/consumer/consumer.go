package consumer

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/necutya/decentrilized_apps/lab2/worker-service/internal/service"
)

const queueName = "device.events"

type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	svc  *service.WorkerService
}

func New(url string, svc *service.WorkerService) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &Consumer{conn: conn, ch: ch, svc: svc}, nil
}

func (c *Consumer) Close() {
	c.ch.Close()
	c.conn.Close()
}

func (c *Consumer) Consume() error {
	msgs, err := c.ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	log.Printf("consuming queue %s", queueName)
	for d := range msgs {
		if err := c.svc.Process(d.Body); err != nil {
			log.Printf("process error: %v — nacking", err)
			d.Nack(false, false)
			continue
		}
		d.Ack(false)
	}
	return nil
}
