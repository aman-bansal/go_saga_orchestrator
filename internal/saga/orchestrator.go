package saga

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/aman-bansal/go_saga_orchestrator/internal/event"
	"github.com/aman-bansal/go_saga_orchestrator/internal/kafka_manager"
	"github.com/aman-bansal/go_saga_orchestrator/internal/mysql_storage"
	"github.com/aman-bansal/go_saga_orchestrator/orchestrator"
	"math"
)

type DefaultSagaOrchestratorRegistry struct {
	sagas          map[string]*SagaOrchestrator
	kafkaPublisher kafka_manager.KafkaEventProducer
	mysqlClient    mysql_storage.StorageClient
}

func (d *DefaultSagaOrchestratorRegistry) Start(sagaId string, data []byte) error {
	saga := d.sagas[sagaId]
	fmt.Println("starting saga for : " + saga.name)
	err := d.kafkaPublisher.Produce(event.KafkaEvent{
		SagaId:    saga.sagaId,
		EventType: saga.sagaWorkflows[0].Transaction.TransactionKey(),
		State:     event.TRANSACTION_START,
		Data:      data,
	})
	if err != nil {
		return err
	}
	return nil
}

func NewDefaultSagaOrchestratorRegistration() orchestrator.SagaOrchestratorRegistration {
	return &DefaultSagaOrchestratorRegistration{}
}

type DefaultSagaOrchestratorRegistration struct {
	brokerHosts []string
	topic       string
	mysqlConfig mysql_storage.MysqlConfig
	sagas       []*SagaOrchestrator
}

func (d *DefaultSagaOrchestratorRegistration) PublishTo(channel string) orchestrator.SagaOrchestratorRegistration {
	d.topic = channel
	return d
}

func (d *DefaultSagaOrchestratorRegistration) WithKafkaConfig(brokerHosts []string) orchestrator.SagaOrchestratorRegistration {
	d.brokerHosts = brokerHosts
	return d
}

func (d *DefaultSagaOrchestratorRegistration) WithMysqlConfig(host, username, password, dbName string) orchestrator.SagaOrchestratorRegistration {
	d.mysqlConfig = mysql_storage.MysqlConfig{
		Host:   host,
		DbName: dbName,
		User:   username,
		Pass:   password,
	}
	return d
}

func (d *DefaultSagaOrchestratorRegistration) AddSagaOrchestrator(saga SagaOrchestrator) orchestrator.SagaOrchestratorRegistration {
	if d.sagas == nil {
		d.sagas = make([]*SagaOrchestrator, 0)
	}
	d.sagas = append(d.sagas, &saga)
	return d
}

func (d *DefaultSagaOrchestratorRegistration) Register() (orchestrator.SagaOrchestratorRegistry, error) {
	//save to mysql DB
	//subscribe to channel via consumer group_id channel_Orchestrator to receive event when one transaction gets completed event
	//publish events
	mysqlClient, err := mysql_storage.NewStorageClient(d.mysqlConfig)
	if err != nil {
		return nil, err
	}
	kafkaEventConsumer, err := kafka_manager.NewKafkaEventConsumer(d.brokerHosts, d.topic, d.topic+"_ORCHESTRATOR")
	if err != nil {
		return nil, err
	}

	kafkaPublisher := kafka_manager.NewKafkaEventProducer(d.brokerHosts, d.topic)

	sagas := make(map[string]*SagaOrchestrator)
	for _, saga := range d.sagas {
		sagas[saga.sagaId] = saga
	}

	registry := &DefaultSagaOrchestratorRegistry{
		sagas:          sagas,
		kafkaPublisher: kafkaPublisher,
		mysqlClient:    mysqlClient,
	}

	go func(*DefaultSagaOrchestratorRegistry) { consumeOrchestratorMessage(registry, kafkaEventConsumer.MessageChannel()) }(registry)

	return registry, nil
}

type SagaWorkflowEvent struct {
	Transaction  orchestrator.Transaction
	Compensation orchestrator.Compensation
	Deadline     int //in seconds
}

type SagaOrchestrator struct {
	sagaId        string
	name          string
	channel       string
	sagaWorkflows []SagaWorkflowEvent
}

func NewDefaultSagaOrchestratorBuilder() orchestrator.SagaOrchestratorBuilder {
	return &DefaultSagaOrchestratorBuilder{}
}

type DefaultSagaOrchestratorBuilder struct {
	sagaId             string
	name               string
	channel            string
	orchestratedEvents []SagaWorkflowEvent
	brokerHosts        []string
	mysqlConfig        mysql_storage.MysqlConfig
}

func (d *DefaultSagaOrchestratorBuilder) SetSagaId(sagaId string) orchestrator.SagaOrchestratorBuilder {
	d.sagaId = sagaId
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

func (d *DefaultSagaOrchestratorBuilder) Build() SagaOrchestrator {
	return SagaOrchestrator{
		sagaId:        d.sagaId,
		name:          d.name,
		channel:       d.channel,
		sagaWorkflows: d.orchestratedEvents,
	}
}

//consume message
// if transaction complete and failed message
// if compensation complete and failed message
// trigger next saga transaction or compensation
func consumeOrchestratorMessage(registry *DefaultSagaOrchestratorRegistry, channel <-chan *sarama.ConsumerMessage) {
	for {
		msg := <-channel
		kafkaEvent := new(event.KafkaEvent)
		e := json.Unmarshal(msg.Value, kafkaEvent)
		if e != nil {
			continue
		}

		//if only saga is present then send the next event
		//do your thing i.e. trigger next else mark complete
		var err error = nil
		if saga, ok := registry.sagas[kafkaEvent.SagaId]; ok {
			_ = saga.sagaId
			switch kafkaEvent.State {
			case event.COMPENSATION_COMPLETE:
				err = triggerNextCmpIfNecessary(registry.kafkaPublisher, saga, kafkaEvent.EventType, kafkaEvent.Data)
			case event.COMPENSATION_FAIL:
				//todo what to do
			case event.TRANSACTION_FAIL:
				err = triggerCmpWhenTxnFails(registry.kafkaPublisher, saga, kafkaEvent.EventType, kafkaEvent.Data)
			case event.TRANSACTION_COMPLETE:
				err = triggerNextTxnIfNecessary(registry.kafkaPublisher, saga, kafkaEvent.EventType, kafkaEvent.Data)
			}
		}

		//check how to manage publihser error
		if err != nil {

		}
	}
}

func triggerNextTxnIfNecessary(kafkaPublisher kafka_manager.KafkaEventProducer, saga *SagaOrchestrator, key string, data []byte) error {
	var index = 0
	for idx, workflow := range saga.sagaWorkflows {
		if workflow.Transaction.TransactionKey() == key {
			index = idx
		}
	}

	if index == 0 {
		return errors.New("no transaction found")
	}

	if index == len(saga.sagaWorkflows)-1 {
		return nil
	}

	return kafkaPublisher.Produce(event.KafkaEvent{
		SagaId:    saga.sagaId,
		EventType: saga.sagaWorkflows[index+1].Transaction.TransactionKey(),
		State:     event.TRANSACTION_START,
		Data:      data,
	})
}

func triggerNextCmpIfNecessary(kafkaPublisher kafka_manager.KafkaEventProducer, saga *SagaOrchestrator, key string, data []byte) error {
	var index = math.MaxInt64
	for idx, workflow := range saga.sagaWorkflows {
		if workflow.Compensation.CompensationKey() == key {
			index = idx
		}
	}

	if index == math.MaxInt64 {
		return errors.New("no compensation found")
	}

	if index == 0 {
		return nil
	}

	return kafkaPublisher.Produce(event.KafkaEvent{
		SagaId:    saga.sagaId,
		EventType: saga.sagaWorkflows[index-1].Compensation.CompensationKey(),
		State:     event.COMPENSATION_START,
		Data:      data,
	})
}

func triggerCmpWhenTxnFails(kafkaPublisher kafka_manager.KafkaEventProducer, saga *SagaOrchestrator, key string, data []byte) error {
	var workflow *SagaWorkflowEvent = nil
	for _, wf := range saga.sagaWorkflows {
		if wf.Transaction.TransactionKey() == key {
			workflow = &wf
		}
	}

	if workflow == nil {
		return errors.New("workflow is not preset")
	}

	return kafkaPublisher.Produce(event.KafkaEvent{
		SagaId:    saga.sagaId,
		EventType: workflow.Compensation.CompensationKey(),
		State:     event.COMPENSATION_START,
		Data:      data,
	})
}
