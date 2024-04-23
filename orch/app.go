package orch

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"

	"github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/workflow"
)

// Application is an initialized Orchestrator ready to run
type Application struct {
	ctx    context.Context
	logger *slog.Logger

	eventsStream chan Event
	runnersCh    chan *ActionRunner

	daprClient     client.Client
	daprWorker     *workflow.WorkflowWorker
	daprWorkerLock sync.RWMutex

	closers []func() error
}

func NewApplication(ctx context.Context, logger *slog.Logger, eb EventBus, daprClient client.Client) (*Application, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	logger = logger.WithGroup("dapr_actions.orch")

	daprWorker, err := workflow.NewWorker(workflow.WorkerWithDaprClient(daprClient))
	if err != nil {
		return nil, ErrorDapr{Ctx: context.TODO(), Msg: "initializing dapr workflow worker", Err: err}
	} // errors below this point need to call daprWorker.Shutdown

	eventsStream := make(chan Event)
	go func() {
		for e := range eb.Incoming(ctx) {
			eventsStream <- e
		}
	}()

	runnersCh := make(chan *ActionRunner)

	closers := []func() error{
		daprWorker.Shutdown,
		func() error { close(eventsStream); return nil },
		func() error { close(runnersCh); return nil },
	}

	app := &Application{
		logger:       logger,
		daprClient:   daprClient,
		daprWorker:   daprWorker,
		eventsStream: eventsStream,
		runnersCh:    runnersCh,
		closers:      closers,
	}
	return app, nil
}

// EventBus is used to to read incoming Events.
type EventBus interface {
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
}

func (app *Application) Run(ctx context.Context) error {
	app.daprWorkerLock.Lock()
	defer app.daprWorkerLock.Unlock()
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
		app.handleErr(err)
		return
	}

	for _, action := range actions {
		runner, err := app.createRunnerForAction(ctx, action)
		if err != nil {
			app.handleErr(err)
			continue
		}

		app.queueRunner(runner)
	}
}

// TODO
func (app *Application) ActionsForEvent(ctx context.Context, event Event) ([]Action, error) {
	var actions []Action

	// TODO check a registry for matching actions for this event

	// TODO loop results
	action := Action{
		app:          app,
		Ctx:          event.Ctx,
		ID:           "0", // TODO index of loop, as string
		EventID:      event.ID,
		WorkflowName: "BuildDaprActions",
	}

	actions = append(actions, action)
	//

	return actions, nil
}

func (app *Application) createRunnerForAction(ctx context.Context, action Action) (*ActionRunner, error) {
	return nil, nil
}

func (app *Application) RegisterDaprWorkflow(wf workflow.Workflow) {
	app.daprWorkerLock.Lock()
	defer app.daprWorkerLock.Unlock()
	app.daprWorker.RegisterWorkflow(wf)
}

func (app *Application) queueRunner(runner *ActionRunner) {
	go func() {
		app.runnersCh <- runner
	}()
}

func (app *Application) handleErr(err error) {
	panic(err)
}
