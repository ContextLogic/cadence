package cadence

import (
	"context"

	"go.temporal.io/sdk/client"
)

type (
	Client interface {
		Register(interface{}, []interface{})
		ExecuteWorkflow(context.Context, StartWorkflowOptions, interface{}, ...interface{}) (client.WorkflowRun, error)
	}
	clientImpl struct {
		register Register
		temporal *Temporal
	}
)

func NewClient(c ClientOptions, w WorkerOptions, queue string) (Client, error) {
	temporal, err := NewTemporalClient(c, w, queue)
	if err != nil {
		return nil, err
	}
	return &clientImpl{
		register: NewRegister(temporal),
		temporal: temporal,
	}, nil
}

func (c *clientImpl) Register(w interface{}, a []interface{}) {
	c.register.RegisterWorkflow(w)
	c.register.RegisterActivity(a)
	c.register.RegisterWorker()
}

func (c *clientImpl) ExecuteWorkflow(ctx context.Context, options StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	return c.temporal.BaseClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions(options), workflow)
}
