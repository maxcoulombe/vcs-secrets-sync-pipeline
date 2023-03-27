locals {
  lambda_handler = "vcs-secrets-sync-pipeline"
  name           = "vcs-secrets-sync-pipeline"
  region         = "us-east-1"
}

data "archive_file" "vcs_secrets_sync_pipeline" {
  type        = "zip"
  source_file = "../lambda/bin/vcs-secrets-sync-pipeline"
  output_path = "./vcs-secrets-sync-pipeline.zip"
}

resource "aws_lambda_function" "vcs_secrets_sync" {
  filename         = data.archive_file.vcs_secrets_sync_pipeline.output_path
  function_name    = local.name
  role             = aws_iam_role.sync_pipeline_lambda.arn
  handler          = local.lambda_handler
  source_code_hash = filebase64sha256(data.archive_file.vcs_secrets_sync_pipeline.output_path)
  runtime          = "go1.x"
  memory_size      = 1024
  timeout          = 30
}
