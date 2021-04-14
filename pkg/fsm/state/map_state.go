package state

import (
	"encoding/json"
	"fmt"

	"github.com/ContextLogic/cadence/pkg/models"
	"github.com/ContextLogic/cadence/pkg/utils/jsonpath"
	"go.temporal.io/sdk/workflow"
)

type MapState struct {
	StateImpl

	Iterator Iterator

	InputPath  *jsonpath.Path `json:",omitempty"`
	OutputPath *jsonpath.Path `json:",omitempty"`
	ResultPath *jsonpath.Path `json:",omitempty"`
	Parameters interface{}    `json:",omitempty"`

	Catch          []*models.Catcher `json:",omitempty"`
	Retry          []*models.Retrier `json:",omitempty"`
	MaxConcurrency *int              `json:",omitempty"`
	Next           *string           `json:",omitempty"`
	End            *bool             `json:",omitempty"`
}

type Iterator struct {
	States  States
	StartAt string
}

func NewMapState(name string, data []byte) (*MapState, error) {
	t := &MapState{}
	err := json.Unmarshal(data, t)
	if err != nil {
		return nil, err
	}
	t.SetName(&name)
	return t, nil
}

func (s *MapState) process(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
	var resp []interface{}
	switch items := input.(type) {
	case map[string]interface{}:
		inChan := workflow.NewBufferedChannel(ctx, len(items))
		outChan := workflow.NewBufferedChannel(ctx, len(items))
		for _, item := range items {
			inChan.Send(ctx, item)
		}
		inChan.Close()
		for x := 0; x < *s.MaxConcurrency; x++ {
			workflow.Go(ctx, func(ctx workflow.Context) {
				var inVal interface{}
				for inChan.Receive(ctx, &inVal) {
					if inVal == nil {
						outChan.Send(ctx, fmt.Errorf("Input item can not be nil"))
					} else {
						output, _, err := s.Iterator.Execute(ctx, *s, inVal)
						if err != nil {
							outChan.Send(ctx, err)
						} else {
							outChan.Send(ctx, output)
						}
					}
				}

			})
		}

		var respVal interface{}
		for i := 0; i < len(items); i++ {
			outChan.Receive(ctx, &respVal)
			resp = append(resp, respVal)
		}
		outChan.Close()
	}

	return interface{}(resp), s.NextState(s.Next, s.End), nil
}

func executeIteratorAsync(p *MapState, b Iterator, ctx workflow.Context, input interface{}) workflow.Future {
	future, settable := workflow.NewFuture(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		output, _, err := b.Execute(ctx, *p, input)
		settable.Set(output, err)
	})
	return future
}

func (s *MapState) Execute(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
	f := ProcessResult(s.ResultPath, s.process)
	f = ProcessParams(s.Parameters, f)
	f = ProcessInputOutput(s.InputPath, s.OutputPath, f)
	f = ProcessRetrier(s.GetName(), s.Retry, f)
	f = ProcessCatcher(s.Catch, f)
	return ProcessError(s, f)(ctx, input)
}

func (m *Iterator) Execute(ctx workflow.Context, s MapState, input interface{}) (interface{}, *string, error) {
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

func (m *Iterator) Tasks() []*TaskState {
	var tasks []*TaskState
	tasks = append(tasks, TasksFromStates(m.States)...)
	return tasks
}

func (s *MapState) Validate() error {
	t := models.Map
	s.SetType(&t)

	if err := ValidateNameAndType(s); err != nil {
		return fmt.Errorf("%v %v", ErrorPrefix(s), err)
	}

	return nil
}

func (s *MapState) SetType(t *models.StateType) {
	s.Type = t
}

func (s *MapState) GetType() *models.StateType {
	return s.Type
}
