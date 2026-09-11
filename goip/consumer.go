package goip

import "github.com/IBM/sarama"

type SMSCommand struct {
	Slot  int    `json:"slot"`
	Phone string `json:"phone"`
	Text  string `json:"text"`
}
type GoIPConsumer struct {
	client   *GoIPClient
	consumer sarama.ConsumerGroup
}

func NewGoIPConsumer(broker []string, groupID string, goipclient *GoIPClient) (*GoIPConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	consumer, err := sarama.NewConsumerGroup(broker, groupID, config)
	if err != nil {
		return nil, err
	}
	return &GoIPConsumer{
		client:   goipclient,
		consumer: consumer,
	}, nil
}
