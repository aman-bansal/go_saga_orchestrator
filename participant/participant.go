package participant

type SagaParticipant interface {
	WithKafkaConfig(brokerHosts []string) SagaParticipant
	WithMysqlConfig(host string, username string, password string) SagaParticipant
	ListenTo(channel string) SagaParticipant
	AddTransactionProcessor(transactionType string, processor Processor) SagaParticipant
	AddCompensationProcessor(compensationType string, processor Processor) SagaParticipant
	Start() error
}

type Processor interface {
	Process(data []byte) error
}