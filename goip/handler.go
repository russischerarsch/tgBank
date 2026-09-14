package goip

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/IBM/sarama"
)

type ConsumerHandler struct {
	client *GoIPClient
}

func NewConsumerHandler(client *GoIPClient) *ConsumerHandler {
	return &ConsumerHandler{client: client}
}
func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Println("🔌 Kafka consumer connected")
	return nil
}

func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("🔌 Kafka consumer disconnected")
	return nil
}
func (c *GoIPConsumer) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	return c.consumer.Consume(ctx, topics, handler)
}

// Close — закрыть consumer
func (c *GoIPConsumer) Close() error {
	return c.consumer.Close()
}
func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var command SMSCommand
		if err := json.Unmarshal(msg.Value, &command); err != nil {
			log.Println("failed to parse data")
			continue
		}
		log.Printf("received a command slot=%d, phone=%s, text=%s", command.Slot, command.Phone, command.Text)
		err := h.client.SendSMS(command.Phone, strconv.Itoa(command.Slot), command.Text)
		if err != nil {
			return err
		}
		log.Printf("✅ SMS отправлена на %s через слот %d", command.Phone, command.Slot)
		session.MarkMessage(msg, "")
	}
	return nil
}
