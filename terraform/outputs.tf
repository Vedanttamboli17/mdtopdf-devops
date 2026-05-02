output "ec2_public_ip" {
  description = "SSH and browser address of your server"
  value       = aws_instance.app.public_ip
}

output "s3_bucket_name" {
  value = aws_s3_bucket.main.bucket
}

output "dynamodb_table_name" {
  value = aws_dynamodb_table.jobs.name
}

output "sqs_queue_url" {
  description = "Paste this into your .env file"
  value       = aws_sqs_queue.main.url
}