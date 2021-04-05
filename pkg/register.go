package cadence

import (
	"context"
	"fmt"
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
		RegisterNamespace(string, RegisterNamespaceOptions) error
	}
	registerImpl struct {
		temporal *Temporal
	}
)

func NewRegister(temporal *Temporal) Register {
	return &registerImpl{temporal}
}

func (r *registerImpl) RegisterWorkflow(w interface{}) {
	t := reflect.TypeOf(w)
	for i := 0; i < t.NumMethod(); i++ {
		fmt.Println(t.Method(i).Name)
		r.temporal.WorkerClient.RegisterWorkflow(t.Method(i).Func.Interface())
	}
}

func (r *registerImpl) RegisterActivity(a interface{}) {
	t := reflect.TypeOf(a)
	for i := 0; i < t.NumMethod(); i++ {
		r.temporal.WorkerClient.RegisterActivity(t.Method(i).Func.Interface())
	}
}

func (r *registerImpl) RegisterWorker() {
	go r.temporal.WorkerClient.Run(worker.InterruptCh())
}

func (r *registerImpl) RegisterNamespace(namespace string, options RegisterNamespaceOptions) error {
	retention := time.Duration(options.Retention) * time.Hour * 24
	err := r.temporal.NamespaceClient.Register(
		context.Background(),
		&workflowservice.RegisterNamespaceRequest{
			Namespace:                        namespace,
			WorkflowExecutionRetentionPeriod: &retention,
		},
	)
	if err != nil && err.Error() != "Namespace already exists." {
		return err
	}
	return nil
}
