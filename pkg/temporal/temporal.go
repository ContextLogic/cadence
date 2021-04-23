package temporal

import (
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
)

type Temporal struct {
	BaseClient            client.Client
	WorkflowServiceClient workflowservice.WorkflowServiceClient
}

func New(options client.Options) (t *Temporal, err error) {
	t = &Temporal{}
	t.BaseClient, err = client.NewClient(options)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(options.HostPort, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	t.WorkflowServiceClient = workflowservice.NewWorkflowServiceClient(conn)
	return t, nil
}
