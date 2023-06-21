resource "aws_dynamodb_table" "hollywood_theater" {
  name = "hollywood-theater"
  attribute {
    name = "id"
    type = "S"
  }
  hash_key = "id"
}
