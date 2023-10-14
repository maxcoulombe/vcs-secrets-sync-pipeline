# vcs-secrets-sync-pipeline

## Setup

```
chmod +x deploy.sh
./deploy.sh
```

## Invoke

```
aws sqs send-message \
--queue-url my-queue-url \
--message-group-id test \
--message-deduplication-id $(echo $RANDOM | md5sum | head -c 20; echo) \
--message-body '{
  "group_id": "some-id"
}'
```

## Teardown

```
chmod +x destroy.sh
./destroy.sh
```