package test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	st "github.com/ContextLogic/cadence/pkg/fsm/state"
	wf "github.com/ContextLogic/cadence/pkg/fsm/workflow"
	"github.com/sirupsen/logrus"
)

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

var succeedMachine = []byte(`
{
	"Name": "SuccessFlow",  
	"StartAt": "Example1",
	"States": {
		"Example1": {
			"Type": "Succeed"
		}
	}
}
`)

var failMachine = []byte(`
{
	"Name": "FailFlow",   
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

var choiceMachine = []byte(`
{
	"Type": "Choice",
	"Choices": [
			{
			  "Variable": "$.value",
			  "NumericEquals": 0,
			  "Next": "ValueIsZero"
			},
			{
			  "And": [
				{
				  "Variable": "$.value",
				  "NumericGreaterThanEquals": 20.5
				},
				{
				  "Variable": "$.value",
				  "NumericLessThan": 30
				}
			  ],
			  "Next": "ValueInTwenties"
			}
		],
	"Default": "DefaultState"
}
`)

var passMachine = []byte(`
{
	"Name": "PassFlow",    
	"StartAt": "Example1",
	"States": {
		"Example1": {
			"Type": "Pass",
			"Next": "Example2"
		},
		"Example2": {
			"Type": "Pass",
			"Result": {
				"test": "example_output"
			},
			"End": true
		}
	}
}
`)

var parallelMachine = []byte(`
{
	"Name": "ParallelFlow",    
	"StartAt": "Example1",
	"States": {
		"Example1": {
			"Type": "Parallel",
			"End": true,
			"Branches": [
				{
					"StartAt": "Branch1",
					"States": {
						"Branch1": {
							"Type": "Pass",
							"Result": {
								"branch1": true
							},
							"End": true
						}
					}
				},
				{
					"StartAt": "Branch2",
					"States": {
						"Branch2": {
							"Type": "Pass",
							"Result": {
								"branch2": true
							},
							"End": true
						}
					}
				}
			  ]
		}
	}
}
`)

var waitMachine = []byte(`
{
	"Name": "WaitFlow",    
	"StartAt": "Example1",
	"States": {
		"Example1": {
			"Type": "Wait",
			"Seconds" : 60,
			"Next": "Example2"
		},
		"Example2": {
			"Type": "Wait",
			"SecondsPath" : "$.input_seconds",
			"Next": "Example3"
		},
		"Example3": {
			"Type": "Wait",
			"TimestampPath" : "$.input_timestamp",
			"End": true
		}
	}
}
`)

var mapMachine = []byte(`
{
	"Name": "mapFlow",
	"StartAt": "MapExecution",
	"States": {
		"MapExecution": {
			"Type": "Map",
			"MaxConcurrency": 2,
			"Iterator": {
				"StartAt": "Activity3",
				"States": {
					"Activity3": {
					"Type": "Task",
					"Resource": "example:activity:Activity3",
					"End": true
					}
				}
			},
			"Next": "SuccessState"
		},
		"SuccessState": {
		"Type": "Succeed"
		}
	}
}
`)

var taskMachine = []byte(`
{
	"Name": "TaskFlow",
	"StartAt": "State1",
	"States": {
		"State1": {
			"Type": "Task",
			"Resource": "example:activity:Activity1",
			"Next": "State2"
		},
		"State2": {
			"Type": "Task",
			"Resource": "example:activity:Activity2",
			"End": true
		}
	}
}
`)

func Activity1(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity executed")

	if val, ok := input["first"]; ok {
		input["first"] = val.(string) + activityName
	}
	return input, nil
}

func Activity2(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity executed")
	if val, ok := input["first"]; ok {
		input["first"] = val.(string) + activityName
	}
	return input, nil
}
func Activity3(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	activityInfo := activity.GetInfo(ctx)
	taskToken := string(activityInfo.TaskToken)
	activityName := activityInfo.ActivityType.Name

	logger.WithFields(logrus.Fields{"input": input, "taskToken": taskToken, "activityName": activityName}).Info("activity executed")

	return "Updated", nil
}
func (s *UnitTestSuite) Test_Succeed_State() {
	w := &wf.Workflow{}
	input := map[string]interface{}{"test": "example_input"}
	if err := json.Unmarshal(succeedMachine, &w); err == nil {
		w.TaskStates = append(w.TaskStates, st.TasksFromStates(w.States)...)
		s.env.ExecuteWorkflow(w.Execute, input)
		var response map[string]interface{}
		err = s.env.GetWorkflowResult(&response)
		s.Equal(true, s.env.IsWorkflowCompleted())
		s.NoError(err)
	} else {
		s.Assert().Error(err)
	}
}

func (s *UnitTestSuite) Test_Fail_State() {
	w := &wf.Workflow{}
	input := map[string]interface{}{"test": "example_input"}
	if err := json.Unmarshal(failMachine, &w); err == nil {
		w.TaskStates = append(w.TaskStates, st.TasksFromStates(w.States)...)
		s.env.ExecuteWorkflow(w.Execute, input)
		var response []map[string]bool
		err = s.env.GetWorkflowResult(&response)
		s.Equal(true, s.env.IsWorkflowCompleted())
		s.True(strings.Contains(s.env.GetWorkflowError().Error(), "FailState"))
	} else {
		s.Assert().Error(err)
	}
}

func (s *UnitTestSuite) Test_Choice_State() {
	state, _ := st.NewChoiceState("test", choiceMachine)
	input1 := map[string]interface{}{"value": 0}
	_, next, _ := state.Execute(nil, input1)
	s.Equal(*next, "ValueIsZero")

	input2 := map[string]interface{}{"value": 25}
	_, next, _ = state.Execute(nil, input2)
	s.Equal(*next, "ValueInTwenties")

	input3 := map[string]interface{}{"value": 40}
	_, next, _ = state.Execute(nil, input3)
	s.Equal(*next, "DefaultState")
}

func (s *UnitTestSuite) Test_Pass_Workflow() {
	w := &wf.Workflow{}
	input := map[string]interface{}{"test": "example_input"}
	if err := json.Unmarshal(passMachine, &w); err == nil {
		w.TaskStates = append(w.TaskStates, st.TasksFromStates(w.States)...)
		output, err := w.Execute(nil,
			input,
		)
		converted := output.(map[string]interface{})
		s.Equal("example_output", converted["test"])
		s.True(err == nil)
	} else {
		s.Assert().Error(err)
	}
}

func (s *UnitTestSuite) Test_Wait_Workflow() {
	w := &wf.Workflow{}
	if err := json.Unmarshal(waitMachine, &w); err == nil {
		s.env.SetStartTime(time.Now().UTC().Truncate(time.Minute))
		startTime := s.env.Now().UTC()
		inputTime := startTime.Add(4 * time.Minute).Format(time.RFC3339)
		input := map[string]interface{}{"input_seconds": 60, "input_timestamp": inputTime}
		s.env.ExecuteWorkflow(w.Execute, input)

		s.True(s.env.IsWorkflowCompleted())
		s.NoError(s.env.GetWorkflowError())

		endTime := s.env.Now().UTC()

		// Check that the workflow took 4 minutes
		s.Equal(4*time.Minute, endTime.Sub(startTime))

		var result map[string]interface{}
		err = s.env.GetWorkflowResult(&result)
		s.NoError(err)

		s.Equal(float64(60), result["input_seconds"])
		s.Equal(inputTime, result["input_timestamp"])

	} else {
		s.Assert().Error(err)
	}
}

func (s *UnitTestSuite) Test_Parallel_Workflow() {
	w := &wf.Workflow{}
	input := map[string]interface{}{"test": "example_input"}
	if err := json.Unmarshal(parallelMachine, &w); err == nil {
		w.TaskStates = append(w.TaskStates, st.TasksFromStates(w.States)...)
		s.env.ExecuteWorkflow(w.Execute, input)
		var response []map[string]bool
		err = s.env.GetWorkflowResult(&response)
		s.NoError(err)
		s.True(response[0]["branch1"])
		s.True(response[1]["branch2"])
	} else {
		s.Assert().Error(err)
	}
}

func (s *UnitTestSuite) Test_Task_Workflow() {
	w := &wf.Workflow{}
	activityMap := map[string]func(context.Context, map[string]interface{}) (interface{}, error){
		"example:activity:Activity1": Activity1,
		"example:activity:Activity2": Activity2,
	}

	input := map[string]interface{}{"first": "Hello", "second": "Hello again"}
	if err := json.Unmarshal(taskMachine, &w); err == nil {
		w.TaskStates = append(w.TaskStates, st.TasksFromStates(w.States)...)
		w.RegisterTaskHandlers(activityMap)
		f := func(ctx workflow.Context, input interface{}) (interface{}, error) {
			return w.Execute(
				workflow.WithActivityOptions(
					ctx,
					workflow.ActivityOptions{
						ScheduleToStartTimeout: time.Minute,
						StartToCloseTimeout:    time.Minute,
						HeartbeatTimeout:       time.Minute,
					},
				),
				input,
			)
		}
		s.env.RegisterActivity(Activity1)
		s.env.RegisterActivity(Activity2)
		s.env.ExecuteWorkflow(f, input)

		response := map[string]interface{}{}
		err = s.env.GetWorkflowResult(&response)
		s.NoError(err)
		s.Equal(2, len(response))
		s.Equal("HelloActivity1Activity2", response["first"])
		s.Equal("Hello again", response["second"])

	} else {
		s.Assert().Error(err)
	}
}

func (s *UnitTestSuite) Test_Map_Workflow() {
	w := &wf.Workflow{}
	activityMap := map[string]func(context.Context, map[string]interface{}) (interface{}, error){
		"example:activity:Activity3": Activity3,
	}

	input := map[string]interface{}{"first": map[string]interface{}{"1": "hello1", "2": "hello2"}, "second": map[string]interface{}{"3": "hello3", "4": "hello4"}}
	if err := json.Unmarshal(mapMachine, &w); err == nil {
		w.TaskStates = append(w.TaskStates, st.TasksFromStates(w.States)...)
		w.RegisterTaskHandlers(activityMap)
		f := func(ctx workflow.Context, input interface{}) (interface{}, error) {
			return w.Execute(
				workflow.WithActivityOptions(
					ctx,
					workflow.ActivityOptions{
						ScheduleToStartTimeout: time.Minute,
						StartToCloseTimeout:    time.Minute,
						HeartbeatTimeout:       time.Minute,
					},
				),
				input,
			)
		}
		s.env.RegisterActivity(Activity3)
		s.env.ExecuteWorkflow(f, input)
		var response []string
		err = s.env.GetWorkflowResult(&response)
		s.NoError(err)
		s.Equal(2, len(response))
		s.Equal("Updated", response[0])
		s.Equal("Updated", response[1])

	} else {
		s.Assert().Error(err)
	}
}
