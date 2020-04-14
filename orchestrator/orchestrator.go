package orchestrator

import "github.com/aman-bansal/go_saga_orchestrator/internal/saga"

type SagaOrchestratorRegistry interface {
	Start(sagaId string, data []byte)	error
}

type SagaOrchestratorRegistration interface {
	WithKafkaConfig(brokerHosts []string) SagaOrchestratorRegistration
	WithMysqlConfig(host string, username string, password string, dbName string) SagaOrchestratorRegistration
	PublishTo(channel string) SagaOrchestratorRegistration //kafka channel name. should be already created
	AddSagaOrchestrator(saga.SagaOrchestrator) SagaOrchestratorRegistration
	Register() (SagaOrchestratorRegistry, error)
}

type SagaOrchestratorBuilder interface {
	SetSagaId(sagaId string) SagaOrchestratorBuilder
	Name(name string) SagaOrchestratorBuilder
	Add(transaction Transaction, compensation Compensation, deadline int) SagaOrchestratorBuilder
	Build() saga.SagaOrchestrator
}

type Transaction interface {
	RunTransaction(data []byte) ([]byte, error)
	TransactionKey() string //should be unique within a saga
}

type Compensation interface {
	RunCompensation(data []byte) ([]byte, error)
	CompensationKey() string //should be unique within a saga
}