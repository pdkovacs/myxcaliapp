resource "aws_apigatewayv2_api" "xcaliapp_prod" {
  name          = "xcaliapp_prod"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_route" "listdrawings" {
  api_id    = aws_apigatewayv2_api.xcaliapp_prod.id
  route_key = "GET /drawing/{proxy+}"

  target = "integrations/${aws_apigatewayv2_integration.listdrawings.id}"
}

resource "aws_apigatewayv2_integration" "listdrawings" {
  api_id           = aws_apigatewayv2_api.xcaliapp_prod.id
  integration_type = "AWS_PROXY"

  connection_type      = "INTERNET"
  description          = "prod-lambda"
  integration_method   = "POST"
  integration_uri      = aws_lambda_function.xcali-prod[index(local.function_names, "listdrawings")].invoke_arn
  passthrough_behavior = "WHEN_NO_MATCH"
}

resource "aws_apigatewayv2_route" "getdrawing" {
  api_id    = aws_apigatewayv2_api.xcaliapp_prod.id
  route_key = "GET /drawing/{title}/{proxy+}"

  target = "integrations/${aws_apigatewayv2_integration.getdrawing.id}"
}

resource "aws_apigatewayv2_integration" "getdrawing" {
  api_id           = aws_apigatewayv2_api.xcaliapp_prod.id
  integration_type = "AWS_PROXY"

  connection_type      = "INTERNET"
  description          = "prod-lambda"
  integration_method   = "POST"
  integration_uri      = aws_lambda_function.xcali-prod[index(local.function_names, "getdrawing")].invoke_arn
  passthrough_behavior = "WHEN_NO_MATCH"
}

resource "aws_apigatewayv2_route" "putdrawing" {
  api_id    = aws_apigatewayv2_api.xcaliapp_prod.id
  route_key = "PUT /drawing/{title}/{proxy+}"

  target = "integrations/${aws_apigatewayv2_integration.putdrawing.id}"
}

resource "aws_apigatewayv2_integration" "putdrawing" {
  api_id           = aws_apigatewayv2_api.xcaliapp_prod.id
  integration_type = "AWS_PROXY"

  connection_type      = "INTERNET"
  description          = "prod-lambda"
  integration_method   = "POST"
  integration_uri      = aws_lambda_function.xcali-prod[index(local.function_names, "putdrawing")].invoke_arn
  passthrough_behavior = "WHEN_NO_MATCH"
}

resource "aws_apigatewayv2_stage" "xcaliapp_prod" {
  name   = "$default"
  api_id = aws_apigatewayv2_api.xcaliapp_prod.id

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.xcaliapp_prodapigw.arn
    format          = "$context.identity.sourceIp - - [$context.requestTime] \"$context.httpMethod $context.routeKey $context.protocol\" $context.status $context.responseLength $context.requestId $context.integrationErrorMessage"
  }

  auto_deploy = true

  default_route_settings {
    throttling_burst_limit = 2
    throttling_rate_limit  = 1
  }

  # depends_on = [
  #   aws_apigatewayv2_route.xcaliapp_prod
  # ]
}


resource "aws_cloudwatch_log_group" "xcaliapp_prodapigw" {
  name = "/aws/apigw/xcaliapp_prod"
}

resource "aws_iam_role" "xcaliapp_prodapigw_cloudwatch" {
  name = "xcaliapp_prodapigw_cloudwatch"

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

resource "aws_iam_role_policy" "xcaliapp_prodapigw_cloudwatch" {
  name = "xcaliapp_prodapigw_cloudwatch"
  role = aws_iam_role.xcaliapp_prodapigw_cloudwatch.id

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

resource "aws_api_gateway_account" "xcaliapp_prod" {
  cloudwatch_role_arn = aws_iam_role.xcaliapp_prodapigw_cloudwatch.arn
}

output "base_url" {
  value = "${aws_apigatewayv2_api.xcaliapp_prod.id}.execute-api.eu-west-1.amazonaws.com"
}
