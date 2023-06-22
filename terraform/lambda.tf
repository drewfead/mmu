data "aws_iam_policy_document" "assume_lambda_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda" {
  name               = "AssumeLambdaRole"
  description        = "Role for lambda to assume lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_lambda_role.json
}

resource "aws_lambda_function" "theater_scraper" {
  filename         = "${path.module}/../lambda/bin/lambda-handler.zip"
  function_name    = "theater-scraper"
  role             = aws_iam_role.lambda.arn
  handler          = "mmu"
  source_code_hash = filebase64sha256("${path.module}/../lambda/bin/lambda-handler.zip")

  runtime = "go1.x"
  timeout = 20
}
