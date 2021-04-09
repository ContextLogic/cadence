package options

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type (
	Options struct {
		ClientOptions client.Options
		WorkerOptions worker.Options
		TaskQueue     string
	}
)
