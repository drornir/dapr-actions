package orch

import (
	"context"
	"fmt"
	"sync"
)

type ActionTemplate struct {
	Name         string
	EventMatcher func(Event) bool
	WorkflowName string
}

// ActionsIndex is a database for querying Actions
type ActionsIndex struct {
	templates []ActionTemplate
	mux       sync.RWMutex
}

func (ai *ActionsIndex) ForEvent(ctx context.Context, event Event) ([]Action, error) {
	var actions []Action
	ai.mux.RLock()
	defer ai.mux.RUnlock()

	for _, template := range ai.templates {
		if template.EventMatcher(event) {
			action := Action{
				Ctx:          event.Ctx,
				ID:           fmt.Sprintf("%s/%s", template.Name, event.ID),
				EventID:      event.ID,
				WorkflowName: template.WorkflowName,
				TemplateName: template.Name,
			}
			actions = append(actions, action)
		}
	}

	return actions, nil
}

func (ai *ActionsIndex) RegisterActionTemplate(t ActionTemplate) {
	ai.mux.Lock()
	defer ai.mux.Unlock()
	// TODO make sure name doesn't exist? seems redundant tbh
	ai.templates = append(ai.templates, t)
}
