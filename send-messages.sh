for i in {1..1000}
do
  aws sqs send-message --queue-url $SQS_URL --message-group-id error --message-deduplication-id $(echo $RANDOM | md5sum | head -c 20; echo) --message-body '{"group_id": "error"}'
done
