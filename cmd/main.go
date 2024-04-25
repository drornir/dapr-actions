package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/workflow"

	"github.com/drornir/dapr-actions/orch"
)

var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	AddSource: false,
	Level:     slog.LevelDebug,
	ReplaceAttr: func(_groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	},
}))

func main() {
	ctx := context.Background()
	ctx, ctxCancel := signal.NotifyContext(ctx, os.Interrupt)
	defer ctxCancel()

	var daprClient client.Client
	if c, err := client.NewClient(); err != nil {
		fatal("creating dapr client: %w", err)
	} else {
		daprClient = c
	}
	defer daprClient.Close()

	inMemEventBus := &orch.InMemQueueBus{}
	defer inMemEventBus.Close()
	testEvent := orch.Event{
		Ctx: ctx,
		ID:  "0",
	}
	inMemEventBus.Emit(ctx, testEvent)

	orchestrator, err := orch.NewApplication(ctx, logger, inMemEventBus, daprClient)
	if err != nil {
		fatal("creating new orchestrator: %w", err)
	}

	orchestrator.RegisterDaprWorkflow(ExampleWorkflow)

	if err := orchestrator.Run(ctx); err != nil {
		fatal("orchestrator exited with error: %w", err)
	}
}

func fatal(f string, a ...any) {
	err := fmt.Errorf(f, a...)
	panic(err)
}

func ExampleWorkflow(wctx *workflow.WorkflowContext) (any, error) {
	i
}
