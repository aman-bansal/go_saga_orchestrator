package kafka_manager

import (
	"fmt"
	"github.com/Shopify/sarama"
)

type KafkaConsumerGroupHandler struct {
	messageChannel    chan *sarama.ConsumerMessage
}

func (h *KafkaConsumerGroupHandler) Setup(cgs sarama.ConsumerGroupSession) error {
	fmt.Printf("setup is called, metadata = %+v\n", cgs)
	return nil
}

func (h *KafkaConsumerGroupHandler) Cleanup(cgs sarama.ConsumerGroupSession) error {
	fmt.Printf("cleanup is called, metadata = %+v\n", cgs)
	return nil
}

func (h *KafkaConsumerGroupHandler) ConsumeClaim(cgs sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	fmt.Println(nil, "Running KafkaGroupConsumer")
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				fmt.Println("channel already closed nothing to read")
				return nil
			}
			if msg == nil {
				fmt.Println("rebalance/quit triggered, exiting")
				return nil
			}
			//handle message
			h.messageChannel <- msg
			cgs.MarkMessage(msg, "")

		case <-cgs.Context().Done():
			return nil
		}
	}
}
