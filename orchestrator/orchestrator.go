package orchestrator

type SagaOrchestrator interface {
	Start(data []byte) error
}

type SagaOrchestratorBuilder interface {
	WithKafkaConfig(brokerHosts []string)
	WithMysqlConfig(host string, username string, password string)
	SetSagaId(sagaId string) SagaOrchestratorBuilder
	SetApplicationOwner(owner string) SagaOrchestratorBuilder
	PublishTo(channel string) SagaOrchestratorBuilder //kafka channel name. should be already created
	Name(name string) SagaOrchestratorBuilder
	Add(transaction Transaction, compensation Compensation, deadline int) SagaOrchestratorBuilder
	Build() (SagaOrchestrator, error)
}

type Transaction interface {
	RunTransaction(data []byte) ([]byte, error)
	TransactionKey() string //should be unique within a saga
}

type Compensation interface {
	RunCompensation(data []byte) ([]byte, error)
	CompensationKey() string //should be unique within a saga
}