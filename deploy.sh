#!/bin/bash

echo -e "\nStarting deployment\n"
rm -rf ./bin

cd lambda
go test ./...
go mod tidy
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/vcs-secrets-sync-pipeline
cd ../

cd terraform
terraform init -input=false
terraform apply -input=false -auto-approve

cd ../

echo -e "\nDone\n"