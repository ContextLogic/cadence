package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

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

func getWorkflows(path string) []*Workflow {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	value, _ := ioutil.ReadAll(f)
	m := map[string]interface{}{}
	if err := json.Unmarshal(value, &m); err != nil {
		panic(err)
	}
	workflows := []*Workflow{}
	for k, v := range m {
		bytes, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		workflows = append(workflows, &Workflow{
			name: k,
			json: bytes,
		})
	}
	return workflows
}

func main() {
	e := getWorkflow("./cmd/workflow.json")
	w, err := wf.New(e.json)
	if err != nil {
		panic(fmt.Errorf("error loading workflow %w", err))
	}
	logger.WithFields(logrus.Fields{"workflow": w}).Info("workflows")

	activityMap := map[string]func(context.Context, interface{}) (interface{}, error){
		"example:activity:Activity1": Activity1,
		"example:activity:Activity2": Activity2,
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
			"foo": 1,
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

func Activity1(ctx context.Context, input interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity1 executed")

	return input, nil
}

func Activity2(ctx context.Context, input interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity2 executed")

	return input, nil
}
