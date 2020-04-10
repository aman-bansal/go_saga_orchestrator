package event

type EventState string

const (
	EVENT EventState = ""

	TRANSACTION_START = "TRANSACTION_START"
	TRANSACTION_COMPLETE = "TRANSACTION_COMPLETE"
	TRANSACTION_FAIL = "TRANSACTION_FAIL"

	COMPENSATION_START = "COMPENSATION_START"
	COMPENSATION_COMPLETE = "COMPENSATION_COMPLETE"
	COMPENSATION_FAIL = "COMPENSATION_FAIL"
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


