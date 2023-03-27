resource "aws_sqs_queue" "vcs_secrets_sync" {
  name       = "vcs-secrets-sync.fifo"
  fifo_queue = true

  visibility_timeout_seconds  = 5
  content_based_deduplication = false
  sqs_managed_sse_enabled     = true
  deduplication_scope         = "messageGroup"
  fifo_throughput_limit       = "perMessageGroupId"

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.vcs_secrets_sync_deadletter.arn
    maxReceiveCount     = 3
  })
}

resource "aws_sqs_queue" "vcs_secrets_sync_deadletter" {
  name       = "vcs-secrets-sync-deadletter.fifo"
  fifo_queue = true
}

resource "aws_sqs_queue_redrive_allow_policy" "vcs_secrets_sync" {
  queue_url = aws_sqs_queue.vcs_secrets_sync_deadletter.id

  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.vcs_secrets_sync.arn]
  })
}

resource "aws_lambda_event_source_mapping" "vcs_secrets_sync" {
  event_source_arn = aws_sqs_queue.vcs_secrets_sync.arn
  function_name    = aws_lambda_function.vcs_secrets_sync.arn
}