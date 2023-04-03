locals {
  name    = "vcs-secrets-sync-pipeline"
  timeout = 30
}

# Executable

data "archive_file" "vcs_secrets_sync_pipeline" {
  type        = "zip"
  source_file = "../lambda/bin/${local.name}"
  output_path = "./${local.name}.zip"
}

# Function

resource "aws_lambda_function" "vcs_secrets_sync" {
  function_name    = local.name
  handler          = local.name
  description      = "Synchronize secrets between VCS and external secrets stores"
  filename         = data.archive_file.vcs_secrets_sync_pipeline.output_path
  source_code_hash = filebase64sha256(data.archive_file.vcs_secrets_sync_pipeline.output_path)
  role             = aws_iam_role.sync_pipeline_lambda.arn
  runtime          = "go1.x"
  timeout          = local.timeout
}

# Access

resource "aws_iam_role" "sync_pipeline_lambda" {
  name               = "${local.name}-lambda"
  assume_role_policy = data.aws_iam_policy_document.sync_pipeline_lambda_assume_role.json
}

data "aws_iam_policy_document" "sync_pipeline_lambda_assume_role" {
  policy_id = "${local.name}-lambda-assume-role"
  version   = "2012-10-17"
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole"
    ]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "sync_pipeline_logs" {
  policy_id = "${local.name}-lambda-logs"
  version   = "2012-10-17"
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = [
      "*"
    ]
  }
}

resource "aws_iam_policy" "sync_pipeline_logs" {
  name   = "${local.name}-lambda-logs"
  policy = data.aws_iam_policy_document.sync_pipeline_logs.json
}

resource "aws_iam_role_policy_attachment" "sync_pipeline_logs" {
  role       = aws_iam_role.sync_pipeline_lambda.name
  policy_arn = aws_iam_policy.sync_pipeline_logs.arn
}

data "aws_iam_policy_document" "sync_pipeline_sqs" {
  policy_id = "${local.name}-lambda-sqs"
  version   = "2012-10-17"

  statement {
    actions = [
      "sqs:GetQueueAttributes",
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
    ]

    resources = [
      "*"
    ]
  }
}

resource "aws_iam_policy" "sync_pipeline_sqs" {
  name   = "${local.name}-lambda-sqs"
  policy = data.aws_iam_policy_document.sync_pipeline_sqs.json
}

resource "aws_iam_role_policy_attachment" "sync_pipeline_sqs" {
  role       = aws_iam_role.sync_pipeline_lambda.name
  policy_arn = aws_iam_policy.sync_pipeline_sqs.arn
}