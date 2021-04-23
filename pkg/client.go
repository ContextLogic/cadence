package client

import (
	"context"
	"fmt"

	"github.com/ContextLogic/cadence/pkg/fsm/workflow"
	"github.com/ContextLogic/cadence/pkg/models"
	"github.com/ContextLogic/cadence/pkg/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type (
	Client interface {
		Register(*workflow.Workflow, models.ActivityMap) error
		ExecuteWorkflow(context.Context, client.StartWorkflowOptions, interface{}, interface{}) (client.WorkflowRun, error)
	}
	clientImpl struct {
		temporal *temporal.Temporal
		workers  map[string]worker.Worker
	}
)

func New(client client.Options, workers map[string]worker.Options) (Client, error) {
	t, err := temporal.New(client)
	if err != nil {
		return nil, err
	}
	c := &clientImpl{
		temporal: t,
		workers:  map[string]worker.Worker{},
	}
	for q, opt := range workers {
		c.workers[q] = worker.New(t.BaseClient, string(q), opt)
	}
	return c, nil
}

func (c *clientImpl) Register(w *workflow.Workflow, activities models.ActivityMap) error {
	wk, ok := c.workers[w.Queue]
	if !ok {
		return fmt.Errorf(fmt.Sprintf("No worker registered for workflow queue: %s", w.Queue))
	}

	w.RegisterWorkflow(wk)
	w.RegisterActivities(activities, wk)
	w.RegisterTaskHandlers(activities)
	go wk.Run(worker.InterruptCh())

	return nil
}

func (c *clientImpl) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, input interface{}) (client.WorkflowRun, error) {
	return c.temporal.BaseClient.ExecuteWorkflow(ctx, options, workflow, input)
}
