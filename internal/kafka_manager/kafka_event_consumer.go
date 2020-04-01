package kafka_manager

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"time"
)

type KafkaEventConsumer interface {
	MessageChannel() <-chan *sarama.ConsumerMessage
}

type DefaultKafkaEventConsumer struct {
	messageChannel chan *sarama.ConsumerMessage
	brokerHosts    []string
	topics         string
	groupId        string
}

//handle gracefull shutdown
func NewKafkaEventConsumer(brokerHosts []string, topic string, groupId string) (KafkaEventConsumer, error) {
	ctx := context.Background()
	client, err := sarama.NewClient(brokerHosts, nil)
	if err != nil {
		fmt.Println("Could not create consumer: ", err)
		return nil, err
	}

	consumerGroup, err := sarama.NewConsumerGroupFromClient(groupId, client)
	if err != nil {
		fmt.Println("Could not create consumer: ", err)
		return nil, err
	}

	eventConsumer := &DefaultKafkaEventConsumer{
		messageChannel: make(chan *sarama.ConsumerMessage),
		brokerHosts:    brokerHosts,
		topics:         topic,
		groupId:        groupId,
	}

	go func(kc *DefaultKafkaEventConsumer) {
		time.Sleep(2 * time.Millisecond)
		for {
			err := consumerGroup.Consume(ctx, []string{topic}, &KafkaConsumerGroupHandler{eventConsumer.messageChannel})
			if err != nil {
				fmt.Printf("consumer returned with error = %v, exiting", err)
				return
			} else {
				fmt.Println("consumer returned without error")
			}
		}
	} (eventConsumer)

	return eventConsumer, nil
}

func (k *DefaultKafkaEventConsumer) MessageChannel() <-chan *sarama.ConsumerMessage {
	return k.messageChannel
}