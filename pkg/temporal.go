package cadence

import (
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"
)

type Temporal struct {
	BaseClient            client.Client
	WorkerClient          worker.Worker
	WorkflowServiceClient workflowservice.WorkflowServiceClient
	NamespaceClient       client.NamespaceClient
}

func NewTemporalClient(c ClientOptions, w WorkerOptions, queue string) (t *Temporal, err error) {
	t = &Temporal{}
	t.BaseClient, err = client.NewClient(
		client.Options{
			HostPort:  c.HostPort,
			Namespace: c.Namespace,
		},
	)
	if err != nil {
		return nil, err
	}
	t.WorkerClient = worker.New(t.BaseClient, queue, worker.Options{
		MaxConcurrentActivityTaskPollers: w.MaxConcurrentActivityTaskPollers,
		MaxConcurrentWorkflowTaskPollers: w.MaxConcurrentActivityTaskPollers,
	})

	conn, err := grpc.Dial(c.HostPort, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	t.WorkflowServiceClient = workflowservice.NewWorkflowServiceClient(conn)

	t.NamespaceClient, err = client.NewNamespaceClient(
		client.Options{HostPort: c.HostPort},
	)
	if err != nil {
		return nil, err
	}

	return t, nil
}
