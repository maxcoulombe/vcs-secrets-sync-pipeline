#!/bin/bash

echo -e "\nStarting deployment\n"
rm -rf ./bin

cd lambda
go test ./...
env GOOS=linux GOARCH=amd64 go build -o ./bin/hack-week-lambda
cd ../

cd terraform
terraform init -input=false
terraform apply -input=false -auto-approve

cd ../

echo -e "\nDone\n"