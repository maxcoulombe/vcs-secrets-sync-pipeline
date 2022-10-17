#!/bin/bash

echo -e "\nStarting deletion\n"

cd terraform
terraform init -input=false
terraform destroy -input=false -auto-approve

cd ../

echo -e "\nDone\n"