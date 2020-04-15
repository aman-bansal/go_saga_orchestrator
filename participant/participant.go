package participant

type SagaParticipantRegistration interface {
	WithKafkaConfig(brokerHosts []string) SagaParticipantRegistration
	ListenTo(channel string) SagaParticipantRegistration
	AddTransactionProcessor(transactionType string, processor Processor) SagaParticipantRegistration
	AddCompensationProcessor(compensationType string, processor Processor) SagaParticipantRegistration
	Register() error
}

type Processor interface {
	Process(data []byte) ([]byte, error)
}