package temporal

import (
	"context"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"
)

type (
	Temporal struct {
		BaseClient            client.Client
		WorkerClient          worker.Worker
		WorkflowServiceClient workflowservice.WorkflowServiceClient
		NamespaceClient       client.NamespaceClient
	}
)

func New() (t *Temporal, err error) {
	t = &Temporal{}

	t.BaseClient, err = client.NewClient(
		client.Options{HostPort: "temporal-fe-dev.service.consul:7233", Namespace: "dummy"},
	)
	if err != nil {
		return nil, err
	}
	t.WorkerClient = worker.New(t.BaseClient, "TASK_QUEUE_dummy", worker.Options{})

	conn, err := grpc.Dial("temporal-fe-dev.service.consul:7233", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	t.WorkflowServiceClient = workflowservice.NewWorkflowServiceClient(conn)

	t.NamespaceClient, err = client.NewNamespaceClient(
		client.Options{HostPort: "temporal-fe-dev.service.consul:7233"},
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *Temporal) RegisterNamespace(namespace string, retention int) error {
	r := time.Duration(retention) * time.Hour * 24
	err := t.NamespaceClient.Register(context.Background(), &workflowservice.RegisterNamespaceRequest{
		Namespace:                        namespace,
		WorkflowExecutionRetentionPeriod: &r,
	})
	if err != nil && err.Error() != "Namespace already exists." {
		return err
	}
	return nil
}
