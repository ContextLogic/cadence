package cadence

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type (
	// ClientOptions are optional parameters for temporal Client creation
	// https://github.com/temporalio/sdk-go/blob/master/internal/client.go#L325
	ClientOptions client.Options

	// Options is used to configure a worker instance
	// https://github.com/temporalio/sdk-go/blob/master/internal/worker.go#L36
	WorkerOptions worker.Options

	ActivityOptions          workflow.ActivityOptions
	RegisterNamespaceOptions struct {
		Retention int
	}
)
