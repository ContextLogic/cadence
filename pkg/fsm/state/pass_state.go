package state

import (
	"encoding/json"
	"fmt"

	"github.com/ContextLogic/cadence/pkg/models"
	"github.com/ContextLogic/cadence/pkg/utils/jsonpath"
	"go.temporal.io/sdk/workflow"
)

type PassState struct {
	StateImpl

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`
	ResultPath *jsonpath.Path `json:",omitempty"`
	Parameters interface{}    `json:",omitempty"` // TODO: Create a struct for Parameters?

	Result interface{} `json:",omitempty"`

	Next *string `json:",omitempty"`
	End  *bool   `json:",omitempty"`
}

func NewPassState(name string, data []byte) (*PassState, error) {
	t := &PassState{}
	err := json.Unmarshal(data, t)
	if err != nil {
		return nil, err
	}
	t.SetName(&name)
	return t, nil
}

func (s *PassState) Execute(ctx workflow.Context, input interface{}) (output interface{}, next *string, err error) {
	f := ProcessResult(s.ResultPath, s.process)
	f = ProcessParams(s.Parameters, f)
	f = ProcessInputOutput(s.InputPath, s.OutputPath, f)
	return ProcessError(s, f)(ctx, input)
}

func (s *PassState) process(ctx workflow.Context, input interface{}) (output interface{}, next *string, err error) {
	output = input
	if s.Result != nil {
		output = s.Result
	}
	return output, s.NextState(s.Next, s.End), nil
}

func (s *PassState) Validate() error {
	t := models.Pass
	s.SetType(&t)

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", ErrorPrefix(s), err)
	}

	// Next xor End
	if err := IsEndValid(s.Next, s.End); err != nil {
		return fmt.Errorf("%v %v", ErrorPrefix(s), err)
	}

	return nil
}

func (s *PassState) SetType(t *models.StateType) {
	s.Type = t
}

func (s *PassState) GetType() *models.StateType {
	return s.Type
}
