# vault-hack-week-lambda

## Setup

```
chmod +x deploy.sh
./deploy.sh
```

## Invoke

```
aws lambda invoke --function-name hack-week-lambda --payload '{}' response.json
```

## Teardown

```
chmod +x destroy.sh
./destroy.sh
```