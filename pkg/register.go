package cadence

import (
	"context"
	"reflect"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/worker"
)

type (
	Register interface {
		RegisterWorkflow(interface{})
		RegisterActivity(interface{})
		RegisterWorker()
		RegisterNamespace(string) error
	}
	registerImpl struct{}
)

func NewRegister() Register {
	return registerImpl{}
}

func (r *registerImpl) RegisterWorkflow(w interface{}) {
	t := reflect.TypeOf(w)
	for i := 0; i < t.NumMethod(); i++ {
		Temporal.WorkerClient.RegisterWorkflow(t.Method(i).Func.Interface())
	}
}

func (r *registerImpl) RegisterActivity(a interface{}) {
	t = reflect.TypeOf(a)
	for i := 0; i < t.NumMethod(); i++ {
		Temporal.WorkerClient.RegisterActivity(t.Method(i).Func.Interface())
	}
}

func (r *registerImpl) RegisterWorker() {
	go Temporal.WorkerClient.Run(worker.InterruptCh())
}

func (r *registerImpl) RegisterNamespace(namespace string, options RegisterOptions) error {
	r := time.Duration(options.Retention) * time.Hour * 24
	err := Temporal.NamespaceClient.Register(
		context.Background(),
		&workflowservice.RegisterNamespaceRequest{
			Namespace:                        namespace,
			WorkflowExecutionRetentionPeriod: &r,
		},
	)
	if err != nil && err.Error() != "Namespace already exists." {
		return err
	}
	return nil
}
