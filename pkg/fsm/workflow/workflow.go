package workflow

import (
	"encoding/json"
	"fmt"
	"time"

	s "github.com/ContextLogic/cadence/pkg/fsm/state"
	"github.com/ContextLogic/cadence/pkg/models"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

var RegisteredActivities = make(map[string]*struct{})

type Workflow struct {
	Name           string   `json:"Name"`
	Queue          string   `json:"Queue"`
	States         s.States `json:"States"`
	StartAt        string   `json:"StartAt"`
	Comment        string   `json:"Comment"`
	Version        string   `json:"Version"`
	TimeoutSeconds int32    `json:"TimeoutSeconds"`

	TaskStates []*s.TaskState `json:"-"`
}

func New(raw []byte, queue string) (*Workflow, error) {
	w := &Workflow{Queue: queue}
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, err
	}

	var tasks = s.TasksFromStates(w.States)
	w.TaskStates = append(w.TaskStates, tasks...)
	return w, nil
}

func (wf *Workflow) Execute(ctx workflow.Context, input interface{}) (interface{}, error) {
	n := &wf.StartAt
	for {
		s, ok := wf.States[*n]
		if !ok {
			return nil, fmt.Errorf("next state invalid in workflow (%v)", *n)
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

func (wf *Workflow) RegisterWorkflow(wk worker.Worker) {
	f := func(ctx workflow.Context, input interface{}) (interface{}, error) {
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
	wk.RegisterWorkflowWithOptions(f, workflow.RegisterOptions{Name: wf.Name})
}

func (wf *Workflow) RegisterActivities(activities models.ActivityMap, wk worker.Worker) {
	for _, task := range wf.TaskStates {
		a, ok := activities[*task.Resource]
		if !ok {
			continue
		}
		if _, ok := RegisteredActivities[*task.Resource]; ok {
			continue
		}
		RegisteredActivities[*task.Resource] = nil
		wk.RegisterActivityWithOptions(a, activity.RegisterOptions{Name: *task.Resource})
	}
}

func (wf *Workflow) RegisterWorker(wc worker.Worker) {
	go wc.Run(worker.InterruptCh())
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
