package saga

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/aman-bansal/go_saga_orchestrator/internal/event"
	"github.com/aman-bansal/go_saga_orchestrator/internal/kafka_manager"
	"github.com/aman-bansal/go_saga_orchestrator/participant"
)

type TransactionProcessor struct {
	tType     string
	processor participant.Processor
}

type CompensationProcessor struct {
	cType     string
	processor participant.Processor
}

type DefaultSagaParticipantRegistry struct {
	channel        string
	txn            map[string]*TransactionProcessor
	cmp            map[string]*CompensationProcessor
	brokerHosts    []string
	kafkaPublisher kafka_manager.KafkaEventProducer
}

func NewDefaultSagaParticipantRegistry() participant.SagaParticipantRegistration {
	return &DefaultSagaParticipantRegistry{}
}

func (d *DefaultSagaParticipantRegistry) WithKafkaConfig(brokerHosts []string) participant.SagaParticipantRegistration {
	d.brokerHosts = brokerHosts
	return d
}

func (d *DefaultSagaParticipantRegistry) ListenTo(channel string) participant.SagaParticipantRegistration {
	d.channel = channel
	return d
}

func (d *DefaultSagaParticipantRegistry) AddTransactionProcessor(transactionType string, processor participant.Processor) participant.SagaParticipantRegistration {
	if d.txn == nil {
		d.txn = make(map[string]*TransactionProcessor)
	}

	d.txn[transactionType] = &TransactionProcessor{
		tType:     transactionType,
		processor: processor,
	}
	return d
}

func (d *DefaultSagaParticipantRegistry) AddCompensationProcessor(compensationType string, processor participant.Processor) participant.SagaParticipantRegistration {
	if d.cmp == nil {
		d.cmp = make(map[string]*CompensationProcessor)
	}

	d.cmp[compensationType] = &CompensationProcessor{
		cType:     compensationType,
		processor: processor,
	}
	return d
}

func (d *DefaultSagaParticipantRegistry) Register() error {
	// compensation can be null but not transaction
	// subscribe to events on channel via id channel + type
	// publish to channel when done based on error
	d.kafkaPublisher = kafka_manager.NewKafkaEventProducer(d.brokerHosts, d.channel)
	kafkaEventConsumer, err := kafka_manager.NewKafkaEventConsumer(d.brokerHosts, d.channel, d.channel+"_PARTICIPANT")
	if err != nil {
		return err
	}

	go func() { consumeParticipantMessage(d, kafkaEventConsumer.MessageChannel()) }()
	return nil
}

//consume message
// if transaction STARTS check if you have that
// if compensation STARTS check if you have that
// send complete/fail event
func consumeParticipantMessage(d *DefaultSagaParticipantRegistry, channel <-chan *sarama.ConsumerMessage) {
	for {
		msg := <-channel
		kafkaEvent := new(event.KafkaEvent)
		err := json.Unmarshal(msg.Value, kafkaEvent)
		if err != nil {
			continue
		}

		//todo check what to do with err from publisher
		if kafkaEvent.State == event.COMPENSATION_START {
			//do your thing i.e. send next event
			if cmp, ok := d.cmp[kafkaEvent.EventType]; ok {
				data, err := cmp.processor.Process(kafkaEvent.Data)
				if err != nil {
					err = d.kafkaPublisher.Produce(sendProcessResultEvent(kafkaEvent.SagaId, kafkaEvent.EventType, event.COMPENSATION_FAIL, nil))
				}
				err = d.kafkaPublisher.Produce(sendProcessResultEvent(kafkaEvent.SagaId, kafkaEvent.EventType, event.COMPENSATION_COMPLETE, data))
			}
		} else if kafkaEvent.State == event.TRANSACTION_START {
			//do your thing i.e. send next event
			if txn, ok := d.txn[kafkaEvent.EventType]; ok {
				data, err := txn.processor.Process(kafkaEvent.Data)
				if err != nil {
					err = d.kafkaPublisher.Produce(sendProcessResultEvent(kafkaEvent.SagaId, kafkaEvent.EventType, event.TRANSACTION_FAIL, nil))
				}
				err = d.kafkaPublisher.Produce(sendProcessResultEvent(kafkaEvent.SagaId, kafkaEvent.EventType, event.TRANSACTION_COMPLETE, data))
			}
		}
	}
}

func sendProcessResultEvent(sagaId, eventType string, state event.EventState, data []byte) event.KafkaEvent {
	return event.KafkaEvent{
		SagaId:    sagaId,
		EventType: eventType,
		State:     state,
		Data:      data,
	}
}