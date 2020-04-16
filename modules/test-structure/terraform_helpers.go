package test_structure

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
)

// The output of the module after it has been applied, as returned by terraform.OutputAll
type ModuleOutputs map[string]interface{}

// A validation function to test your terraform code
type Validator func(t *testing.T, moduleOutputs ModuleOutputs, testData interface{}, workingDir string)

type UnitTestConfig struct {
	TerraformOptions *terraform.Options
	// If you do not provide a working directory it will default to ./.terratest-unit-test/TEST_FUNCTION_NAME/
	WorkingDirectory string
	Setup            func(t *testing.T, testData interface{}, workingDir string)
	TearDown         func(t *testing.T, moduleOutputs ModuleOutputs, testData interface{}, workingDir string)
	Validators       []Validator
	TestData         interface{}
}

func UnitTest(t *testing.T, testConfig *UnitTestConfig) {
	// TODO: Load the terraform options, and test data if init_apply is skipped
	// Probably best to factor out a method in test_structure to do the check and reuse that
	// method here rather than reimplementing

	// If the client hasn't provided a working dir make one ourselves
	workingDir := testConfig.WorkingDirectory
	if workingDir == "" {
		workingDir = filepath.Join(".terratest-unit-test", t.Name())
	}

	defer RunTestStage(t, "destroy", func() {
		options := LoadTerraformOptions(t, workingDir)
		terraform.Destroy(t, options)
	})

	// TearDown will be nil if they haven't set it
	if testConfig.TearDown != nil {
		defer RunTestStage(t, "teardown", func() {
			options := LoadTerraformOptions(t, workingDir)
			outputs := terraform.OutputAll(t, options)
			testConfig.TearDown(t, outputs, testConfig.TestData, workingDir)
		})
	}

	// Setup will be nil if they haven't set it
	if testConfig.Setup != nil {
		RunTestStage(t, "setup", func() {
			testConfig.Setup(t, testConfig.TestData, workingDir)
		})
	}

	RunTestStage(t, "init_apply", func() {
		SaveTerraformOptions(t, workingDir, testConfig.TerraformOptions)
		terraform.InitAndApply(t, testConfig.TerraformOptions)
	})

	RunTestStage(t, "validate", func() {
		outputs := terraform.OutputAll(t, testConfig.TerraformOptions)

		for _, validator := range testConfig.Validators {
			validator(t, outputs, testConfig.TestData, "")
		}
	})
}
