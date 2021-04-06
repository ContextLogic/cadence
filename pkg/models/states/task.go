package states

import (
	"encoding/json"
	"fmt"

	errs "github.com/ContextLogic/cadence/pkg/errors"
	"github.com/ContextLogic/cadence/pkg/utils/jsonpath"
	"go.temporal.io/sdk/workflow"
)

type (
	TaskHandler func(workflow.Context, string, interface{}) (interface{}, error)

	TaskState struct {
		StateImpl

		InputPath  *jsonpath.Path `json:",omitempty"`
		OutputPath *jsonpath.Path `json:",omitempty"`
		ResultPath *jsonpath.Path `json:",omitempty"`
		Parameters interface{}    `json:",omitempty"`

		Resource *string `json:",omitempty"`

		Catch []*Catcher `json:",omitempty"`
		Retry []*Retrier `json:",omitempty"`

		Next *string `json:",omitempty"`
		End  *bool   `json:",omitempty"`

		TimeoutSeconds   int `json:",omitempty"`
		HeartbeatSeconds int `json:",omitempty"`

		Handler TaskHandler `json:"-"`
	}
)

func NewTaskState(name string, data []byte) (*TaskState, error) {
	t := &TaskState{}
	err := json.Unmarshal(data, t)
	if err != nil {
		return nil, err
	}
	t.SetName(&name)
	return t, nil
}

func (s *TaskState) process(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
	if s.Handler != nil {
		result, err := s.Handler(ctx, *s.Resource, input)
		if err != nil {
			return nil, nil, err
		}
		return result, NextState(s.Next, s.End), nil
	}

	return nil, nil, errs.ErrTaskHandlerNotRegistered
}

// Input must include the Task name in $.Task
func (s *TaskState) Execute(ctx workflow.Context, input interface{}) (output interface{}, next *string, err error) {
	f := ProcessResult(s.ResultPath, s.process)
	f = ProcessParams(s.Parameters, f)
	f = ProcessInputOutput(s.InputPath, s.OutputPath, f)
	f = ProcessRetrier(s.GetName(), s.Retry, f)
	f = ProcessCatcher(s.Catch, f)
	return ProcessError(s, f)(ctx, input)
}

func (s *TaskState) Validate() error {
	t := Task
	s.SetType(&t)

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %w", ErrorPrefix(s), err)
	}

	if err := IsEndValid(s.Next, s.End); err != nil {
		return fmt.Errorf("%v %w", ErrorPrefix(s), err)
	}

	if s.Resource == nil {
		return fmt.Errorf("%v Requires Resource", ErrorPrefix(s))
	}

	// TODO: implement custom handlers
	//if s.taskHandler != nil {
	//}

	if err := IsCatchValid(s.Catch); err != nil {
		return err
	}

	if err := IsRetryValid(s.Retry); err != nil {
		return err
	}

	return nil
}

func (s *TaskState) SetType(t *StateType) {
	s.Type = t
}

func (s *TaskState) GetType() *StateType {
	return s.Type
}

func (s *TaskState) RegisterHandler(f TaskHandler) {
	s.Handler = f
}

func (s *TaskState) DeregisterHandler() {
	s.Handler = nil
}
