# ── DOCX Queue ────────────────────────────────────────────────
resource "aws_sqs_queue" "docx_dlq" {
  name                      = "${var.project_name}-docx-dlq"
  message_retention_seconds = 1209600
  tags                      = { Project = var.project_name }
}

resource "aws_sqs_queue" "docx_queue" {
  name                       = "${var.project_name}-docx-queue"
  visibility_timeout_seconds = 300
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 20

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.docx_dlq.arn
    maxReceiveCount     = 3
  })

  tags = { Project = var.project_name }
}

# ── Thumbnail Queue ───────────────────────────────────────────
resource "aws_sqs_queue" "thumb_dlq" {
  name                      = "${var.project_name}-thumb-dlq"
  message_retention_seconds = 1209600
  tags                      = { Project = var.project_name }
}

resource "aws_sqs_queue" "thumb_queue" {
  name                       = "${var.project_name}-thumb-queue"
  visibility_timeout_seconds = 90
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 20

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.thumb_dlq.arn
    maxReceiveCount     = 3
  })

  tags = { Project = var.project_name }
}

# ── Notification Queue ────────────────────────────────────────
resource "aws_sqs_queue" "notification_dlq" {
  name                      = "${var.project_name}-notification-dlq"
  message_retention_seconds = 1209600
  tags                      = { Project = var.project_name }
}

resource "aws_sqs_queue" "notification_queue" {
  name                       = "${var.project_name}-notification-queue"
  visibility_timeout_seconds = 90
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 20

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.notification_dlq.arn
    maxReceiveCount     = 3
  })

  tags = { Project = var.project_name }
}

# ── Tenants DynamoDB Table ────────────────────────────────────
resource "aws_dynamodb_table" "tenants" {
  name         = "${var.project_name}-tenants"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "api_key"

  attribute {
    name = "api_key"
    type = "S"
  }

  tags = { Project = var.project_name }
}