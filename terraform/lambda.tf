provider "aws" {}

locals {
  lambda_handler  = "hack-week-lambda"
  name            = "hack-week-lambda"
  region          = "us-east-1"
}

data "archive_file" "hack_week_lambda" {
  type        = "zip"
  source_file = "../lambda/bin/hack-week-lambda"
  output_path = "./hack-week-lambda.zip"
}

data "aws_iam_policy_document" "hack_week_lambda" {
  policy_id = "${local.name}-lambda"
  version   = "2012-10-17"
  statement {
    effect  = "Allow"
    actions = [
      "sts:AssumeRole"
    ]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "hack_week_lambda_mq" {
  policy_id = "${local.name}-lambd-mq"
  version   = "2012-10-17"

  statement {
    actions = [
      "mq:CreateUser",
      "mq:UpdateUser",
      "mq:DeleteUser",
      "mq:RebootBroker",
    ]

    resources = [aws_mq_broker.hack_week_active_mq.arn]
  }
}

resource "aws_iam_role" "hack_week_lambda" {
  name                = "${local.name}-lambda"
  assume_role_policy  = data.aws_iam_policy_document.hack_week_lambda.json
  inline_policy {
    name = "mq-access"
    policy = data.aws_iam_policy_document.hack_week_lambda_mq.json
  }
}

resource "aws_lambda_function" "hack_week_lambda" {
  filename          = data.archive_file.hack_week_lambda.output_path
  function_name     = local.name
  role              = aws_iam_role.hack_week_lambda.arn
  handler           = local.lambda_handler
  source_code_hash  = filebase64sha256(data.archive_file.hack_week_lambda.output_path)
  runtime           = "go1.x"
  memory_size       = 1024
  timeout           = 30
}
