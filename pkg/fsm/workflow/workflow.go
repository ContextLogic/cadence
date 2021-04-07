package workflow

import (
	"encoding/json"
	"fmt"
	"time"

	s "github.com/ContextLogic/cadence/pkg/fsm/state"
	"github.com/ContextLogic/cadence/pkg/models"
	"github.com/ContextLogic/cadence/pkg/temporal"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type Workflow struct {
	States         s.States       `json:"States"`
	TaskStates     []*s.TaskState `json:"-"`
	StartAt        string         `json:"StartAt"`
	Comment        string         `json:"Comment"`
	Version        string         `json:"Version"`
	TimeoutSeconds int32          `json:"TimeoutSeconds"`
}

func New(raw []byte) (*Workflow, error) {
	var w Workflow
	err := json.Unmarshal(raw, &w)
	if err != nil {
		return nil, err
	}

	for _, state := range w.States {
		task, ok := state.(*s.TaskState)
		if !ok {
			continue
		}
		w.TaskStates = append(w.TaskStates, task)
	}

	return &w, nil
}

func (wf *Workflow) Execute(ctx workflow.Context, input interface{}) (interface{}, error) {
	n := &wf.StartAt

	for {
		s, ok := wf.States[*n]
		if !ok {
			return nil, fmt.Errorf("next state invalid (%v)", *n)
		}

		output, next, err := s.Execute(ctx, input)
		if err != nil {
			return nil, err
		}

		if next == nil {
			return output, nil
		}

		n = next
		input = output
	}
}

func (wf *Workflow) RegisterWorkflow(name string) {
	f := func(ctx workflow.Context, input interface{}) (interface{}, error) {
		flow := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
			return *wf
		})

		var w Workflow
		err := flow.Get(&w)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize the workflow: %w", err)
		}

		return wf.Execute(
			workflow.WithActivityOptions(
				ctx,
				workflow.ActivityOptions{
					ScheduleToStartTimeout: time.Minute,
					StartToCloseTimeout:    time.Minute,
					HeartbeatTimeout:       time.Second * 20,
				},
			),
			input,
		)
	}
	temporal.WorkerClient.RegisterWorkflowWithOptions(f, workflow.RegisterOptions{Name: name})
}

func (wf *Workflow) RegisterActivities(activities models.ActivityMap) {
	for _, task := range wf.TaskStates {
		a, ok := activities[*task.Resource]
		if !ok {
			continue
		}
		temporal.WorkerClient.RegisterActivityWithOptions(a, activity.RegisterOptions{Name: *task.Resource})
	}
}

func (wf *Workflow) RegisterWorker() {
	go temporal.WorkerClient.Run(worker.InterruptCh())
}

func (wf *Workflow) RegisterTaskHandlers(activities models.ActivityMap) {
	for _, task := range wf.TaskStates {
		a, ok := activities[*task.Resource]
		if !ok {
			continue
		}
		task.RegisterHandler(
			func(ctx workflow.Context, resource string, input interface{}) (interface{}, error) {
				var result interface{}
				err := workflow.ExecuteActivity(ctx, a, input).Get(ctx, &result)
				if err != nil {
					return nil, err
				}
				return result, nil
			},
		)
	}
}
