---
layout: collection-browser-doc
title: Unit tests, integration tests, end-to-end tests
category: testing-best-practices
excerpt: >-
  See the talk about unit tests, integration tests, end-to-end tests, dependency injection, test parallelism, retries, error handling, and static analysis.
tags: ["testing-best-practices"]
order: 201
nav_title: Documentation
nav_title_link: /docs/
---


For an introduction to Terratest, including unit tests, integration tests, end-to-end tests, dependency injection, test
parallelism, retries, error handling, and static analysis, see the talk "Automated Testing for Terraform, Docker,
Packer, Kubernetes, and More".

<iframe width="100%" height="450" allowfullscreen src="https://www.youtube.com/embed/xhHOW0EF5u8"></iframe>

Link to the video at [infoq.com](https://www.infoq.com/presentations/automated-testing-terraform-docker-packer/).

## Slides

Slides to the video can be found here: [Slides: How to test infrastructure code](https://www.slideshare.net/brikis98/how-to-test-infrastructure-code-automated-testing-for-terraform-kubernetes-docker-packer-and-more){:target="_blank"}.

## Unit test helpers

The function terraform.UnitTest will help unit test a single module, the UnitTest function will do the following things

**_TODO_**: As this is a one day project I will only attempt to do step 3, 4, and 6 today. Time permitting I will add 2
as well 

1. Copy your terraform code to a temporary folder
2. Optionally call a setup function you provide
3. Init and Apply the terraform code
4. Call a list of validators you provide
5. Optionally call a tearDown function you provide
6. Destroy the terraform code

Most of these steps are wrapped into stages which you can skip (see [Iterating locally using test stages]({% link _docs/02_testing-best-practices/iterating-locally-using-test-stages.md %})). This means you can skip individual stages
by setting env var SKIP_\<stage_name\> to any value.

If you skip the init_apply step then UnitTest will automatically load your previously applied outputs from the working directory.

It's common to need to pass data around between validators (for example the awsRegion, or random names you generated to
namespace that test run), so the validation function also receives an arg testData (which is defined as the [empty interface](https://tour.golang.org/methods/14)). You can define your own struct to pass data around
and cast this argument back to your struct to ensure type safety and give you access to shared data between tests.

The full list of stages is:

* setup
* init_apply
* validate
* teardown
* destroy

Signature of UnitTest:

```go
// The output of the module after it has been applied, as returned by terraform.OutputAll
type ModuleOutputs map[string]interface{}

// A validation function to test your terraform code
type Validator func(t *testing.T, moduleOutputs terraform.ModuleOutputs, testData interface{}, workingDir string)

type UnitTestConfig struct {
  TerraformOptions *terraform.Options
  // If you do not provide a working directory it will default to ./.terratest-unit-test/TEST_FUNCTION_NAME/
  WorkingDirectory string
  Setup func(t *testing.T, workingDir string)
  TearDown func(t *testing.T, moduleOutputs terraform.ModuleOutputs, workingDir string)
  Validators []Validator
  TestData interface{}
}

func UnitTest(t *testing.T, testConfig *terraform.UnitTestConfig)
```

Examples:

This example shows the simplest example with multiple Validators, with the default working directory
```go
func TestWebServer(t *testing.T) {
  terraformOptions := &terraform.Options {
    // The path to where your Terraform code is located
    TerraformDir: "../web-server",
  }

  terrform.UnitTest(&terraform.UnitTestConfig{
    TerraformOptions: terraformOptions,
    Validators: []terraform.Validator{
      // A list of validation functions to test your module
      validateWebServer,
      validateImageHosting,
      validateSSLApplied,
    },
  })
}

func validateWebServer(t *testing.T, outputs terraform.ModuleOutputs, testDataStruct interface{}, workingDir string) {
}
```

This example shows using Setup and Validators, and passing some testData around, with a custom working directory
```go
type TestData struct {
  Name string
  AwsRegion string
}

func TestWebServer(t *testing.T) {
  testData := &TestData{
    Name: random.UniqueId()
    AwsRegion: aws.GetRandomStableRegion(t, nil, nil)
  }

  terraformOptions := &terraform.Options {
    // The path to where your Terraform code is located
    TerraformDir: "../web-server",

    Vars: map[string]interface{}{
      "Name": testData.Name
    }

    EnvVars: map[string]string{
      "AWS_DEFAULT_REGION": testData.AwsRegion,
    },
  }

  terrform.UnitTest(&terraform.UnitTestConfig{
    TerraformOptions: terraformOptions,
    WorkingDirectory "../web-server",
    Setup: setup,
    TestData: testData
    Validators: []terraform.Validator{
      // A list of validation functions to test your module
      validateWebServer,
      validateImageHosting,
      validateSSLApplied,
    },
  })
}

func validateWebServer(t *testing.T, outputs terraform.ModuleOutputs, testDataStruct interface{}, workingDir string) {
  testData := testDataStruct.(*TestData)

  randomName := testData.Name
  ...
}

func setup(t *testing.T, workingDir string) {
  ...
}
```


## Integration test helpers

**_TODO: As this is a one day project I won't attempt to complete this today _**

Integration tests are very similar to unit tests, only you are testing multiple modules together. To facilitate this
terraform provides an IntegrationTest function, the main difference to the UnitTest is the IntegrationTestConfig takes a
list of modules, each with a unique name, and it calls a callback function in your code to create the terraform options
as they are needed. All previously applied modules will have their outputs passed to this function so you can use
outputs from modules applied earlier in the construction of your terraformOptions. It passes a map of ModuleOutputs 
which are keyed on the unique name you gave the module.

The destroys will be performed in the reverse order to applies.

The working directory will contain sub directories for each module (named by your unique name) to save the
terraformOptions and outputs to.

If you skip the init_apply step then IntegrationTest will automatically load your previously applied outputs from the working directory.

Signature of IntegrationTest:

```go
// Output of every module applied, by name
type AllModuleOutputs map[string]ModuleOutputs

type ModuleUnderTest struct {
  Name  string
  OptionsGenerator func(t *testing.T, allModuleOutputs *terraform.AllModuleOutputs, workingDir string) *terraform.Options
}

IntegrationTest(t *testing.T, testConfig *terraform.IntegrationTestConfig

// A validation function to test your terraform code
type IntegrationTestValidator func(t *testing.T, allModuleOutputs terraform.AllModuleOutputs, workingDir string)

type IntegrationTestConfig struct {
  Modules []*ModuleUnderTest
  // If you do not provide a working directory it will default to ./.terratest-integration-test/TEST_FUNCTION_NAME/
  WorkingDirectory string
  Setup func(t *testing.T, workingDir string)
  TearDown func(t *testing.T, allModuleOutputs terraform.AllModuleOutputs, workingDir string)
  Validators []IntegrationTestValidator
}
```
