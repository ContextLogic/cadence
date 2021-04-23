package client

import (
	"context"

	"github.com/ContextLogic/cadence/pkg/models"
	"github.com/ContextLogic/cadence/pkg/temporal"
	resolver "github.com/ContextLogic/cadence/pkg/utils/workflow_resolver"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type (
	Client interface {
		Register(interface{}, map[string]worker.Options, models.ActivityMap) error
		ExecuteWorkflow(context.Context, client.StartWorkflowOptions, interface{}, interface{}) (client.WorkflowRun, error)
	}
	clientImpl struct {
		temporal *temporal.Temporal
		workers  map[string]worker.Worker
		status   map[string]models.WorkerStatus
	}
)

func New(client client.Options) (Client, error) {
	t, err := temporal.New(client)
	if err != nil {
		return nil, err
	}
	return &clientImpl{
		temporal: t,
		workers:  make(map[string]worker.Worker),
		status:   make(map[string]models.WorkerStatus),
	}, nil
}

func (c *clientImpl) Register(w interface{}, workers map[string]worker.Options, activities models.ActivityMap) error {
	workflows, err := resolver.ResolveWorkflow(w)
	if err != nil {
		return err
	}
	for _, wf := range workflows {
		opt, ok := workers[wf.Queue]
		if !ok {
			// using default worker options
			opt = worker.Options{}
		}
		if _, ok := c.workers[wf.Queue]; !ok {
			c.workers[wf.Queue] = worker.New(c.temporal.BaseClient, wf.Queue, opt)
			c.status[wf.Queue] = models.WorkerInitializing
		}
		wf.RegisterWorkflow(c.workers[wf.Queue])
		wf.RegisterActivities(activities, c.workers[wf.Queue])
		wf.RegisterTaskHandlers(activities)

		if s, ok := c.status[wf.Queue]; ok {
			if s == models.WorkerInitializing {
				c.status[wf.Queue] = models.WorkerRunning
				go c.workers[wf.Queue].Run(worker.InterruptCh())
			}
		}
	}

	return nil
}

func (c *clientImpl) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, input interface{}) (client.WorkflowRun, error) {
	return c.temporal.BaseClient.ExecuteWorkflow(ctx, options, workflow, input)
}
