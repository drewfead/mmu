resource "aws_lambda_function" "hollywood_theater" {
  filename         = "${path.module}/../build/hollywoodtheater.zip"
  function_name    = "hollywood_theater_scraper"
  role             = aws_iam_role.lambda_role.arn
  handler          = "hollywod_theater_scraper.lambda_handler"
  source_code_hash = filebase64sha256("${path.module}/../build/hollywoodtheater.zip")

  runtime = "go1.x"

  environment {
    variables = {
    }
  }
}
