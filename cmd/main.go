package main

import (
	"context"
	"fmt"
	"time"

	cadence "github.com/ContextLogic/cadence/pkg"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type (
	DummyWorkflow struct {
		Activity *DummyActivity
	}
	DummyActivity struct{}
)

func (w *DummyActivity) DummyActivityCreateOrder(ctx context.Context) (string, error) {
	fmt.Println("create order")
	return "ok", nil
}

func (w *DummyActivity) DummyActivityApprovePayment(ctx context.Context) (string, error) {
	fmt.Println("approve payment")
	return "ok", nil
}

func (w *DummyWorkflow) DummyWorkflowEntry(ctx workflow.Context) (interface{}, error) {
	var response string
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * time.Duration(1),
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * time.Duration(1),
			BackoffCoefficient: 1.0,
			MaximumInterval:    time.Second * time.Duration(1),
			MaximumAttempts:    10,
		},
	})

	err := workflow.ExecuteActivity(ctx, w.Activity.DummyActivityCreateOrder).Get(ctx, &response)
	if err != nil {
		return nil, err
	}
	fmt.Println(response)
	err = workflow.ExecuteActivity(ctx, w.Activity.DummyActivityApprovePayment).Get(ctx, &response)
	if err != nil {
		return nil, err
	}
	fmt.Println(response)
	return response, nil
}

func main() {
	dummyw := &DummyWorkflow{}
	dummya := &DummyActivity{}
	err := cadence.RegisterNamespace("dummy", cadence.RegisterOptions{Retension: 1})
	if err != nil {
		panic(err)
	}
	cadence.Register(dummyw, dummya)
}
