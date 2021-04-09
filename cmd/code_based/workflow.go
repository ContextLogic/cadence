package dummy

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/ContextLogic/cadence/cmd/code_based/models"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type (
	DummyWorkflow struct {
		Activities *DummyActivities
	}
	DummyActivities struct {
		Root string
	}
)

func NewDummyWorkflow() *DummyWorkflow {
	_, filename, _, _ := runtime.Caller(0)
	return &DummyWorkflow{
		Activities: &DummyActivities{
			Root: path.Join(path.Dir(filename), "."),
		},
	}
}

func (a *DummyActivities) ReadProfile(path string) (map[string]string, error) {
	profile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer profile.Close()

	value, _ := ioutil.ReadAll(profile)
	results := make(map[string]string)
	json.Unmarshal([]byte(value), &results)
	return results, nil
}

func (a *DummyActivities) DummyCreateOrder(ctx context.Context, order models.Order) (*models.OrderResponse, error) {
	profile, err := a.ReadProfile(path.Join(a.Root, "flags/order.json"))
	if err != nil {
		return nil, err
	}

	switch profile["status"] {
	case "valid":
		return &models.OrderResponse{
			Order:  &order,
			Status: "succeeded",
		}, nil
	case "invalid":
		return &models.OrderResponse{
			Order:  &order,
			Status: "invalid_order",
		}, nil
	default:
		return nil, errors.New("failed to create order")
	}
}

func (a *DummyActivities) DummyApprovePayment(ctx context.Context, order models.Order) (*models.OrderResponse, error) {
	profile, err := a.ReadProfile(path.Join(a.Root, "flags/payment.json"))
	if err != nil {
		return nil, err
	}

	switch profile["status"] {
	case "valid":
		return &models.OrderResponse{
			Order:  &order,
			Status: "succeeded",
		}, nil
	case "invalid":
		return &models.OrderResponse{
			Order:  &order,
			Status: "invalid_payment",
		}, nil
	default:
		return nil, errors.New("failed to process payment")
	}
}

func (a *DummyActivities) DummyDeclineOrder(ctx context.Context, order models.Order) (*models.OrderResponse, error) {
	return &models.OrderResponse{
		Order:  &order,
		Status: "order is declined",
	}, nil
}

func (w *DummyWorkflow) DummyWorkflow(ctx workflow.Context, input interface{}) (interface{}, error) {
	data, _ := json.Marshal(input)
	order := models.Order{}
	err := json.Unmarshal(data, &order)
	if err != nil {
		return nil, err
	}

	response := models.OrderResponse{}
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * time.Duration(1),
			BackoffCoefficient: 1.0,
			MaximumInterval:    time.Second * time.Duration(1),
			MaximumAttempts:    10,
		},
	})

	err = workflow.ExecuteActivity(ctx, w.Activities.DummyCreateOrder, order).Get(ctx, &response)
	if err != nil {
		return nil, err
	}
	if response.Status != "succeeded" {
		workflow.ExecuteActivity(ctx, w.Activities.DummyDeclineOrder, order).Get(ctx, nil)
		return response, nil
	}

	err = workflow.ExecuteActivity(ctx, w.Activities.DummyApprovePayment, order).Get(ctx, &response)
	if err != nil {
		return nil, err
	}
	if response.Status != "succeeded" {
		workflow.ExecuteActivity(ctx, w.Activities.DummyDeclineOrder, order).Get(ctx, nil)
		return response, nil
	}
	return response, nil
}
