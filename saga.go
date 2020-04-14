package go_saga_orchestrator

import (
	"github.com/aman-bansal/go_saga_orchestrator/internal/saga"
	"github.com/aman-bansal/go_saga_orchestrator/orchestrator"
	"github.com/aman-bansal/go_saga_orchestrator/participant"
)

func GetSagaOrchestratorBuilder() orchestrator.SagaOrchestratorBuilder {
	return saga.NewDefaultSagaOrchestratorBuilder()
}

func GetSagaParticipant() participant.SagaParticipant {
	return saga.NewDefaultSagaParticipant()
}
