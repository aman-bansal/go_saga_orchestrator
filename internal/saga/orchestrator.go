package saga

import (
	"github.com/aman-bansal/go_saga_orchestrator/internal/kafka_manager"
	"github.com/aman-bansal/go_saga_orchestrator/orchestrator"
)

type SagaWorkflowEvent struct {
	Transaction  orchestrator.Transaction
	Compensation orchestrator.Compensation
	Deadline     int //in seconds
}

type DefaultSagaOrchestrator struct {
	name           string
	channel        string
	sagaWorkflow   []SagaWorkflowEvent
	kafkaPublisher kafka_manager.KafkaEventProducer
	meta           Meta
}

func (d DefaultSagaOrchestrator) Start(data []byte) error {
	panic("implement me")
}

type Meta struct {
	sagaId           string
	ownerApplication string
}

type DefaultSagaOrchestratorBuilder struct {
	sagaId             string
	owner              string
	name               string
	channel            string
	orchestratedEvents []SagaWorkflowEvent
	brokerHosts        []string
}

func NewDefaultSagaOrchestratorBuilder() *DefaultSagaOrchestratorBuilder {
	return &DefaultSagaOrchestratorBuilder{}
}

func (d *DefaultSagaOrchestratorBuilder) WithKafkaConfig(brokerHosts []string) orchestrator.SagaOrchestratorBuilder {
	d.brokerHosts = brokerHosts
	return d
}

func (d *DefaultSagaOrchestratorBuilder) WithMysqlConfig(host string, username string, password string) orchestrator.SagaOrchestratorBuilder {
	panic("implement me")
}

func (d *DefaultSagaOrchestratorBuilder) SetSagaId(sagaId string) orchestrator.SagaOrchestratorBuilder {
	d.sagaId = sagaId
	return d
}

func (d *DefaultSagaOrchestratorBuilder) SetApplicationOwner(owner string) orchestrator.SagaOrchestratorBuilder {
	d.owner = owner
	return d
}

func (d *DefaultSagaOrchestratorBuilder) PublishTo(channel string) orchestrator.SagaOrchestratorBuilder {
	d.channel = channel
	return d
}

func (d *DefaultSagaOrchestratorBuilder) Name(name string) orchestrator.SagaOrchestratorBuilder {
	d.name = name
	return d
}

func (d *DefaultSagaOrchestratorBuilder) Add(transaction orchestrator.Transaction, compensation orchestrator.Compensation, deadline int) orchestrator.SagaOrchestratorBuilder {
	if d.orchestratedEvents == nil {
		d.orchestratedEvents = make([]SagaWorkflowEvent, 0)
	}

	d.orchestratedEvents = append(d.orchestratedEvents, SagaWorkflowEvent{
		Transaction:  transaction,
		Compensation: compensation,
		Deadline:     deadline,
	})
	return d
}

func (d *DefaultSagaOrchestratorBuilder) Build() (orchestrator.SagaOrchestrator, error) {
	//save to mysql DB
	//subscribe to channel via consumer group_id channel_Orchestrator to receive event when one transaction gets completed event
	//publish eventsx
	kafkaEventConsumer, err := kafka_manager.NewKafkaEventConsumer(d.brokerHosts, d.channel, d.channel+"_ORCHESTRATOR")
	if err != nil {
		return nil, err
	}

	//consume message
	// if transaction complete and failed message
	// if compensation complete and failed message
	// trigger next saga transaction or compensation
	go func() {
		kafkaEventConsumer.MessageChannel()
	}()

	//consume this event
	kafkaPublisher := kafka_manager.NewKafkaEventProducer(d.brokerHosts, d.channel)
	return &DefaultSagaOrchestrator{
		name:           d.name,
		channel:        d.channel,
		sagaWorkflow:   d.orchestratedEvents,
		kafkaPublisher: kafkaPublisher,
		meta: Meta{
			sagaId:           d.sagaId,
			ownerApplication: d.owner,
		},
	}, nil
}
