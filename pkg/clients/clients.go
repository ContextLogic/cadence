package clients

import (
	"github.com/ContextLogic/cadence/pkg/clients/temporal"
)

var Temporal *temporal.Temporal

func init() {
	var err error
	Temporal, err = temporal.New()
	if err != nil {
		panic(err)
	}
}
