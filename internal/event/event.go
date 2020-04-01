package event

type EventState string

const (
	EVENT EventState = ""

	TRANSACTION_COMPLETED = "TRANSACTION_COMPLETED"
	TRANSACTION_FAILED = "TRANSACTION_COMPLETED"

	COMPENSATION_COMPLETED = "COMPENSATION_COMPLETED"
	COMPENSATION_FAILED = "COMPENSATION_FAILED"
)

type Event interface {
	GetSagaId() int64
	GetEventType() string
	GetData() []byte
	GetState() EventState
}

type KafkaEvent struct {
	sagaId    int64
	eventType string
	state     EventState
	data      []byte
}

func (k *KafkaEvent) GetSagaId() int64 {
	return k.sagaId
}

func (k *KafkaEvent) GetEventType() string {
	return k.eventType
}

func (k *KafkaEvent) GetData() []byte {
	return k.data
}

func (k *KafkaEvent) GetState() EventState {
	return k.state
}


