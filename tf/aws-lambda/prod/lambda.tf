locals {
  function_names    = ["listdrawings", "getdrawing", "putdrawing"]
  archive_filenames = [for f in local.function_names : "${f}.zip"]
  impl_directories  = [for f in local.function_names : "${path.module}/../../../aws-lambda/${f}"]
}

data "aws_s3_bucket" "store" {
  bucket = var.s3_bucket
}

resource "aws_iam_policy" "read_write_s3_bucket" {
  name = "xcaliapp_read_write_s3_bucket"
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "s3listbucket",
        "Effect" : "Allow",
        "Action" : [
          "s3:ListBucket",
        ],
        "Resource" : [
          "arn:aws:s3:::${var.s3_bucket}"
        ]
      },
      {
        "Sid" : "s3readwritebucketobjects",
        "Effect" : "Allow",
        "Action" : [
          "s3:*Object",
        ],
        "Resource" : [
          "arn:aws:s3:::${var.s3_bucket}/*"
        ]
      }
    ]
  })
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_xcaliapp_lambda"

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
  count         = length(local.function_names)
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.xcali-prod[count.index].function_name
  principal     = "apigateway.amazonaws.com"

  # The /*/* portion grants access from any method on any resource
  # within the API Gateway "REST API".
  source_arn = "${aws_apigatewayv2_api.xcaliapp_prod.execution_arn}/*/*"
}

data "archive_file" "lambda" {
  count       = length(local.function_names)
  type        = "zip"
  source_file = "${local.impl_directories[count.index]}/bootstrap"
  output_path = "${local.impl_directories[count.index]}/${local.archive_filenames[count.index]}"
}

resource "aws_lambda_function" "xcali-prod" {
  count         = length(local.function_names)
  function_name = local.function_names[count.index]
  architectures = ["arm64"]
  filename      = "${local.impl_directories[count.index]}/${local.archive_filenames[count.index]}"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"

  source_code_hash = data.archive_file.lambda[count.index].output_base64sha256

  runtime = "provided.al2023"

  environment {
    variables = {
      DRAWINGS_BUCKET_NAME = var.s3_bucket
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_logs,
    # aws_cloudwatch_log_group.xcali-prod[count.index],
  ]
}

resource "aws_cloudwatch_log_group" "xcali-prod" {
  count             = length(local.function_names)
  name              = "/aws/lambda/${local.function_names[count.index]}"
  retention_in_days = 14
}

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
  name        = "xcaliapp_lambda_logging"
  path        = "/"
  description = "IAM policy for logging from a lambda"
  policy      = data.aws_iam_policy_document.lambda_logging.json
}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}

