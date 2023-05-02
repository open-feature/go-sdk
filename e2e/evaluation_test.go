package e2e_test

import (
	"testing"

	"github.com/cucumber/godog"
	e2e "github.com/open-feature/go-sdk-contrib/tests/flagd/pkg/integration"
)

func TestEvaluation(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite := godog.TestSuite{
		Name:                "evaluation.feature",
		ScenarioInitializer: e2e.InitializeEvaluationScenario(),
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../test-harness/features/evaluation.feature"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run evaluation tests")
	}
}
