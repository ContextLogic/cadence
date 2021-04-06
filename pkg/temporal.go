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
}

func NewTemporalClient(c ClientOptions, w WorkerOptions, queue string) (t *Temporal, err error) {
	t = &Temporal{}
	t.BaseClient, err = client.NewClient(client.Options(c))
	if err != nil {
		return nil, err
	}
	t.WorkerClient = worker.New(t.BaseClient, queue, worker.Options(w))

	conn, err := grpc.Dial(c.HostPort, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	t.WorkflowServiceClient = workflowservice.NewWorkflowServiceClient(conn)

	return t, nil
}
