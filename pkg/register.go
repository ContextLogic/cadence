package cadence

import (
	"context"
	"reflect"
	"time"

	"github.com/ContextLogic/cadence/pkg/clients"
	"go.temporal.io/api/workflowservice/v1"
)

func Register(w interface{}, a interface{}) {
	t := reflect.TypeOf(w)
	for i := 0; i < t.NumMethod(); i++ {
		clients.Temporal.WorkerClient.RegisterActivity(t.Method(i).Func.Interface())
	}

	t = reflect.TypeOf(a)
	for i := 0; i < t.NumMethod(); i++ {
		clients.Temporal.WorkerClient.RegisterActivity(t.Method(i).Func.Interface())
	}
}

func RegisterNamespace(namespace string) error {
	r := time.Duration(1) * time.Hour * 24
	err := clients.Temporal.NamespaceClient.Register(
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
