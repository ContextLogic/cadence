package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	dummy "github.com/ContextLogic/cadence/cmd/code_based"
	"github.com/ContextLogic/cadence/cmd/code_based/models"
	client "github.com/ContextLogic/cadence/pkg"
	"github.com/ContextLogic/cadence/pkg/fsm/workflow"
	"github.com/ContextLogic/cadence/pkg/options"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/activity"
	sdk "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var logger = logrus.New()

const (
	hostport         = "temporal-fe-dev.service.consul:7233"
	namespace        = "cadence_lib"
	taskqueue_prefix = "TASK_QUEUE_cadence_lib"
)

func init() {
	logger.SetFormatter(&logrus.JSONFormatter{})
}

func LoadWorkflows(path string) []*workflow.Workflow {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	value, _ := ioutil.ReadAll(f)
	m := []map[string]interface{}{}
	if err := json.Unmarshal(value, &m); err != nil {
		panic(err)
	}

	workflows := []*workflow.Workflow{}
	for _, v := range m {
		bytes, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		w, err := workflow.New(bytes)
		if err != nil {
			panic(err)
		}
		logger.WithFields(logrus.Fields{"workflow": w}).Info("workflows")
		workflows = append(workflows, w)
	}
	return workflows
}

func Activity1(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity executed")

	return input, nil
}

func Activity2(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity executed")

	return input, nil
}

func CreateClient(queue string) (client.Client, error) {
	return client.New(options.Options{
		ClientOptions: sdk.Options{
			HostPort:  hostport,
			Namespace: namespace,
		},
		WorkerOptions: worker.Options{}, // use temporal's default value
		TaskQueue:     strings.Join([]string{taskqueue_prefix, queue}, "_"),
	})
}

func RunDSLBasedWorkflows() {
	c, err := CreateClient("dsl")
	if err != nil {
		panic(err)
	}
	wfs := LoadWorkflows("./cmd/dsl_based/workflows.json")
	for idx, wf := range wfs {
		activityMap := map[string]func(context.Context, map[string]interface{}) (interface{}, error){
			"example:activity:Activity1": Activity1,
			"example:activity:Activity2": Activity2,
		}

		c.RegisterDSLBased(wf, activityMap)
		instance, err := c.ExecuteWorkflow(
			context.Background(),
			sdk.StartWorkflowOptions{
				ID:        strings.Join([]string{namespace, strconv.Itoa(int(time.Now().Unix()))}, "_"),
				TaskQueue: strings.Join([]string{taskqueue_prefix, "dsl"}, "_"),
			},
			wf.Name,
			map[string]interface{}{
				"foo": 4,
			},
		)
		if err != nil {
			panic(err)
		}

		if idx == 0 {
			response := map[string]interface{}{}
			err = instance.Get(context.Background(), &response)
		} else {
			response := []map[string]interface{}{}
			err = instance.Get(context.Background(), &response)
		}
		if err != nil {
			panic(err)
		}
	}
}

func RunCodeBasedWorkflows() {
	name := "example:workflow:CodeExampleWorkflow"
	c, err := CreateClient("code")
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}
	d := dummy.NewDummyWorkflow()
	c.RegisterCodeBased(
		name,
		d.DummyWorkflow,
		[]interface{}{
			d.Activities.DummyCreateOrder,
			d.Activities.DummyApprovePayment,
			d.Activities.DummyDeclineOrder,
		},
	)

	order := models.Order{ProductID: "toy", CustomerID: "lshu", ShippingAddress: "1 ave"}
	instance, err := c.ExecuteWorkflow(
		context.Background(),
		sdk.StartWorkflowOptions{
			ID:        strings.Join([]string{namespace, strconv.Itoa(int(time.Now().Unix()))}, "_"),
			TaskQueue: strings.Join([]string{taskqueue_prefix, "code"}, "_"),
		},
		name,
		order,
	)
	if err != nil {
		panic(err)
	}

	response := &models.OrderResponse{}
	err = instance.Get(context.Background(), &response)
	if err != nil {
		panic(err)
	}
}

func main() {
	RunDSLBasedWorkflows()
	RunCodeBasedWorkflows()
}
