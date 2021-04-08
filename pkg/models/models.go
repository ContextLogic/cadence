package models

import (
	"context"

	"github.com/ContextLogic/cadence/pkg/utils/jsonpath"
	"go.temporal.io/sdk/workflow"
)

const (
	StatesAll                    = "States.ALL"
	StatesTimeout                = "States.Timeout"
	StatesTaskFailed             = "States.TaskFailed"
	StatesPermissions            = "States.Permissions"
	StatesResultPathMatchFailure = "States.ResultPathMatchFailure"
	StatesBranchFailed           = "States.BranchFailed"
	StatesNoChoiceMatched        = "States.NoChoiceMatched"
)

const (
	Task     StateType = "Task"
	Fail     StateType = "Fail"
	Succeed  StateType = "Succeed"
	Choice   StateType = "Choice"
	Wait     StateType = "Wait"
	Pass     StateType = "Pass"
	Parallel StateType = "Parallel"
)

type (
	Execution func(workflow.Context, interface{}) (interface{}, *string, error)

	StateType string

	Catcher struct {
		ErrorEquals []*string      `json:",omitempty"`
		ResultPath  *jsonpath.Path `json:",omitempty"`
		Next        *string        `json:",omitempty"`
	}

	Retrier struct {
		ErrorEquals     []*string `json:",omitempty"`
		IntervalSeconds *int      `json:",omitempty"`
		MaxAttempts     *int      `json:",omitempty"`
		BackoffRate     *float64  `json:",omitempty"`
		Attempts        int
	}

	TaskHandler func(workflow.Context, string, interface{}) (interface{}, error)

	ActivityMap map[string]func(ctx context.Context, input interface{}) (interface{}, error)
)
