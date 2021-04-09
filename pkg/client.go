package client

import (
	"context"

	"github.com/ContextLogic/cadence/pkg/fsm/workflow"
	"github.com/ContextLogic/cadence/pkg/models"
	"github.com/ContextLogic/cadence/pkg/options"
	"github.com/ContextLogic/cadence/pkg/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	sdk "go.temporal.io/sdk/workflow"
)

type (
	Client interface {
		RegisterDSLBased(*workflow.Workflow, models.ActivityMap)
		RegisterCodeBased(string, interface{}, []interface{})
		ExecuteWorkflow(context.Context, client.StartWorkflowOptions, interface{}, interface{}) (client.WorkflowRun, error)
	}
	clientImpl struct {
		temporal *temporal.Temporal
	}
)

func New(options options.Options) (Client, error) {
	t, err := temporal.New(options)
	if err != nil {
		return nil, err
	}
	return &clientImpl{
		temporal: t,
	}, nil
}

func (c *clientImpl) RegisterDSLBased(w *workflow.Workflow, activities models.ActivityMap) {
	w.RegisterWorkflow(c.temporal.WorkerClient)
	w.RegisterActivities(activities, c.temporal.WorkerClient)
	w.RegisterWorker(c.temporal.WorkerClient)
	w.RegisterTaskHandlers(activities)
}

func (c *clientImpl) RegisterCodeBased(name string, w interface{}, activities []interface{}) {
	c.temporal.WorkerClient.RegisterWorkflowWithOptions(w, sdk.RegisterOptions{Name: name})
	for _, a := range activities {
		c.temporal.WorkerClient.RegisterActivity(a)
	}
	go c.temporal.WorkerClient.Run(worker.InterruptCh())
}

func (c *clientImpl) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, input interface{}) (client.WorkflowRun, error) {
	return c.temporal.BaseClient.ExecuteWorkflow(ctx, options, workflow, input)
}
