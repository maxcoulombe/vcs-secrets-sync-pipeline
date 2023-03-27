data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

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