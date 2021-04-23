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

	client "github.com/ContextLogic/cadence/pkg"
	"github.com/ContextLogic/cadence/pkg/fsm/workflow"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/activity"
	temporal "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var logger = logrus.New()

const (
	hostport  = "temporal-fe-dev.service.consul:7233"
	namespace = "cadence_lib"
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
		w, err := workflow.New(bytes, "example")
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

func RunWorkflows() {
	c, err := client.New(
		temporal.Options{
			HostPort:  hostport,
			Namespace: namespace,
		},
		map[string]worker.Options{
			"example": worker.Options{},
		},
	)
	if err != nil {
		panic(err)
	}

	wfs := LoadWorkflows("./cmd/workflows.json")
	activityMap := map[string]func(context.Context, map[string]interface{}) (interface{}, error){
		"example:activity:Activity1": Activity1,
		"example:activity:Activity2": Activity2,
	}
	for _, wf := range wfs {
		logger.Info(fmt.Sprintf("register workflow: %s", wf.Name))
		err := c.Register(wf, activityMap)
		if err != nil {
			panic(err)
		}

		logger.Info(fmt.Sprintf("run workflow: %s", wf.Name))
		instance, err := c.ExecuteWorkflow(
			context.Background(),
			temporal.StartWorkflowOptions{
				ID:        strings.Join([]string{namespace, strconv.Itoa(int(time.Now().Unix()))}, "_"),
				TaskQueue: "example",
			},
			wf.Name,
			map[string]interface{}{
				"foo": 4,
			},
		)
		if err != nil {
			panic(err)
		}

		var response interface{}
		err = instance.Get(context.Background(), &response)
		if err != nil {
			panic(err)
		}
		logger.WithFields(logrus.Fields{"wf": wf.Name, "result": response}).Info("workflow result")
	}
}

func main() {
	RunWorkflows()
}
