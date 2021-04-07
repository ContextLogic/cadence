package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	wf "github.com/ContextLogic/cadence/pkg/fsm/workflow"
	"github.com/ContextLogic/cadence/pkg/temporal"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func init() {
	logger.SetFormatter(&logrus.JSONFormatter{})
}

type Workflow struct {
	name string
	json []byte
}

func getWorkflow(path string) Workflow {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	value, _ := ioutil.ReadAll(f)
	return Workflow{
		name: "example:workflow:ExampleWorkflow",
		json: value,
	}
}

func main() {
	e := getWorkflow("./cmd/workflow.json")
	w, err := wf.New(e.json)
	if err != nil {
		panic(fmt.Errorf("error loading workflow %w", err))
	}
	logger.WithFields(logrus.Fields{"workflow": w}).Info("workflows")

	activityMap := map[string]func(context.Context, interface{}) (interface{}, error){
		"example:activity:Activity1": WorkflowActivity1,
		"example:activity:Activity2": WorkflowActivity2,
	}

	w.RegisterWorkflow(e.name)
	w.RegisterActivities(activityMap)
	w.RegisterWorker()
	w.RegisterTaskHandlers(activityMap)

	instance, err := temporal.BaseClient.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			ID:        strings.Join([]string{"cadence_lib", strconv.Itoa(int(time.Now().Unix()))}, "_"),
			TaskQueue: "TASK_QUEUE_cadence_lib",
		},
		e.name,
		map[string]interface{}{
			"hello": "world",
		},
	)
	if err != nil {
		panic(err)
	}
	response := map[string]interface{}{}
	err = instance.Get(context.Background(), &response)
	if err != nil {
		panic(err)
	}
}

func WorkflowActivity1(ctx context.Context, input interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity executed")

	return input, nil
}

func WorkflowActivity2(ctx context.Context, input interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity executed")

	return input, nil
}
