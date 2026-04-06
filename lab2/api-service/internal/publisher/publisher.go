package publisher

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

const queueName = "device.events"

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func New(url string) (*Publisher, error) {
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
	return &Publisher{conn: conn, ch: ch}, nil
}

func (p *Publisher) Close() {
	p.ch.Close()
	p.conn.Close()
}

type Event struct {
	Event  string `json:"event"`
	Device any    `json:"device"`
}

func (p *Publisher) Publish(ctx context.Context, eventType string, device any) error {
	body, err := json.Marshal(Event{Event: eventType, Device: device})
	if err != nil {
		return err
	}
	return p.ch.PublishWithContext(ctx, "", queueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}
