package consumer

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TransferConsumer struct {
	ch        *amqp.Channel
	queueName string
}

func NewTransferConsumer(ch *amqp.Channel, queueName string) *TransferConsumer {
	return &TransferConsumer{
		ch:        ch,
		queueName: queueName,
	}
}

func (c *TransferConsumer) Start() error {
	_, err := c.ch.QueueDeclare(
		c.queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("[TransferConsumer] queue declare error: %v", err)
		return err
	}

	msgs, err := c.ch.Consume(
		c.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("[TransferConsumer] consume error: %v", err)
		return err
	}

	for msg := range msgs {
		log.Printf("[TransferConsumer] message received: %s", msg.Body)
		if err := msg.Ack(false); err != nil {
			log.Printf("[TransferConsumer] ack error: %v", err)
			return err
		}
	}

	return nil
}
