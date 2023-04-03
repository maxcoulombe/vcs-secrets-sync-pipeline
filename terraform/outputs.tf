output "vcs_secrets_sync_queue_arn" {
  value = aws_sqs_queue.vcs_secrets_sync.arn
}

output "access_key_id" {
  value = aws_iam_access_key.vcs_secrets_sync_message_producer.id
}

output "secret_access_key" {
  value     = aws_iam_access_key.vcs_secrets_sync_message_producer.secret
  sensitive = true
}