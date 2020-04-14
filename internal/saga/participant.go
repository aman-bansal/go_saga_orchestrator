package saga

import (
	"github.com/aman-bansal/go_saga_orchestrator/internal/kafka_manager"
	"github.com/aman-bansal/go_saga_orchestrator/internal/mysql_storage"
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

type DefaultSagaParticipant struct {
	channel        string
	txn            TransactionProcessor
	cmp            CompensationProcessor
	brokerHosts    []string
	kafkaPublisher kafka_manager.KafkaEventProducer
	mysqlClient    mysql_storage.StorageClient
}

func NewDefaultSagaParticipant() participant.SagaParticipant {
	return &DefaultSagaParticipant{}
}

func (d *DefaultSagaParticipant) WithKafkaConfig(brokerHosts []string) participant.SagaParticipant {
	d.brokerHosts = brokerHosts
	return d
}

func (d *DefaultSagaParticipant) WithMysqlConfig(host string, username string, password string, dbName string) participant.SagaParticipant {
	panic("implement me")
}

func (d *DefaultSagaParticipant) ListenTo(channel string) participant.SagaParticipant {
	d.channel = channel
	return d
}

func (d *DefaultSagaParticipant) AddTransactionProcessor(transactionType string, processor participant.Processor) participant.SagaParticipant {
	d.txn = TransactionProcessor{
		tType:     transactionType,
		processor: processor,
	}
	return d
}

func (d *DefaultSagaParticipant) AddCompensationProcessor(compensationType string, processor participant.Processor) participant.SagaParticipant {
	d.cmp = CompensationProcessor{
		cType:     compensationType,
		processor: processor,
	}
	return d
}

func (d *DefaultSagaParticipant) Start() error {
	// compensation can be null but not transaction
	// subscribe to events on channel via id channel + type
	// publish to channel when done based on error

	kafkaEventConsumer, err := kafka_manager.NewKafkaEventConsumer(d.brokerHosts, d.channel, d.channel+"_" + d.txn.tType + "_" + d.cmp.cType)
	if err != nil {
		return err
	}

	kafkaPublisher := kafka_manager.NewKafkaEventProducer(d.brokerHosts, d.channel)

	return nil
}


