package temporal

import (
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"
)

var (
	BaseClient            client.Client
	WorkerClient          worker.Worker
	WorkflowServiceClient workflowservice.WorkflowServiceClient
)

func init() {
	var err error
	BaseClient, err = client.NewClient(client.Options{
		HostPort:  "temporal-fe-dev.service.consul:7233",
		Namespace: "cadence_lib",
	})
	if err != nil {
		panic(err)
	}
	WorkerClient = worker.New(BaseClient, "TASK_QUEUE_cadence_lib", worker.Options{})

	conn, err := grpc.Dial("temporal-fe-dev.service.consul:7233", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	WorkflowServiceClient = workflowservice.NewWorkflowServiceClient(conn)
}
