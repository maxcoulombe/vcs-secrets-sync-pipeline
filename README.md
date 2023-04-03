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
  "public_tenant_id": "my-public-tenant-id", 
  "private_tenant_id": "my-private-tenant-id", 
  "app_name": "my-app", 
  "secret_name": "my-secret",
  "secret_token": "my-secret-token",
  "integration_type": "aws-sm", 
  "integration_token": "my-integration-token", 
  "operation": "load"
}'
```

## Teardown

```
chmod +x destroy.sh
./destroy.sh
```