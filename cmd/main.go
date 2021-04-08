package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ContextLogic/cadence/pkg/fsm/workflow"
	"github.com/ContextLogic/cadence/pkg/temporal"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

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

func main() {
	wfs := LoadWorkflows("./cmd/workflows.json")
	for _, wf := range wfs {
		activityMap := map[string]func(context.Context, interface{}) (interface{}, error){
			"example:activity:Activity1": Activity1,
			"example:activity:Activity2": Activity2,
		}

		wf.RegisterWorkflow(wf.Name)
		wf.RegisterActivities(activityMap)
		wf.RegisterWorker()
		wf.RegisterTaskHandlers(activityMap)

		instance, err := temporal.BaseClient.ExecuteWorkflow(
			context.Background(),
			client.StartWorkflowOptions{
				ID:        strings.Join([]string{"cadence_lib", strconv.Itoa(int(time.Now().Unix()))}, "_"),
				TaskQueue: "TASK_QUEUE_cadence_lib",
			},
			wf.Name,
			map[string]interface{}{
				"foo": 3,
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
}

func Activity1(ctx context.Context, input interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity executed")

	return input, nil
}

func Activity2(ctx context.Context, input interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity executed")

	return input, nil
}
