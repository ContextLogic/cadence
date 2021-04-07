package test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	wf "github.com/ContextLogic/cadence/pkg/models/workflow"
	"github.com/ContextLogic/cadence/pkg/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
)

var asl_successState = []byte(`
{
	"StartAt": "Example1",
	"States": {
		"Example1": {
			"Type": "Succeed"
		}
	}
}
`)

var asl_failState = []byte(`
{
	"StartAt": "Example1",
	"States": {
		"Example1": {
			"Type": "Fail",
			"Error": "ExampleError",
			"Cause": "This is an example error",
			"End": true
		}
	}
}
`)

var logger = logrus.New()

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) TestWorkflowSucceedState() {
	w, err := wf.New(asl_successState)
	if err != nil {
		panic(fmt.Errorf("error loading workflow %w", err))
	}
	logger.WithFields(logrus.Fields{"workflow": w}).Info("workflows")

	name := "TestWorkflowSucceedState"
	w.RegisterWorkflow(name)
	w.RegisterWorker()

	instance, err := temporal.BaseClient.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			ID:        strings.Join([]string{"cadence_lib", strconv.Itoa(int(time.Now().Unix()))}, "_"),
			TaskQueue: "TASK_QUEUE_cadence_lib",
		},
		name,
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

func (s *UnitTestSuite) TestWorkflowFailState() {
	w, err := wf.New(asl_failState)
	if err != nil {
		panic(fmt.Errorf("error loading workflow %w", err))
	}
	logger.WithFields(logrus.Fields{"workflow": w}).Info("workflows")

	name := "TestWorkflowFailState"
	w.RegisterWorkflow(name)
	w.RegisterWorker()

	instance, err := temporal.BaseClient.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			ID:        strings.Join([]string{"cadence_lib", strconv.Itoa(int(time.Now().Unix()))}, "_"),
			TaskQueue: "TASK_QUEUE_cadence_lib",
		},
		name,
		map[string]interface{}{
			"hello": "world",
		},
	)
	if err != nil {
		panic(err)
	}
	response := map[string]interface{}{}
	err = instance.Get(context.Background(), &response)
	if err == nil {
		panic(err)
	}
}
