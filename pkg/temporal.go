package cadence

import (
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"
)

var Temporal Temporal

type Temporal struct {
	BaseClient            client.Client
	WorkerClient          worker.Worker
	WorkflowServiceClient workflowservice.WorkflowServiceClient
	NamespaceClient       client.NamespaceClient
}

func init(c ClientOptions, w WorkerOptions, queue string) (err error) {
	Temporalt.BaseClient, err = client.NewClient(
		client.Options{
			HostPort:  c.HostPort,
			Namespace: c.Namespace,
		},
	)
	if err != nil {
		return err
	}
	Temporal.WorkerClient = worker.New(t.BaseClient, queue, worker.Options{
		MaxConcurrentActivityTaskPollers: w.MaxConcurrentActivityTaskPollers,
		MaxConcurrentWorkflowTaskPollers: w.MaxConcurrentActivityTaskPollers,
	})

	conn, err := grpc.Dial(Options.ClientOptions.HostPort, grpc.WithInsecure())
	if err != nil {
		return err
	}
	Temporal.WorkflowServiceClient = workflowservice.NewWorkflowServiceClient(conn)

	Temporal.NamespaceClient, err = client.NewNamespaceClient(
		client.Options{HostPort: Options.ClientOptions.HostPort},
	)
	return err
}
