package event

type EventState string

const (
	TRANSACTION_START    = "TRANSACTION_START"
	TRANSACTION_COMPLETE = "TRANSACTION_COMPLETE"
	TRANSACTION_FAIL     = "TRANSACTION_FAIL"

	COMPENSATION_START    = "COMPENSATION_START"
	COMPENSATION_COMPLETE = "COMPENSATION_COMPLETE"
	COMPENSATION_FAIL     = "COMPENSATION_FAIL"
)

type KafkaEvent struct {
	SagaId    string     `json:"saga_id"`
	EventType string     `json:"event_type"`
	State     EventState `json:"state"`
	Data      []byte     `json:"data"`
}
