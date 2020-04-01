package go_saga_orchestrator

import (
	"github.com/aman-bansal/go_saga_orchestrator/internal/saga"
	"github.com/aman-bansal/go_saga_orchestrator/orchestrator"
)

func GetSagaOrchestratorBuilder() orchestrator.SagaOrchestratorBuilder {
	return saga.NewDefaultSagaOrchestratorBuilder()
}

func GetSagaParticipant() {

}
