package participant

type SagaParticipantRegistration interface {
	WithKafkaConfig(brokerHosts []string) SagaParticipantRegistration
	ListenTo(channel string) SagaParticipantRegistration
	AddTransactionProcessor(transactionKey string, processor Processor) SagaParticipantRegistration
	AddCompensationProcessor(compensationKey string, processor Processor) SagaParticipantRegistration
	Register() error
}

type Processor interface {
	Process(data []byte) ([]byte, error)
}