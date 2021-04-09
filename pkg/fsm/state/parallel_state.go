package state

import (
	"encoding/json"
	"fmt"

	"github.com/ContextLogic/cadence/pkg/models"
	"github.com/ContextLogic/cadence/pkg/utils/jsonpath"
	"go.temporal.io/sdk/workflow"
)

type ParallelState struct {
	StateImpl

	Branches []Branch

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`
	ResultPath *jsonpath.Path `json:",omitempty"`
	Parameters interface{}    `json:",omitempty"`

	Catch []*models.Catcher `json:",omitempty"`
	Retry []*models.Retrier `json:",omitempty"`

	Next *string `json:",omitempty"`
	End  *bool   `json:",omitempty"`
}

type Branch struct {
	States  States
	StartAt string
}

func NewParallelState(name string, data []byte) (*ParallelState, error) {
	t := &ParallelState{}
	err := json.Unmarshal(data, t)
	if err != nil {
		return nil, err
	}
	t.SetName(&name)
	return t, nil
}

func (s *ParallelState) process(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
	// You can use the context passed in to activity as a way to cancel the activity like standard GO way.
	// Cancelling a parent context will cancel all the derived contexts as well.
	// In the parallel block, we want to execute all of them in parallel and wait for all of them.
	// if one activity fails then we want to cancel all the rest of them as well.
	childCtx, cancelHandler := workflow.WithCancel(ctx)
	selector := workflow.NewSelector(ctx)
	var activityErr error

	var resp []interface{}

	for _, branch := range s.Branches {
		f := executeAsync(s, branch, childCtx, input)
		selector.AddFuture(f, func(f workflow.Future) {
			var r interface{}
			err := f.Get(ctx, &r)
			if err != nil {
				// cancel all pending activities
				cancelHandler()
				activityErr = err
			}
			resp = append(resp, r)
		})
	}

	for i := 0; i < len(s.Branches); i++ {
		selector.Select(ctx) // this will wait for one branch
		if activityErr != nil {
			return nil, nil, activityErr
		}
	}

	return interface{}(resp), s.NextState(s.Next, s.End), nil
}

func (s *ParallelState) Execute(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
	f := ProcessResult(s.ResultPath, s.process)
	f = ProcessParams(s.Parameters, f)
	f = ProcessInputOutput(s.InputPath, s.OutputPath, f)
	f = ProcessRetrier(s.GetName(), s.Retry, f)
	f = ProcessCatcher(s.Catch, f)
	return ProcessError(s, f)(ctx, input)
}

func executeAsync(p *ParallelState, b Branch, ctx workflow.Context, input interface{}) workflow.Future {
	future, settable := workflow.NewFuture(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		output, _, err := b.Execute(ctx, *p, input)
		settable.Set(output, err)
	})
	return future
}

func (m *Branch) Execute(ctx workflow.Context, s ParallelState, input interface{}) (interface{}, *string, error) {
	nextState := &m.StartAt

	for {
		s := m.States[*nextState]
		if s == nil {
			return nil, nil, fmt.Errorf("next state is invalid (%v)", *nextState)
		}

		output, next, err := s.Execute(ctx, input)

		if err != nil {
			return nil, nil, err
		}

		if next == nil || *next == "" {
			return output, nil, nil
		}

		nextState = next
		input = output
	}
}

func (m *Branch) Tasks() []*TaskState {
	var tasks []*TaskState
	tasks = append(tasks, TasksFromStates(m.States)...)
	return tasks
}

func (s *ParallelState) Validate() error {
	t := models.Parallel
	s.SetType(&t)

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", ErrorPrefix(s), err)
	}

	return nil
}

func (s *ParallelState) SetType(t *models.StateType) {
	s.Type = t
}

func (s *ParallelState) GetType() *models.StateType {
	return s.Type
}
