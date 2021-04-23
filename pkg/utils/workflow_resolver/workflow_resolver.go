package workflow_resolver

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/ContextLogic/cadence/pkg/fsm/workflow"
)

func ResolveWorkflow(i interface{}) ([]*workflow.Workflow, error) {
	workflows := []*workflow.Workflow{}
	switch i.(type) {
	case string:
		return resolveWorkflowFromFile(i.(string))
	case *workflow.Workflow:
		workflows = append(workflows, i.(*workflow.Workflow))
	case []interface{}:
		for _, w := range i.([]interface{}) {
			if _, ok := w.(*workflow.Workflow); !ok {
				return nil, errors.New("invalid workflow type")
			}
			workflows = append(workflows, w.(*workflow.Workflow))
		}
	default:
		return nil, errors.New("invalide workflow")
	}
	return workflows, nil
}

func resolveWorkflowFromFile(path string) ([]*workflow.Workflow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	value, _ := ioutil.ReadAll(f)
	m := []map[string]interface{}{}
	if err := json.Unmarshal(value, &m); err != nil {
		return nil, err
	}

	workflows := []*workflow.Workflow{}
	for _, v := range m {
		bytes, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		w, err := workflow.New(bytes, "example")
		if err != nil {
			return nil, err
		}
		workflows = append(workflows, w)
	}

	return workflows, nil
}
