package main

import (
	"context"
	"fmt"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	client "github.com/ContextLogic/cadence/pkg"
	"github.com/ContextLogic/cadence/pkg/models"
	"github.com/sirupsen/logrus"
	temporal "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const (
	hostport  = "temporal-fe-dev.service.consul:7233"
	namespace = "cadence_lib"
)

var (
	wp         string
	c          client.Client
	logger     = logrus.New()
	activities = models.ActivityMap{
		"example:activity:Activity1": Activity1,
		"example:activity:Activity2": Activity2,
	}
	wfnames = []string{
		"example:workflow:ExampleWorkflow1",
		"example:workflow:ExampleWorkflow2",
		"example:workflow:ExampleWorkflow3",
	}
)

func init() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	_, filename, _, _ := runtime.Caller(0)
	wp = path.Join(path.Dir(filename), "workflows.json")
	var err error
	c, err = client.New(temporal.Options{HostPort: hostport, Namespace: namespace})
	if err != nil {
		panic(err)
	}
}

func main() {
	workerOptions := map[string]worker.Options{
		// queue: worker_options
		"example": worker.Options{},
	}
	logger.Info("register workflow")
	err := c.Register(wp, workerOptions, activities)
	if err != nil {
		panic(err)
	}

	for _, wfname := range wfnames {
		logger.Info(fmt.Sprintf("run workflow: %s", wfname))
		instance, err := c.ExecuteWorkflow(
			context.Background(),
			temporal.StartWorkflowOptions{
				ID:        strings.Join([]string{namespace, strconv.Itoa(int(time.Now().Unix()))}, "_"),
				TaskQueue: "example",
			},
			wfname,
			map[string]interface{}{"foo": 4},
		)
		if err != nil {
			panic(err)
		}

		var response interface{}
		err = instance.Get(context.Background(), &response)
		if err != nil {
			panic(err)
		}
		logger.WithFields(logrus.Fields{
			"wf":     wfname,
			"result": response,
		}).Info("workflow result")
	}
}
