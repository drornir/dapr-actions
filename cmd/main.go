package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/dapr/go-sdk/client"

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

	var eventBus orch.EventBus

	orchestrator, err := orch.NewApplication(ctx, logger, eventBus, daprClient)
	if err != nil {
		fatal("creating new orchestrator: %w", err)
	}

	if err := orchestrator.Run(ctx); err != nil {
		fatal("orchestrator exited with error: %w", err)
	}
}

func fatal(f string, a ...any) {
	err := fmt.Errorf(f, a...)
	panic(err)
}
