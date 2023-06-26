## Deploying lambda to AWS using Terraform locally
- Navigate to the lambda/scraper-lambda directory and execute the command `make` (on Linux/Mac; local makefile build doesn't work for Windows at the moment).  This will build the go executable and then create a zip file called lambda-handler.zip which will be deployed to AWS.
- Navigate to the terraform directory and execute the following commands:
```
terraform init
terraform plan
terraform apply
```
**NOTE**: you will need to have the AWS cli set up to run these commands.
