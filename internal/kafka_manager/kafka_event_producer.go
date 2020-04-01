package kafka_manager

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/aman-bansal/go_saga_orchestrator/internal/event"
)

type KafkaEventProducer interface {
	Produce(event event.Event) error
	Close()
}

type DefaultKafkaEventProducer struct {
	brokerHosts []string
	topic       string
	groupId     string
	producer    sarama.SyncProducer
}

func NewKafkaEventProducer(brokerHosts []string, topic string) KafkaEventProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(brokerHosts, config)
	if err != nil {
		// Should not reach here
		panic(err)
	}
	return &DefaultKafkaEventProducer{
		brokerHosts: brokerHosts,
		topic:       topic,
		producer:    producer,
	}
}

func (k *DefaultKafkaEventProducer) Produce(event event.Event) error {
	msg := &sarama.ProducerMessage{
		Topic: k.topic,
		Value: sarama.ByteEncoder(event.GetData()),
	}

	partition, offset, err := k.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", k.topic, partition, offset)
	return nil
}

func (k *DefaultKafkaEventProducer) Close() {
	if err := k.producer.Close(); err != nil {
		// Should not reach here
		panic(err)
	}
}