package client

import (
	"context"
	"errors"

	"github.com/ContextLogic/cadence/pkg/fsm/workflow"
	"github.com/ContextLogic/cadence/pkg/models"
	"github.com/ContextLogic/cadence/pkg/options"
	"github.com/ContextLogic/cadence/pkg/temporal"
	"go.temporal.io/sdk/client"
)

type (
	Client interface {
		Register(*workflow.Workflow, models.ActivityMap)
		ExecuteWorkflow(context.Context, client.StartWorkflowOptions, interface{}, ...interface{}) (client.WorkflowRun, error)
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

func (c *clientImpl) Register(w *workflow.Workflow, activities models.ActivityMap) {
	w.RegisterWorkflow(c.temporal.WorkerClient)
	w.RegisterActivities(activities, c.temporal.WorkerClient)
	w.RegisterWorker(c.temporal.WorkerClient)
	w.RegisterTaskHandlers(activities)
}

func (c *clientImpl) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	if len(args) != 1 {
		return nil, errors.New("wrong number of workflow params")
	}
	return c.temporal.BaseClient.ExecuteWorkflow(ctx, options, workflow, args[0])
}
