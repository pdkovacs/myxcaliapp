package main

import (
	awslambda "xcaliapp/aws-lambda"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(awslambda.HandleEcho)
}
