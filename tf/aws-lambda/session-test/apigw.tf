resource "aws_apigatewayv2_api" "xcaliapp_session_test" {
  name          = "xcaliapp_session_test"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_route" "root" {
  api_id    = aws_apigatewayv2_api.xcaliapp_session_test.id
  route_key = "ANY /{proxy+}"

  target = "integrations/${aws_apigatewayv2_integration.session-test-lambda.id}"
}

resource "aws_apigatewayv2_integration" "session-test-lambda" {
  api_id           = aws_apigatewayv2_api.xcaliapp_session_test.id
  integration_type = "AWS_PROXY"

  connection_type      = "INTERNET"
  description          = "session-test-lambda"
  integration_method   = "POST"
  integration_uri      = aws_lambda_function.xcali-session-test.invoke_arn
  passthrough_behavior = "WHEN_NO_MATCH"
}

resource "aws_apigatewayv2_stage" "xcaliapp_session_test" {
  name   = "$default"
  api_id = aws_apigatewayv2_api.xcaliapp_session_test.id

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.xcaliapp_session_test_apigw.arn
    format          = "$context.identity.sourceIp - - [$context.requestTime] \"$context.httpMethod $context.routeKey $context.protocol\" $context.status $context.responseLength $context.requestId $context.integrationErrorMessage"
  }

  auto_deploy = true

  default_route_settings {
    throttling_burst_limit = 2
    throttling_rate_limit  = 1
  }

  # depends_on = [
  #   aws_apigatewayv2_route.xcaliapp_session_test
  # ]
}


resource "aws_cloudwatch_log_group" "xcaliapp_session_test_apigw" {
  name = "/aws/apigw/xcaliapp_session_test"
}

resource "aws_iam_role" "xcaliapp_session_test_apigw_cloudwatch" {
  name = "xcaliapp_session_test_apigw_cloudwatch"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "xcaliapp_session_test_apigw_cloudwatch" {
  name = "session_test_apigw_clouldwatch"
  role = aws_iam_role.xcaliapp_session_test_apigw_cloudwatch.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams",
        "logs:PutLogEvents",
        "logs:GetLogEvents",
        "logs:FilterLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_api_gateway_account" "xcaliapp_session_test" {
  cloudwatch_role_arn = aws_iam_role.xcaliapp_session_test_apigw_cloudwatch.arn
}

output "base_url" {
  value = "${aws_apigatewayv2_api.xcaliapp_session_test.id}.execute-api.eu-west-1.amazonaws.com"
}
