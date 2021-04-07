package states

import (
	"encoding/json"
	"fmt"

	"github.com/ContextLogic/cadence/pkg/utils/jsonpath"
	"go.temporal.io/sdk/workflow"
)

type SucceedState struct {
	StateImpl

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`
}

func NewSucceedState(name string, data []byte) (*SucceedState, error) {
	t := &SucceedState{}
	err := json.Unmarshal(data, t)
	if err != nil {
		return nil, err
	}
	t.SetName(&name)
	return t, nil
}

func (s *SucceedState) process(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
	return input, nil, nil
}

func (s *SucceedState) Execute(ctx workflow.Context, input interface{}) (output interface{}, next *string, err error) {
	f := ProcessInputOutput(s.InputPath, s.OutputPath, s.process)
	return ProcessError(s, f)(ctx, input)
}

func (s *SucceedState) Validate() error {
	t := Succeed
	s.SetType(&t)

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", ErrorPrefix(s), err)
	}

	return nil
}

func (s *SucceedState) SetType(t *StateType) {
	s.Type = t
}

func (s *SucceedState) GetType() *StateType {
	return s.Type
}
