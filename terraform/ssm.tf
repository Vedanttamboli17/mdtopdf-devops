# SSM Parameter Store — worker instances fetch these on startup via IAM role.
# All parameters use SecureString (encrypted at rest with the AWS-managed SSM key).

resource "aws_ssm_parameter" "aws_region" {
  name  = "/docplatform/aws-region"
  type  = "SecureString"
  value = var.aws_region
  tags  = { Project = var.project_name }
}

resource "aws_ssm_parameter" "s3_bucket" {
  name  = "/docplatform/s3-bucket"
  type  = "SecureString"
  value = aws_s3_bucket.main.bucket
  tags  = { Project = var.project_name }
}

resource "aws_ssm_parameter" "dynamodb_table" {
  name  = "/docplatform/dynamodb-table"
  type  = "SecureString"
  value = aws_dynamodb_table.jobs.name
  tags  = { Project = var.project_name }
}

resource "aws_ssm_parameter" "sqs_queue_url" {
  name  = "/docplatform/sqs-queue-url"
  type  = "SecureString"
  value = aws_sqs_queue.main.url
  tags  = { Project = var.project_name }
}

resource "aws_ssm_parameter" "docx_sqs_queue_url" {
  name  = "/docplatform/docx-sqs-queue-url"
  type  = "SecureString"
  value = aws_sqs_queue.docx_queue.url
  tags  = { Project = var.project_name }
}

resource "aws_ssm_parameter" "thumb_sqs_queue_url" {
  name  = "/docplatform/thumb-sqs-queue-url"
  type  = "SecureString"
  value = aws_sqs_queue.thumb_queue.url
  tags  = { Project = var.project_name }
}

resource "aws_ssm_parameter" "notification_sqs_queue_url" {
  name  = "/docplatform/notification-sqs-queue-url"
  type  = "SecureString"
  value = aws_sqs_queue.notification_queue.url
  tags  = { Project = var.project_name }
}
