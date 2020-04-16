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
			func(t *testing.T, outputs ModuleOutputs, testData interface{}, workingDir string) {
				assert.True(t, setupCalled)
			},
		},
		Setup: func(t *testing.T, testData interface{}, workingDir string) {
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
			func(t *testing.T, outputs ModuleOutputs, testData interface{}, workingDir string) {
				assert.True(t, tearDownCalled)
			},
		},
		TearDown: func(t *testing.T, outputs ModuleOutputs, testData interface{}, workingDir string) {
			tearDownCalled = true
		},
	}

	UnitTest(t, testConfig)
}

func validateOutput(t *testing.T, outputs ModuleOutputs, testData interface{}, workingDir string) {
	assert.Equal(t, "Hello, World!", outputs["hello_world"])
}

type DynamoTestData struct {
	TableName string
	Region    string
}

func TestUnitTestDynamoDBExample(t *testing.T) {
	t.Parallel()

	testData := &DynamoTestData{
		TableName: fmt.Sprintf("terratest-aws-dynamodb-example-table-%s", random.UniqueId()),
		Region:    aws.GetRandomStableRegion(t, nil, nil),
	}

	workingDirectory := "../examples/terraform-aws-dynamodb-example"

	testConfig := &UnitTestConfig{
		WorkingDirectory: workingDirectory,
		TerraformOptions: &terraform.Options{

			// The path to where our Terraform code is located
			TerraformDir: "../examples/terraform-aws-dynamodb-example",

			// Variables to pass to our Terraform code using -var options
			Vars: map[string]interface{}{
				"table_name": testData.TableName,
			},

			// Environment variables to set when running Terraform
			EnvVars: map[string]string{
				"AWS_DEFAULT_REGION": testData.Region,
			},
		},
		Validators: []Validator{
			validateDynamoDB,
		},
		TestData: testData,
	}

	UnitTest(t, testConfig)
}

func validateDynamoDB(t *testing.T, outputs ModuleOutputs, dynamoTestData interface{}, workingDir string) {
	testData := dynamoTestData.(*DynamoTestData)

	expectedKmsKeyArn := aws.GetCmkArn(t, testData.Region, "alias/aws/dynamodb")
	expectedKeySchema := []*dynamodb.KeySchemaElement{
		{AttributeName: awsSDK.String("userId"), KeyType: awsSDK.String("HASH")},
		{AttributeName: awsSDK.String("department"), KeyType: awsSDK.String("RANGE")},
	}
	expectedTags := []*dynamodb.Tag{
		{Key: awsSDK.String("Environment"), Value: awsSDK.String("production")},
	}

	table := aws.GetDynamoDBTable(t, testData.Region, testData.TableName)

	assert.Equal(t, "ACTIVE", awsSDK.StringValue(table.TableStatus))
	assert.ElementsMatch(t, expectedKeySchema, table.KeySchema)

	// Verify server-side encryption configuration
	assert.Equal(t, expectedKmsKeyArn, awsSDK.StringValue(table.SSEDescription.KMSMasterKeyArn))
	assert.Equal(t, "ENABLED", awsSDK.StringValue(table.SSEDescription.Status))
	assert.Equal(t, "KMS", awsSDK.StringValue(table.SSEDescription.SSEType))

	// Verify TTL configuration
	ttl := aws.GetDynamoDBTableTimeToLive(t, testData.Region, testData.TableName)
	assert.Equal(t, "expires", awsSDK.StringValue(ttl.AttributeName))
	assert.Equal(t, "ENABLED", awsSDK.StringValue(ttl.TimeToLiveStatus))

	// Verify resource tags
	tags := aws.GetDynamoDbTableTags(t, testData.Region, testData.TableName)
	assert.ElementsMatch(t, expectedTags, tags)
}
