package cadence

import (
	"go.temporal.io/sdk/worker"
)

type (
	Register interface {
		RegisterWorkflow(interface{})
		RegisterActivity([]interface{})
		RegisterWorker()
	}
	registerImpl struct {
		temporal *Temporal
	}
)

func NewRegister(temporal *Temporal) Register {
	return &registerImpl{temporal}
}

func (r *registerImpl) RegisterWorkflow(w interface{}) {
	r.temporal.WorkerClient.RegisterWorkflow(w)
}

func (r *registerImpl) RegisterActivity(a []interface{}) {
	/*
	 * t := reflect.TypeOf(a)
	 * for i := 0; i < t.NumMethod(); i++ {
	 *     fmt.Println(t.Method(i).Name)
	 *     r.temporal.WorkerClient.RegisterActivity(t.Method(i).Func.Interface())
	 * }
	 */
	for _, f := range a {
		r.temporal.WorkerClient.RegisterActivity(f)
	}
}

func (r *registerImpl) RegisterWorker() {
	go r.temporal.WorkerClient.Run(worker.InterruptCh())
}
