package state

import (
	"encoding/json"
	"fmt"

	"github.com/ContextLogic/cadence/pkg/models"
	"github.com/coinbase/step/utils/is"
	"go.temporal.io/sdk/workflow"
)

type FailState struct {
	StateImpl

	Error *string `json:",omitempty"`
	Cause *string `json:",omitempty"`
}

func NewFailState(name string, data []byte) (*FailState, error) {
	t := &FailState{}
	err := json.Unmarshal(data, t)
	if err != nil {
		return nil, err
	}
	t.SetName(&name)
	return t, nil
}

func (s *FailState) Execute(_ workflow.Context, input interface{}) (output interface{}, next *string, err error) {
	return nil, nil, fmt.Errorf("FailState Error:%v, Cause:%v", s.Error, s.Cause)
}

func (s *FailState) Validate() error {
	t := models.Fail
	s.SetType(&t)

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", ErrorPrefix(s), err)
	}

	if is.EmptyStr(s.Error) {
		return fmt.Errorf("%v %v", ErrorPrefix(s), "must contain Error")
	}

	return nil
}

func (s *FailState) SetType(t *models.StateType) {
	s.Type = t
}

func (s *FailState) GetType() *models.StateType {
	return s.Type
}
