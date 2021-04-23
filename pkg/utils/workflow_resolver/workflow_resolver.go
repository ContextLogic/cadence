package workflow_resolver

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/ContextLogic/cadence/pkg/fsm/workflow"
)

func ResolveWorkflow(i interface{}) (map[string]*workflow.Workflow, error) {
	workflows := map[string]*workflow.Workflow{}
	switch i.(type) {
	case string:
		return resolveWorkflowFromFile(i.(string))
	case *workflow.Workflow:
		workflows[i.(*workflow.Workflow).Name] = i.(*workflow.Workflow)
	case []interface{}:
		for _, w := range i.([]interface{}) {
			if _, ok := w.(*workflow.Workflow); !ok {
				return nil, errors.New("invalid workflow type")
			}
			workflows[w.(*workflow.Workflow).Name] = w.(*workflow.Workflow)
		}
	default:
		return nil, errors.New("invalide workflow")
	}
	return workflows, nil
}

func resolveWorkflowFromFile(path string) (map[string]*workflow.Workflow, error) {
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

	workflows := map[string]*workflow.Workflow{}
	for _, v := range m {
		bytes, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		w, err := workflow.New(bytes, "example")
		if err != nil {
			return nil, err
		}
		workflows[w.Name] = w
	}

	return workflows, nil
}
