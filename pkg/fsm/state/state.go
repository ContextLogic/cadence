package state

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ContextLogic/cadence/pkg/models"
	"github.com/ContextLogic/cadence/pkg/utils/jsonpath"
	"github.com/coinbase/step/utils/is"
	"github.com/coinbase/step/utils/to"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
)

type (
	States map[string]State

	State interface {
		Execute(workflow.Context, interface{}) (interface{}, *string, error)
		Validate() error

		SetName(*string)
		SetType(*models.StateType)

		GetName() *string
		GetType() *models.StateType
		NextState(*string, *bool) *string
	}

	StateImpl struct {
		Name    *string
		Type    *models.StateType
		Comment *string `json:",omitempty"`
	}
)

func (s *States) UnmarshalJSON(b []byte) error {
	var raw map[string]*json.RawMessage
	err := json.Unmarshal(b, &raw)

	if err != nil {
		return err
	}

	states := &States{}
	for name, r := range raw {
		si := &StateImpl{}
		if err = json.Unmarshal(*r, si); err != nil {
			return err
		}

		switch *si.Type {
		case "Task":
			state, err := NewTaskState(name, *r)
			if err != nil {
				return err
			}
			(*states)[*state.GetName()] = state
		case "Succeed":
			state, err := NewSucceedState(name, *r)
			if err != nil {
				return err
			}
			(*states)[*state.GetName()] = state
		case "Fail":
			state, err := NewFailState(name, *r)
			if err != nil {
				return err
			}
			(*states)[*state.GetName()] = state
		default:
			return fmt.Errorf("unknown state %q", *si.Type)
		}
	}

	*s = *states
	return nil
}

func (s *StateImpl) GetName() *string {
	return s.Name
}

func (s *StateImpl) SetName(name *string) {
	s.Name = name
}

func (s *StateImpl) NextState(next *string, end *bool) *string {
	if next != nil {
		return next
	}
	return nil
}

func errorOutputFromError(err error) map[string]interface{} {
	return errorOutput(to.Strp(to.ErrorType(err)), to.Strp(err.Error()))
}

func errorOutput(err *string, cause *string) map[string]interface{} {
	errstr := ""
	causestr := ""
	if err != nil {
		errstr = *err
	}
	if cause != nil {
		causestr = *cause
	}
	return map[string]interface{}{
		"Error": errstr,
		"Cause": causestr,
	}
}

func errorIncluded(errorEquals []*string, err error) bool {
	errorType := to.ErrorType(err)

	for _, et := range errorEquals {
		if *et == models.StatesAll || *et == errorType {
			return true
		}
	}

	return false
}

func ProcessRetrier(retryName *string, retriers []*models.Retrier, execution models.Execution) models.Execution {
	return func(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
		// Simulate Retry once, not actually waiting
		output, next, err := execution(ctx, input)
		if len(retriers) == 0 || err == nil {
			return output, next, err
		}

		// Is Error in a Retrier
		for _, retrier := range retriers {
			// If the error type is defined in the retrier AND we have not attempted the retry yet
			if retrier.MaxAttempts == nil {
				// Default retries is 3
				retrier.MaxAttempts = to.Intp(3)
			}

			// Match on first retrier
			if errorIncluded(retrier.ErrorEquals, err) {
				if retrier.Attempts < *retrier.MaxAttempts {
					retrier.Attempts++
					// Returns the name of the state to the state-machine to re-execute
					return input, retryName, nil
				}
				// Finished retrying so continue
				return output, next, err
			}
		}

		// Otherwise, just return
		return output, next, err
	}
}

func ProcessCatcher(catchers []*models.Catcher, execution models.Execution) models.Execution {
	return func(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
		output, next, err := execution(ctx, input)

		if len(catchers) == 0 || err == nil {
			return output, next, err
		}

		for _, catcher := range catchers {
			if errorIncluded(catcher.ErrorEquals, err) {

				eo := errorOutputFromError(err)
				output, err := catcher.ResultPath.Set(input, eo)

				return output, catcher.Next, err
			}
		}

		// Otherwise continue
		return output, next, err
	}
}

func ProcessError(s State, execution models.Execution) models.Execution {
	return func(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
		output, next, err := execution(ctx, input)

		if err != nil {
			return nil, nil, fmt.Errorf("%v %w", ErrorPrefix(s), err)
		}

		return output, next, nil
	}
}

func ProcessInputOutput(inputPath *jsonpath.Path, outputPath *jsonpath.Path, execution models.Execution) models.Execution {
	return func(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
		input, err := inputPath.Get(input)

		if err != nil {
			return nil, nil, fmt.Errorf("Input Error: %v", err)
		}

		output, next, err := execution(ctx, input)

		if err != nil {
			return nil, nil, err
		}

		output, err = outputPath.Get(output)

		if err != nil {
			return nil, nil, fmt.Errorf("Output Error: %v", err)
		}

		return output, next, nil
	}
}

func ProcessParams(params interface{}, execution models.Execution) models.Execution {
	return func(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
		if params == nil {
			return execution(ctx, input)
		}
		// Loop through the input replace values with JSON paths

		input, err := replaceParamsJSONPath(params, input)
		if err != nil {
			return nil, nil, err
		}

		return execution(ctx, input)
	}
}

func replaceParamsJSONPath(params interface{}, input interface{}) (interface{}, error) {
	switch params.(type) {
	case map[string]interface{}:
		newParams := map[string]interface{}{}
		// Recurse over params find keys to replace
		for key, value := range params.(map[string]interface{}) {
			if strings.HasSuffix(key, ".$") {
				key = key[:len(key)-len(".$")]
				// value must be a JSON path string!
				switch value.(type) {
				case string:
				default:
					return nil, fmt.Errorf("value to key %q is not string", key)
				}
				valueStr := value.(string)
				path, err := jsonpath.NewPath(valueStr)
				if err != nil {
					return nil, errors.Wrap(err, "failed parsing path")
				}
				newValue, err := path.Get(input)
				if err != nil {
					return nil, errors.Wrap(err, "failed getting path")
				}
				newParams[key] = newValue
			} else {
				newValue, err := replaceParamsJSONPath(value, input)
				if err != nil {
					return nil, errors.Wrap(err, "failed replacing path")
				}
				newParams[key] = newValue
			}
		}
		return newParams, nil
	}
	return params, nil
}

func ProcessResult(resultPath *jsonpath.Path, execution models.Execution) models.Execution {
	return func(ctx workflow.Context, input interface{}) (interface{}, *string, error) {
		result, next, err := execution(ctx, input)

		if err != nil {
			return nil, nil, err
		}

		if result != nil {
			input, err := resultPath.Set(input, result)

			if err != nil {
				return nil, nil, err
			}

			return input, next, nil
		}
		return input, next, nil
	}
}

//////
// Shared Validity Methods
//////

func IsEndValid(next *string, end *bool) error {
	if end == nil && next == nil {
		return fmt.Errorf("End and Next both undefined")
	}

	if end != nil && next != nil {
		return fmt.Errorf("End and Next both defined")
	}

	if end != nil && !*end {
		return fmt.Errorf("End can only be true or nil")
	}

	return nil
}

func ErrorPrefix(s State) string {
	if !is.EmptyStr(s.GetName()) {
		return fmt.Sprintf("%vState(%v) Error:", *s.GetType(), *s.GetName())
	}

	return fmt.Sprintf("%vState Error:", *s.GetType())
}

func ValidateNameAndType(s State) error {
	if is.EmptyStr(s.GetName()) {
		return fmt.Errorf("Must have Name")
	}

	t := string(*s.GetType())
	if is.EmptyStr(&t) {
		return fmt.Errorf("Must have Type")
	}

	return nil
}

func IsRetryValid(retry []*models.Retrier) error {
	if retry == nil {
		return nil
	}

	for i, r := range retry {
		if err := isErrorEqualsValid(r.ErrorEquals, len(retry)-1 == i); err != nil {
			return err
		}
	}

	return nil
}

func IsCatchValid(catch []*models.Catcher) error {
	if catch == nil {
		return nil
	}

	for i, c := range catch {
		if err := isErrorEqualsValid(c.ErrorEquals, len(catch)-1 == i); err != nil {
			return err
		}

		if is.EmptyStr(c.Next) {
			return fmt.Errorf("Catcher requires Next")
		}
	}
	return nil
}

func isErrorEqualsValid(errorEquals []*string, last bool) error {
	if len(errorEquals) == 0 {
		return fmt.Errorf("Retrier requires ErrorEquals")
	}

	for _, e := range errorEquals {
		// If it is a States. Error, then must match one of the defined values
		if strings.HasPrefix(*e, "States.") {
			switch *e {
			case
				models.StatesAll,
				models.StatesTimeout,
				models.StatesTaskFailed,
				models.StatesPermissions,
				models.StatesResultPathMatchFailure,
				models.StatesBranchFailed,
				models.StatesNoChoiceMatched:
			default:
				return fmt.Errorf("Unknown States.* error found %q", *e)
			}
		}

		if *e == models.StatesAll {
			if len(errorEquals) != 1 {
				return fmt.Errorf(`"States.ALL" ErrorEquals must be only element in list`)
			}

			if !last {
				return fmt.Errorf(`"States.ALL" must be last Catcher/Retrier`)
			}
		}
	}

	return nil
}
