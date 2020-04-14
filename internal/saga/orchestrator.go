package saga

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/aman-bansal/go_saga_orchestrator/internal/event"
	"github.com/aman-bansal/go_saga_orchestrator/internal/kafka_manager"
	"github.com/aman-bansal/go_saga_orchestrator/internal/mysql_storage"
	"github.com/aman-bansal/go_saga_orchestrator/orchestrator"
)

type SagaWorkflowEvent struct {
	Transaction  orchestrator.Transaction
	Compensation orchestrator.Compensation
	Deadline     int //in seconds
}

type DefaultSagaOrchestrator struct {
	sagaId         string
	name           string
	channel        string
	sagaWorkflow   []SagaWorkflowEvent
	kafkaPublisher kafka_manager.KafkaEventProducer
	mysqlClient    mysql_storage.StorageClient
	meta           Meta
}

func (d *DefaultSagaOrchestrator) Start(data []byte) error {
	fmt.Println("starting saga for : " + d.name)
	//todo check what exactly is channel
	err := d.kafkaPublisher.Produce(event.KafkaEvent{
		SagaId:    d.sagaId,
		EventType: d.channel,
		State:     event.COMPENSATION_START,
		Data:      data,
	})
	if err != nil {
		return err
	}
	return nil
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
	mysqlConfig        mysql_storage.MysqlConfig
}

func NewDefaultSagaOrchestratorBuilder() orchestrator.SagaOrchestratorBuilder {
	return &DefaultSagaOrchestratorBuilder{}
}

func (d *DefaultSagaOrchestratorBuilder) WithKafkaConfig(brokerHosts []string) orchestrator.SagaOrchestratorBuilder {
	d.brokerHosts = brokerHosts
	return d
}

func (d *DefaultSagaOrchestratorBuilder) WithMysqlConfig(host string, username string, password string, dbName string) orchestrator.SagaOrchestratorBuilder {
	d.mysqlConfig = mysql_storage.MysqlConfig{
		Host:   host,
		DbName: dbName,
		User:   username,
		Pass:   password,
	}
	return d
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
	//publish events
	mysqlClient, err := mysql_storage.NewStorageClient(d.mysqlConfig)
	if err != nil {
		return nil, err
	}
	kafkaEventConsumer, err := kafka_manager.NewKafkaEventConsumer(d.brokerHosts, d.channel, d.channel+"_ORCHESTRATOR")
	if err != nil {
		return nil, err
	}

	go consumeMessage(kafkaEventConsumer.MessageChannel())

	//consume this event
	kafkaPublisher := kafka_manager.NewKafkaEventProducer(d.brokerHosts, d.channel)
	return &DefaultSagaOrchestrator{
		sagaId:         d.sagaId,
		name:           d.name,
		channel:        d.channel,
		sagaWorkflow:   d.orchestratedEvents,
		kafkaPublisher: kafkaPublisher,
		mysqlClient:    mysqlClient,
		meta: Meta{
			sagaId:           d.sagaId,
			ownerApplication: d.owner,
		},
	}, nil
}

//consume message
// if transaction complete and failed message
// if compensation complete and failed message
// trigger next saga transaction or compensation
func consumeMessage(channel <-chan *sarama.ConsumerMessage) {
	for {
		msg := <-channel
		kafkaEvent := new(event.KafkaEvent)
		err := json.Unmarshal(msg.Value, kafkaEvent)
		if err != nil {
			continue
		}

		if kafkaEvent.State == event.COMPENSATION_COMPLETE || kafkaEvent.State == event.COMPENSATION_FAIL {
			//do your thing
		} else if kafkaEvent.State == event.TRANSACTION_COMPLETE || kafkaEvent.State == event.TRANSACTION_FAIL {

		}
	}
}
