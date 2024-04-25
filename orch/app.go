package orch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/workflow"
)

const DEFAULT_CHANNEL_BUFFER_SIZE = 1

// Application is an initialized Orchestrator ready to run
type Application struct {
	ctx    context.Context
	logger *slog.Logger

	eventsStream <-chan Event
	actionsIndex *ActionsIndex
	runnersCh    chan *ActionRunner

	daprClient client.Client
	daprWorker *workflow.WorkflowWorker

	errsChan chan error
	closers  []func() error
}

func NewApplication(ctx context.Context, logger *slog.Logger, eb EventBusReader, daprClient client.Client) (*Application, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	logger = logger.WithGroup("dapr_actions.orch")

	daprWorker, err := workflow.NewWorker(workflow.WorkerWithDaprClient(daprClient))
	if err != nil {
		return nil, ErrorDapr{Ctx: ctx, Msg: "initializing dapr workflow worker", Err: err}
	} // errors below this point need to call daprWorker.Shutdown

	eventsStream := make(chan Event, DEFAULT_CHANNEL_BUFFER_SIZE)
	go func() {
		for e := range eb.Incoming(ctx) {
			eventsStream <- e
		}
	}()

	runnersCh := make(chan *ActionRunner, DEFAULT_CHANNEL_BUFFER_SIZE)
	errsChan := make(chan error, DEFAULT_CHANNEL_BUFFER_SIZE)

	closers := []func() error{
		daprWorker.Shutdown,
		func() error {
			close(eventsStream)
			close(runnersCh)
			close(errsChan)
			return nil
		},
	}

	app := &Application{
		logger:       logger,
		daprClient:   daprClient,
		daprWorker:   daprWorker,
		eventsStream: eventsStream,
		actionsIndex: &ActionsIndex{ /* TODO */ },
		runnersCh:    runnersCh,
		closers:      closers,
		errsChan:     errsChan,
	}
	return app, nil
}

// EventBusReader is used to to read incoming Events.
type EventBusReader interface {
	Incoming(context.Context) <-chan Event
}

type Event struct {
	Ctx context.Context
	ID  string
}

type Action struct {
	app *Application

	Ctx     context.Context
	ID      string
	EventID string

	WorkflowName string
	TemplateName string
}

func (app *Application) Run(ctx context.Context) error {
	if err := app.daprWorker.Start(); err != nil {
		return ErrorDapr{
			Ctx: ctx,
			Msg: "starting dapr worker",
			Err: err,
		}
	}

	for {
		select {
		case <-ctx.Done():
			app.Close()
			return ctx.Err()

		case event := <-app.eventsStream:
			app.HandleEvent(ctx, event)
		}
	}
}

func (app *Application) Close() error {
	var errs []error
	for _, closer := range app.closers {
		errs = append(errs, closer())
	}

	return errors.Join(errs...)
}

func (app *Application) HandleEvent(ctx context.Context, event Event) {
	// TODO wrap in async runner

	actions, err := app.ActionsForEvent(ctx, event)
	if err != nil {
		app.handleErr("getting actions for event: %w", err)
		return
	}

	for _, action := range actions {
		runner, err := app.createRunnerForAction(action)
		if err != nil {
			app.handleErr("creating runner for action %v: %w", action, err)
			continue
		}

		app.queueRunner(runner)
	}
}

func (app *Application) ActionsForEvent(ctx context.Context, event Event) ([]Action, error) {
	actions, err := app.actionsIndex.ForEvent(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("getting actions for event: %w", err)
	}

	for i := range actions {
		actions[i].app = app
	}

	return actions, nil
}

func (app *Application) createRunnerForAction(action Action) (*ActionRunner, error) {
	runner := &ActionRunner{
		action: &action,
	}

	return runner, nil
}

func (app *Application) RegisterAction(template ActionTemplate) {
	app.actionsIndex.RegisterActionTemplate(template)
}

func (app *Application) RegisterDaprWorkflow(wf workflow.Workflow) error {
	return app.daprWorker.RegisterWorkflow(wf)
}

func (app *Application) queueRunner(runner *ActionRunner) {
	go func() {
		app.runnersCh <- runner
	}()
}

func (app *Application) handleErr(f string, a ...any) {
	err := fmt.Errorf(f, a...)
	app.errsChan <- err
}
