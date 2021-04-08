package state

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ContextLogic/cadence/pkg/models"
	"github.com/ContextLogic/cadence/pkg/utils/jsonpath"
	"go.temporal.io/sdk/workflow"
)

type WaitState struct {
	StateImpl

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`

	Seconds     *float64       `json:",omitempty"`
	SecondsPath *jsonpath.Path `json:",omitempty"`

	Timestamp     *time.Time     `json:",omitempty"`
	TimestampPath *jsonpath.Path `json:",omitempty"`

	Next *string `json:",omitempty"`
	End  *bool   `json:",omitempty"`
}

func NewWaitState(name string, data []byte) (*WaitState, error) {
	t := &WaitState{}
	err := json.Unmarshal(data, t)
	if err != nil {
		return nil, err
	}
	t.SetName(&name)
	return t, nil
}

func (s *WaitState) process(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
	var duration time.Duration

	if s.SecondsPath != nil {
		// Validate the path exists
		secs, err := s.SecondsPath.GetNumber(input)
		if err != nil {
			return nil, nil, err
		}
		duration = time.Duration(*secs) * time.Second

	} else if s.Seconds != nil {
		duration = time.Duration(*s.Seconds) * time.Second

	} else if s.TimestampPath != nil {
		// Validate the path exists
		ts, err := s.TimestampPath.GetTime(input)
		if err != nil {
			return nil, nil, err
		}
		now := workflow.Now(ctx).UTC()
		duration = ts.Sub(now)

	} else if s.Timestamp != nil {
		now := workflow.Now(ctx).UTC()
		duration = s.Timestamp.Sub(now)
	}

	err := workflow.Sleep(ctx, duration)
	if err != nil {
		return nil, nil, err
	}

	return input, s.NextState(s.Next, s.End), nil
}

func (s *WaitState) Execute(ctx workflow.Context, input interface{}) (output interface{}, next *string, err error) {
	f := ProcessInputOutput(s.InputPath, s.OutputPath, s.process)
	return ProcessError(s, f)(ctx, input)
}

func (s *WaitState) Validate() error {
	t := models.Wait
	s.SetType(&t)

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", ErrorPrefix(s), err)
	}

	// Next xor End
	if err := IsEndValid(s.Next, s.End); err != nil {
		return fmt.Errorf("%v %v", ErrorPrefix(s), err)
	}

	exactlyOne := []bool{
		s.Seconds != nil,
		s.SecondsPath != nil,
		s.Timestamp != nil,
		s.TimestampPath != nil,
	}

	count := 0
	for _, c := range exactlyOne {
		if c {
			count++
		}
	}

	if count != 1 {
		return fmt.Errorf("%v Exactly One (Seconds,SecondsPath,TimeStamp,TimeStampPath)", ErrorPrefix(s))
	}

	return nil
}

func (s *WaitState) SetType(t *models.StateType) {
	s.Type = t
}

func (s *WaitState) GetType() *models.StateType {
	return s.Type
}
