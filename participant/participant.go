package participant

type SagaParticipant interface {
	Process() error
}

type SagaParticipantBuilder interface {
	listenTo(channel string) SagaParticipantBuilder
	AddTransactionProcessor(transactionType string, processor Processor) SagaParticipantBuilder
	AddCompensationProcessor(compensationType string, processor Processor) SagaParticipantBuilder
	Build() error
}

type Processor interface {
	Process() error
}