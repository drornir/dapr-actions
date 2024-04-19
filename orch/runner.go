package orch

import (
	"context"

	"github.com/dapr/go-sdk/client"
)

type ActionRunner struct {
	action *Action
}

func (ar *ActionRunner) createWorkflow(ctx context.Context) error {
	ar.daprClient()

}

func (ar *ActionRunner) daprClient() client.Client {
	return ar.action.app.daprClient
}
