package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	cadence "github.com/ContextLogic/cadence/pkg"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	namespace = "cadence_lib"
	taskQueue = "TASK_QUEUE_cadence_lib"
)

var (
	client cadence.Client
	dummy  = &DummyWorkflow{&DummyActivity{}}
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

func (w *DummyWorkflow) DummyWorkflow(ctx workflow.Context) (interface{}, error) {
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

	var err error
	client, err = cadence.NewClient(
		cadence.ClientOptions{
			HostPort:  "temporal-fe-dev.service.consul:7233",
			Namespace: namespace,
		},
		cadence.WorkerOptions{
			MaxConcurrentActivityTaskPollers: 1,
			MaxConcurrentWorkflowTaskPollers: 1,
		},
		taskQueue,
	)
	if err != nil {
		panic(err)
	}

	client.Register(dummy.DummyWorkflow, []interface{}{dummy.Activity.DummyActivityCreateOrder, dummy.Activity.DummyActivityApprovePayment})

	http.HandleFunc("/start", Start)
	http.ListenAndServe(":8080", nil)
}

func Start(w http.ResponseWriter, req *http.Request) {
	we, err := client.ExecuteWorkflow(
		context.Background(),
		cadence.StartWorkflowOptions{
			ID:        strings.Join([]string{namespace, strconv.Itoa(int(time.Now().Unix()))}, "_"),
			TaskQueue: taskQueue,
		},
		dummy.DummyWorkflow,
	)
	if err != nil {
		panic(err)
	}

	response := ""
	err = we.Get(context.Background(), &response)
	if err != nil {
		panic(err)
	}
}
