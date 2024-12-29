locals {
    s3_bucket         = "xcali-session-test"
		function_name     = "sessionTest"
    archive_filename  = "sessionTest.zip"
    impl_dir          = "${path.module}/../../../aws-lambda/sessiontest-lambda"
}

data "aws_s3_bucket" "store" {
	bucket = local.s3_bucket
}

resource "aws_iam_policy" "read_write_s3_bucket" {
	name = "read_write_s3_bucket"
	policy = jsonencode({
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "s3listbucket",
				"Effect": "Allow",
				"Action": [
					"s3:ListBucket",
				],
				"Resource": [
					"arn:aws:s3:::${local.s3_bucket}"
				]
			},
			{
				"Sid": "s3readwritebucketobjects",
				"Effect": "Allow",
				"Action": [
					"s3:*Object",
				],
				"Resource": [
					"arn:aws:s3:::${local.s3_bucket}/*"
				]
			}
		]
	})
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "basic_lambda_exec" {
    role       = aws_iam_role.iam_for_lambda.name
    policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "read_write_s3_bucket" {
    role       = aws_iam_role.iam_for_lambda.name
    policy_arn = aws_iam_policy.read_write_s3_bucket.arn
}

resource "aws_lambda_permission" "apigw" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.xcali-session-test.function_name}"
  principal     = "apigateway.amazonaws.com"

  # The /*/* portion grants access from any method on any resource
  # within the API Gateway "REST API".
  source_arn = "${aws_apigatewayv2_api.xcaliapp_session_test.execution_arn}/*/*"
}

data "archive_file" "lambda" {
  type        = "zip"
  source_file = "${local.impl_dir}/bootstrap"
  output_path = "${local.impl_dir}/${local.archive_filename}"
}

resource "aws_lambda_function" "xcali-session-test" {
  # If the file is not in the current working directory you will need to include a
  # path.module in the filename.
  function_name = local.function_name
  architectures = ["arm64"]
  filename      = "${local.impl_dir}/${local.archive_filename}"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"

  # The filebase64sha256() function is available in Terraform 0.11.12 and later
  # For Terraform 0.11.11 and earlier, use the base64sha256() function and the file() function:
  # source_code_hash = "${base64sha256(file("lambda_function_payload.zip"))}"
  source_code_hash = data.archive_file.lambda.output_base64sha256

  runtime = "provided.al2023"

  environment {
    variables = {
      DRAWINGS_BUCKET_NAME = local.s3_bucket
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_logs,
    aws_cloudwatch_log_group.xcali-session-test,
  ]
}

resource "aws_cloudwatch_log_group" "xcali-session-test" {
  name              = "/aws/lambda/${local.function_name}"
  retention_in_days = 14
}

# See also the following AWS managed policy: AWSLambdaBasicExecutionRole
data "aws_iam_policy_document" "lambda_logging" {
  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:aws:logs:*:*:*"]
  }
}

resource "aws_iam_policy" "lambda_logging" {
  name        = "lambda_logging"
  path        = "/"
  description = "IAM policy for logging from a lambda"
  policy      = data.aws_iam_policy_document.lambda_logging.json
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}