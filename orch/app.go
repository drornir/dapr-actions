package orch

import "context"

// Application is an initialized Orchestrator ready to run
type Application struct {
	EventBus
}

// EventBus is used to to read incoming Events.
type EventBus interface {
	Incoming(context.Context) <-chan Event
	Close()
}
type Event struct {
	Ctx context.Context
	ID  string
}

type Action struct {
	eventID string
}

func (app *Application) Run(ctx context.Context) error {
	eventsStream := app.EventBus.Incoming(ctx)

	for {
		select {
		case <-ctx.Done():
			app.Close()
			return ctx.Err()

		case event := <-eventsStream:
			app.HandleEvent(ctx, event)
		}
	}
}

func (app *Application) HandleEvent(ctx context.Context, event Event) {
	// TODO wrap in async runner

	actions, err := app.ActionsForEvent(ctx, event)
	if err != nil {
		app.handleErr(err)
		return
	}

	for action := range actions {
		worker, err := app.CreateWorkerForAction(ctx, action)
		if err != nil {
			app.handleErr(err)
			return
		}

	}

}

func (app *Application) ActionsForEvent(<-chan Action, error) {
	panic("uninmplemented")
}

func (app *Application) handleErr(err error) {

}

func (app *Application) Close() {
	app.EventBus.Close()
	return
}
