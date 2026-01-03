package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type PosterStackProps struct {
	awscdk.StackProps
}

func NewPosterStack(scope constructs.Construct, id string, props *PosterStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Lambda function
	fn := awslambda.NewFunction(stack, jsii.String("PosterGenerator"), &awslambda.FunctionProps{
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Architecture: awslambda.Architecture_ARM_64(),
		Code:         awslambda.Code_FromAsset(jsii.String("../build"), nil),
		MemorySize:   jsii.Number(512),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(30)),
		Environment: &map[string]*string{
			"GO_ENV": jsii.String("production"),
		},
	})

	// HTTP API Gateway
	httpApi := awsapigatewayv2.NewHttpApi(stack, jsii.String("PosterApi"), &awsapigatewayv2.HttpApiProps{
		ApiName: jsii.String("poster-api"),
		CorsPreflight: &awsapigatewayv2.CorsPreflightOptions{
			AllowOrigins: jsii.Strings("*"),
			AllowMethods: &[]awsapigatewayv2.CorsHttpMethod{
				awsapigatewayv2.CorsHttpMethod_POST,
				awsapigatewayv2.CorsHttpMethod_OPTIONS,
			},
			AllowHeaders: jsii.Strings("Content-Type"),
		},
	})

	// Lambda integration
	integration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("LambdaIntegration"),
		fn,
		nil,
	)

	// Add route
	httpApi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Path:        jsii.String("/generate"),
		Methods:     &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_POST},
		Integration: integration,
	})

	// Output the API URL
	awscdk.NewCfnOutput(stack, jsii.String("ApiUrl"), &awscdk.CfnOutputProps{
		Value:       httpApi.Url(),
		Description: jsii.String("API Gateway URL"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewPosterStack(app, "PosterStack", &PosterStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return nil // Use default account/region from AWS CLI config
}
