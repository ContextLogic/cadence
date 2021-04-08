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

	wf "github.com/ContextLogic/cadence/pkg/fsm/workflow"
	"github.com/ContextLogic/cadence/pkg/temporal"
	"go.temporal.io/sdk/activity"
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

var asl_choiceState = []byte(`
{
    "StartAt": "ChoiceState",
    "States": {
      "ChoiceState": {
        "Type" : "Choice",
        "Choices": [
          {
            "Variable": "$.foo",
            "NumericEquals": 1,
            "Next": "State1"
          },
          {
            "Variable": "$.foo",
            "NumericEquals": 2,
            "Next": "State2"
          }
        ],
        "Default": "DefaultState"
      },
      "State1": {
        "Type": "Task",
        "Resource": "example:activity:Activity1",
        "Next": "SuccessState"
      },
      "State2": {
        "Type": "Task",
        "Resource": "example:activity:Activity2",
        "Next": "SuccessState"
      },
      "FailState": {
        "Type": "Fail",
        "Cause": "No Matches!"
      },
      "SuccessState": {
        "Type": "Succeed"
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

func (s *UnitTestSuite) TestWorkflowChoiceState() {
	w, err := wf.New(asl_choiceState)
	if err != nil {
		panic(fmt.Errorf("error loading workflow %w", err))
	}
	logger.WithFields(logrus.Fields{"workflow": w}).Info("workflows")

	name := "TestWorkflowChoiceState"

	activityMap := map[string]func(context.Context, interface{}) (interface{}, error){
		"example:activity:Activity1": WorkflowActivity1,
		"example:activity:Activity2": WorkflowActivity2,
	}

	w.RegisterWorkflow(name)
	w.RegisterActivities(activityMap)
	w.RegisterWorker()
	w.RegisterTaskHandlers(activityMap)

	instance, err := temporal.BaseClient.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			ID:        strings.Join([]string{"cadence_lib", strconv.Itoa(int(time.Now().Unix()))}, "_"),
			TaskQueue: "TASK_QUEUE_cadence_lib",
		},
		name,
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

func WorkflowActivity1(ctx context.Context, input interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity1 executed")

	return input, nil
}

func WorkflowActivity2(ctx context.Context, input interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity2 executed")

	return input, nil
}
