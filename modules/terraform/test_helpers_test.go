package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitTestSimple(t *testing.T) {
	testPlan := &UnitTestPlan{
		TerraformOptions: &Options{
			TerraformDir: "../../examples/terraform-hello-world-example/",
		},
		Validators: []Validator{
			validateOutput,
		},
	}

	UnitTest(t, testPlan)
}

func TestUnitTestSetup(t *testing.T) {
	setupCalled := false

	testPlan := &UnitTestPlan{
		TerraformOptions: &Options{
			TerraformDir: "../../examples/terraform-hello-world-example/",
		},
		Validators: []Validator{
			validateOutput,
			func(t *testing.T, outputs ModuleOutputs, workingDir string) {
				assert.True(t, setupCalled)
			},
		},
		Setup: func(t *testing.T, workingDir string) {
			setupCalled = true
		},
	}

	UnitTest(t, testPlan)
}

func TestUnitTestTeardown(t *testing.T) {
	tearDownCalled := false

	testPlan := &UnitTestPlan{
		TerraformOptions: &Options{
			TerraformDir: "../../examples/terraform-hello-world-example/",
		},
		Validators: []Validator{
			validateOutput,
			func(t *testing.T, outputs ModuleOutputs, workingDir string) {
				assert.True(t, tearDownCalled)
			},
		},
		TearDown: func(t *testing.T, outputs ModuleOutputs, workingDir string) {
			tearDownCalled = true
		},
	}

	UnitTest(t, testPlan)
}

func validateOutput(t *testing.T, outputs ModuleOutputs, workingDir string) {
	assert.Equal(t, "Hello, World!", outputs["hello_world"])
}
