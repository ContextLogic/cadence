package temporal

import (
	"github.com/ContextLogic/cadence/pkg/options"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"
)

type Temporal struct {
	BaseClient            client.Client
	WorkerClient          worker.Worker
	WorkflowServiceClient workflowservice.WorkflowServiceClient
}

func New(options options.Options) (t *Temporal, err error) {
	t = &Temporal{}
	t.BaseClient, err = client.NewClient(client.Options{
		HostPort:  options.ClientOptions.HostPort,
		Namespace: options.ClientOptions.Namespace,
	})
	if err != nil {
		return nil, err
	}
	t.WorkerClient = worker.New(t.BaseClient, options.TaskQueue, options.WorkerOptions)

	conn, err := grpc.Dial(options.ClientOptions.HostPort, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	t.WorkflowServiceClient = workflowservice.NewWorkflowServiceClient(conn)
	return t, nil
}
