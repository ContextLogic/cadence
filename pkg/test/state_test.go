package test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"

	st "github.com/ContextLogic/cadence/pkg/fsm/state"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

var succeedMachine = []byte(`
 {
			"Type": "Succeed"
		}

`)

func (s *UnitTestSuite) Test_Workflow_Succeed_State() {
	state, _ := st.NewSucceedState("test", succeedMachine)
	exampleInput := map[string]interface{}{"test": "example_input"}
	output, _, _ := state.Execute(nil, exampleInput)
	converted := output.(map[string]interface{})
	s.Equal("example_input", converted["test"])
}
