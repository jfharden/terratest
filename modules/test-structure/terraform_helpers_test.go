package test_structure

import (
	"fmt"
	"testing"

	awsSDK "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestUnitTestSimple(t *testing.T) {
	testConfig := &UnitTestConfig{
		TerraformOptions: &terraform.Options{
			TerraformDir: "../../examples/terraform-hello-world-example/",
		},
		Validators: []Validator{
			validateOutput,
		},
	}

	UnitTest(t, testConfig)
}

func TestUnitTestSetup(t *testing.T) {
	setupCalled := false

	testConfig := &UnitTestConfig{
		TerraformOptions: &terraform.Options{
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

	UnitTest(t, testConfig)
}

func TestUnitTestTeardown(t *testing.T) {
	tearDownCalled := false

	testConfig := &UnitTestConfig{
		TerraformOptions: &terraform.Options{
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

	UnitTest(t, testConfig)
}

func validateOutput(t *testing.T, outputs ModuleOutputs, workingDir string) {
	assert.Equal(t, "Hello, World!", outputs["hello_world"])
}

func TestUnitTestDynamoDBExample(t *testing.T) {
	t.Parallel()

	awsRegion := "eu-west-1"
	expectedTableName := fmt.Sprintf("terratest-aws-dynamodb-example-table-%s", random.UniqueId())

	workingDirectory := "../examples/terraform-aws-dynamodb-example"

	SaveString(t, workingDirectory, "region", awsRegion)
	SaveString(t, workingDirectory, "table_name", awsRegion)

	testConfig := &UnitTestConfig{
		WorkingDirectory: workingDirectory,
		TerraformOptions: &terraform.Options{

			// The path to where our Terraform code is located
			TerraformDir: "../examples/terraform-aws-dynamodb-example",

			// Variables to pass to our Terraform code using -var options
			Vars: map[string]interface{}{
				"table_name": expectedTableName,
			},

			// Environment variables to set when running Terraform
			EnvVars: map[string]string{
				"AWS_DEFAULT_REGION": awsRegion,
			},
		},
		Validators: []Validator{
			validateDynamoDB,
		},
	}

	UnitTest(t, testConfig)
}

func validateDynamoDB(t *testing.T, outputs ModuleOutputs, workingDir string) {
	awsRegion := LoadString(t, workingDir, "region")
	expectedTableName := LoadString(t, workingDir, "table_name")

	expectedKmsKeyArn := aws.GetCmkArn(t, awsRegion, "alias/aws/dynamodb")
	expectedKeySchema := []*dynamodb.KeySchemaElement{
		{AttributeName: awsSDK.String("userId"), KeyType: awsSDK.String("HASH")},
		{AttributeName: awsSDK.String("department"), KeyType: awsSDK.String("RANGE")},
	}
	expectedTags := []*dynamodb.Tag{
		{Key: awsSDK.String("Environment"), Value: awsSDK.String("production")},
	}

	table := aws.GetDynamoDBTable(t, awsRegion, expectedTableName)

	assert.Equal(t, "ACTIVE", awsSDK.StringValue(table.TableStatus))
	assert.ElementsMatch(t, expectedKeySchema, table.KeySchema)

	// Verify server-side encryption configuration
	assert.Equal(t, expectedKmsKeyArn, awsSDK.StringValue(table.SSEDescription.KMSMasterKeyArn))
	assert.Equal(t, "ENABLED", awsSDK.StringValue(table.SSEDescription.Status))
	assert.Equal(t, "KMS", awsSDK.StringValue(table.SSEDescription.SSEType))

	// Verify TTL configuration
	ttl := aws.GetDynamoDBTableTimeToLive(t, awsRegion, expectedTableName)
	assert.Equal(t, "expires", awsSDK.StringValue(ttl.AttributeName))
	assert.Equal(t, "ENABLED", awsSDK.StringValue(ttl.TimeToLiveStatus))

	// Verify resource tags
	tags := aws.GetDynamoDbTableTags(t, awsRegion, expectedTableName)
	assert.ElementsMatch(t, expectedTags, tags)
}
