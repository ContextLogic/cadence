package loader

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	client "github.com/ContextLogic/cadence/pkg"
	"github.com/ContextLogic/cadence/pkg/fsm/workflow"
	"github.com/ContextLogic/cadence/pkg/options"
	sdk "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func LoadWorkflows(path string) []*workflow.Workflow {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	value, _ := ioutil.ReadAll(f)
	m := []map[string]interface{}{}
	if err := json.Unmarshal(value, &m); err != nil {
		panic(err)
	}

	workflows := []*workflow.Workflow{}
	for _, v := range m {
		bytes, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		w, err := workflow.New(bytes)
		if err != nil {
			panic(err)
		}
		workflows = append(workflows, w)
	}
	return workflows
}

func CreateClient(taskqueue_prefix string, queue string, hostport string, namespace string) (client.Client, error) {
	return client.New(options.Options{
		ClientOptions: sdk.Options{
			HostPort:  hostport,
			Namespace: namespace,
		},
		WorkerOptions: worker.Options{},
		TaskQueue:     strings.Join([]string{taskqueue_prefix, queue}, "_"),
	})
}
