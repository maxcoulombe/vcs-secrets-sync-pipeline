# Queues

resource "aws_sqs_queue" "vcs_secrets_sync" {
  name       = "${local.name}.fifo"
  fifo_queue = true

  visibility_timeout_seconds  = local.timeout
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
  name       = "${local.name}-deadletter.fifo"
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
  batch_size       = 1
}

# Producer access

resource "aws_iam_user" "vcs_secrets_sync_message_producer" {
  name = "vcs-secrets-sync-message-producer"
}

resource "aws_iam_user_policy" "vcs_secrets_sync_message_producer" {
  name = "cs-secrets-sync-message-producer-policy"
  user = aws_iam_user.vcs_secrets_sync_message_producer.name

  policy = data.aws_iam_policy_document.vcs_secrets_sync_message_producer.json
}

data "aws_iam_policy_document" "vcs_secrets_sync_message_producer" {
  statement {
    actions = [
      "sqs:SendMessage",
    ]
    resources = [
      aws_sqs_queue.vcs_secrets_sync.arn,
    ]
  }
}

resource "aws_iam_access_key" "vcs_secrets_sync_message_producer" {
  user = aws_iam_user.vcs_secrets_sync_message_producer.name
}
