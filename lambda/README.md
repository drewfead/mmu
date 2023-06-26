##Deploying lambda locally to AWS using Terraform
- Navigate to the lambda/scraper-lambda directory and execute the command `make` (on Linux/Mac; if you're on Windows yer shit outta luck right now).  This will build the go executable and then create a zip file called lambda-handler.zip which will be deployed to AWS.
- Navigate to the terraform directory and execute the following commands:
```
terraform init
terraform plan
terraform apply
```
    - **NOTE**: you will need to have the AWS cli set up to run these commands.
